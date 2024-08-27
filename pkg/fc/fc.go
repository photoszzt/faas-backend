package fc

import (
	"context"
	// "crypto/rand"
	pb "faas-memory/pb/fcdaemon"
	"faas-memory/pkg/util/log"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	zlog "github.com/rs/zerolog/log"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"path"
	"strconv"
	"strings"
)

const (
	defaultCPU = 1
	defaultMem = 512
	// executableMask is the mask needed to check whether or not a file's
	// permissions are executable.
	executableMask         = 0111
	firecrackerDefaultPath = "firecracker"
	defaultNamespace       = "default"
	// DefaultImageSuffix is the default image tag kubelet try to add to image.
	DefaultImageSuffix = ":latest"
)

// Tool is a struct that contains all the meta data of the functions
type Tool struct {
	FirecrackerBinary string
	Ctx               context.Context
	InetDev           string
}

// TryAgainError contains the error message that asks the client to retry
type TryAgainError struct {
	Msg string // description of error
}

// ServiceState contains states that are necessary to start the VM
type ServiceState struct {
	ResmngrAddr    string
	FcdaemonAddr   string
	NetName        string
	CachePort      uint
	UseCustomCache bool
	AvALogLevel    string
}

func (e *TryAgainError) Error() string { return e.Msg }

type namespaceKey struct{}

// ReplicaMeta contains metadata of a serverless vm replica
// >>>>  This is used by fcdaemon, not stored at the gateway/proxy
type ReplicaMeta struct {
	Machine       *firecracker.Machine
	MachineConfig firecracker.Config
	ctx           context.Context
	ctxCancel     context.CancelFunc
	Vmid          string
	vmDir         Dir
	logger        *logrus.Entry
	osvLogFile    *os.File
}

// newReplicaMeta creates a new ReplicaMeta
func newReplicaMeta(ctx context.Context, vmID string) (*ReplicaMeta, error) {
	vmmCtx, vmmCancel := context.WithCancel(ctx)
	namespace := defaultNamespace
	vmmCtx = context.WithValue(vmmCtx, namespaceKey{}, namespace)
	logger := log.G(vmmCtx)
	logger = logger.WithField("vmid", vmID)
	logger = logger.WithField("namespace", namespace)

	vmDir, err := VMDir(namespace, vmID)
	if err != nil {
		vmmCancel()
		return nil, err
	}
	if _, err := os.Stat(vmDir.RootPath()); os.IsNotExist(err) {
		if err := vmDir.Mkdir(); err != nil {
			vmmCancel()
			return nil, err
		}
	}
	return &ReplicaMeta{
		ctx:       vmmCtx,
		ctxCancel: vmmCancel,
		Vmid:      vmID,
		logger:    logger,
		vmDir:     vmDir,
	}, nil
}

// NewTool creates Tool
func NewTool() (*Tool, error) {
	firecrackerBinary, err := exec.LookPath(firecrackerDefaultPath)
	if err != nil {
		return nil, err
	}
	finfo, err := os.Stat(firecrackerBinary)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("binary %q does not exist: %v", firecrackerBinary, err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to stat binary, %q: %v", firecrackerBinary, err)
	}

	if finfo.IsDir() {
		return nil, fmt.Errorf("binary, %q, is a directory", firecrackerBinary)
	} else if finfo.Mode()&executableMask == 0 {
		return nil, fmt.Errorf("binary, %q, is not executable. Check permissions of binary", firecrackerBinary)
	}
	return &Tool{
		FirecrackerBinary: firecrackerBinary, Ctx: context.Background()}, nil
}

//not used due to CNI
// AssignTapIP assign tap ip for osv vm
/*
func (fc *Tool) AssignTapIP(baseIP net.IP) (string, net.IP, net.IP, error) {
	fc.CLock.Lock()
	if fc.counter >= 255 {
		fc.CLock.Unlock()
		return "", nil, nil, fmt.Errorf("we have exausted all the ip addresses for this machine")
	}
	tapIP := net.IPv4(baseIP[12], baseIP[13], baseIP[14], fc.counter+1)
	clientIP := net.IPv4(baseIP[12], baseIP[13], baseIP[14], fc.counter+2)
	tap := fmt.Sprintf("fc_tap%d", fc.counter/4+1)
	fc.counter += 4
	fc.CLock.Unlock()
	return tap, tapIP, clientIP, nil
}
*/

// PrepareReplicaRequest prepares the argument needed for creating replica
func PrepareReplicaRequest(labels map[string]string,
	envs map[string]string,
	image string) (*pb.CreateRequest, error) {

	kernelImagePath, err := getKernelImagePath(labels)
	if err != nil {
		return nil, fmt.Errorf("fail to get kernel image path: %v", err)
	}
	initrdPath := getInitrdPath(labels)
	vmType, exists := labels["VM_TYPE"]
	if !exists {
		vmType = "osv"
	}
	//add space to guarantee it will work if yaml didn't have one at the end (.. of course this happened)
	kernelArgs := " " + getKernelArgs(labels, envs, vmType) + " "
	imageName := getActualImageName(image)
	imageStoragePrefix, exists := labels["IMAGE_STORAGE"]
	if !exists {
		imageStoragePrefix = "/var/run/osv-storage"
	}
	inetDev, exists := labels["INET_DEV"]
	if !exists {
		inetDev = ""
	}
	nameServer := getNameServer(labels)
	cpu := int64(getCPU(labels))
	mem := int64(getMem(labels))
	awsAccessKeyID, awsSecretAccessKey, err := getAwsKeys(envs)
	if err != nil {
		return nil, err
	}
	return &pb.CreateRequest{
		KernelImagePath:    kernelImagePath,
		KernelArgs:         kernelArgs,
		InitrdPath:         initrdPath,
		ImageName:          imageName,
		ImageStoragePrefix: imageStoragePrefix,
		InetDev:            inetDev,
		NameServer:         nameServer,
		AwsAccessKeyID:     awsAccessKeyID,
		AwsSecretAccessKey: awsSecretAccessKey,
		Cpu:                cpu,
		Mem:                mem,
		VmType:             vmType,
	}, nil
}

// CreateReplica creates a new serverless function replica
func CreateReplica(ctx context.Context, newMachineLock *sync.Mutex,
	fcBinary string, vmSetupPool *VMSetupPool, serviceState *ServiceState,
	req *pb.CreateRequest, cpuID uint32, numaNode uint32) (*ReplicaMeta, error) {

	newMachineLock.Lock()

	vmSetup, err := vmSetupPool.GetVMSetup()
	if err != nil {
		return nil, err
	}
	vmidStr := vmSetup.ReplicaMeta.Vmid
	replicaMeta := vmSetup.ReplicaMeta

	root := false
	if req.VmType == "linux" {
		root = true
	}

	socketPath := replicaMeta.vmDir.FirecrackerSockPath()
	var guestDrive models.Drive
	if req.InitrdPath != "" {
		replicateImagePath, err := linkReplicateImage(req.ImageName, vmidStr, req.ImageStoragePrefix)
		if err != nil {
			return nil, fmt.Errorf("fail to link replicate image path %v: %v", replicateImagePath, err)
		}
		guestDrive = models.Drive{
			DriveID:      firecracker.String("rootfs"),
			IsReadOnly:   firecracker.Bool(true),
			IsRootDevice: firecracker.Bool(false),
			PathOnHost:   &replicateImagePath,
		}
	} else {
		replicateImagePath, err := replicateImage(req.ImageName, vmidStr, req.ImageStoragePrefix)
		if err != nil {
			return nil, fmt.Errorf("fail to get replicate image path %v: %v", replicateImagePath, err)
		}
		guestDrive = models.Drive{
			DriveID:      firecracker.String("rootfs"),
			IsReadOnly:   firecracker.Bool(false),
			IsRootDevice: firecracker.Bool(root),
			PathOnHost:   &replicateImagePath,
		}
	}

	fnName := filepath.Base(filepath.Dir(req.ImageName))

	// create osv log
	osvLogFilePath := replicaMeta.vmDir.OsvOutputLogfilePath()
	osvLogFile, err := os.OpenFile(osvLogFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		return nil, fmt.Errorf("fail to open osv log file: %v", err)
	}
	replicaMeta.osvLogFile = osvLogFile

	newMachineLock.Unlock()

	redisAddr := strings.Split(serviceState.FcdaemonAddr, ":")[0]
	redisAddr = fmt.Sprintf("%s:%d", redisAddr, serviceState.CachePort)
	kernelArgs := fmt.Sprintf(" --env=RES_MNGR_ADDR=%s --env=FCD_ADDR=%s  --env=REDIS_CACHE_ADDR=%s --env=OS=%s --env=VMID=%s --env=USE_CUSTOM_CACHE=%t --env=AVA_LOG_LEVEL=%s ",
		serviceState.ResmngrAddr, serviceState.FcdaemonAddr, redisAddr, req.VmType, vmidStr, serviceState.UseCustomCache, serviceState.AvALogLevel) + req.KernelArgs

	//fmt.Printf("args: %s\n\n", kernelArgs)

	config := firecracker.Config{
		SocketPath:      socketPath,
		KernelArgs:      kernelArgs,
		KernelImagePath: req.KernelImagePath,
		InitrdPath:      req.InitrdPath,
		LogFifo:         replicaMeta.vmDir.FirecrackerLogFifoPath(),
		MetricsFifo:     replicaMeta.vmDir.FirecrackerMetricsFifoPath(),
		LogLevel:        "Error",
		Debug:           false,
		NetworkInterfaces: firecracker.NetworkInterfaces{
			firecracker.NetworkInterface{
				StaticConfiguration: vmSetup.networkInterfaces[0].StaticConfiguration,
			},
		},
		MachineCfg: models.MachineConfiguration{
			CPUTemplate: "T2",
			HtEnabled:   firecracker.Bool(false),
			MemSizeMib:  firecracker.Int64(req.Mem),
			VcpuCount:   firecracker.Int64(req.Cpu),
		},
		Drives:            []models.Drive{guestDrive},
		DisableValidation: true,
		VMID:              vmidStr,
		VMType:            req.VmType,
		SeccompLevel:      firecracker.SeccompLevelDisable,
		NetNS:             vmSetup.netns,
	}

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(replicaMeta.logger),
	}
	cmd := firecracker.VMCommandBuilder{}.
		WithBin(fcBinary).
		WithSocketPath(socketPath).
		// WithStderr(os.Stdout).
		// WithStdout(os.Stdout).
		WithStderr(osvLogFile).
		WithStdout(osvLogFile).
		WithArgs([]string{"--seccomp-level", "0"}).
		WithNuma(int(numaNode)).
		WithCPUId(int(cpuID)).
		BuildWithNuma(replicaMeta.ctx)
	cmd.Env = append(cmd.Env, "AWS_ACCESS_KEY_ID="+req.AwsAccessKeyID,
		"AWS_SECRET_ACCESS_KEY="+req.AwsSecretAccessKey,
		"VMID="+vmidStr, "RES_MNGR_ADDR="+serviceState.ResmngrAddr, "FN_NAME="+fnName)
	machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))

	m, err := firecracker.NewMachine(replicaMeta.ctx, config, machineOpts...)

	if err != nil {
		return nil, errors.Wrapf(err, "Failed creating machine")
	}
	replicaMeta.Machine = m
	replicaMeta.MachineConfig = config

	//tried to make launch async, so we send a response to resmngr as soon as possible
	//and avoid a race condition of updateStatus and add us to their map, didnt work
	stime := time.Now()
	if err := m.Start(replicaMeta.ctx); err != nil {
		zlog.Fatal().Msgf("fail to start the machine: %v", err)
	}
	duration := time.Since(stime)
	zlog.Debug().Dur("boot time for one vm", duration).Str("fn", fnName).Msg("startup time measured in fc")

	return replicaMeta, nil
}

func replicateImage(image, vmid, imageStoragePrefix string) (string, error) {
	replicatedImageDir := path.Join(imageStoragePrefix, vmid)
	imageName := path.Base(image)
	err := os.MkdirAll(replicatedImageDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("mkdir: %v", err)
	}
	//fmt.Printf("image dir path: %v\n", replicatedImageDir)
	replicatedImagePath := path.Join(replicatedImageDir, imageName)
	cmd := exec.Command("cp", image, replicatedImagePath)
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("cp: %v", err)
	}
	return replicatedImagePath, nil
}

func linkReplicateImage(image, vmid, imageStoragePrefix string) (string, error) {
	replicatedImageDir := path.Join(imageStoragePrefix, vmid)
	imageName := path.Base(image)
	err := os.MkdirAll(replicatedImageDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("mkdir: %v", err)
	}
	//fmt.Printf("image dir path: %v\n", replicatedImageDir)
	replicatedImagePath := path.Join(replicatedImageDir, imageName)
	cmd := exec.Command("ln", "-s", image, replicatedImagePath)
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("cp: %v", err)
	}
	return replicatedImagePath, nil
}

// func (fc *Tool) shutdownAll() {
// }

// func installSignalHandlers(ctx context.Context, fc *Tool) {
// 	go func() {
// 		// Clear some default handlers installed by the firecracker SDK:
// 		signal.Reset(os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
// 		c := make(chan os.Signal, 1)
// 		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
//
// 		for {
// 			switch s := <-c; {
// 			case s == syscall.SIGTERM || s == os.Interrupt:
// 				logrus.Printf("Caught SIGINT, requesting clean shutdown")
// 				err := m.Shutdown(ctx)
// 				if err != nil {
// 					logrus.Printf("Fail clean shutdown, forcing shutdown")
// 					m.StopVMM()
// 				}
// 				deleteTap(tap)
// 				return
// 			case s == syscall.SIGQUIT:
// 				logrus.Printf("Caught SIGTERM, forcing shutdown")
// 				m.StopVMM()
// 				deleteTap(tap)
// 				return
// 			}
// 		}
// 	}()
// }

/*
func setupIPTable(inetDev, tap string) error {
	if inetDev != "" {
		cmd := exec.Command("sudo", "sh", "-c",
			"echo 1 > /proc/sys/net/ipv4/ip_forward")
		err := cmd.Run()
		if err != nil {
			return err
		}
		cmd = exec.Command("sudo", "iptables", "-t", "nat", "-A", "POSTROUTING", "-o", inetDev, "-j",
			"MASQUERADE")
		err = cmd.Run()
		if err != nil {
			return err
		}
		cmd = exec.Command("sudo", "iptables", "-A", "FORWARD", "-m", "conntrack", "--ctstate",
			"RELATED,ESTABLISHED", "-j", "ACCEPT")
		err = cmd.Run()
		if err != nil {
			return err
		}
		cmd = exec.Command("sudo", "iptables", "-A", "FORWARD", "-i", tap, "-o", inetDev, "-j", "ACCEPT")
		err = cmd.Run()
		if err != nil {
			return err
		}
		cmd = exec.Command("sudo", "iptables", "-A", "INPUT", "-i", tap, "-p", "udp", "-m", "udp", "-m",
			"multiport", "--dports", "53", "-j", "ACCEPT")
		err = cmd.Run()
		if err != nil {
			return err
		}
		cmd = exec.Command("sudo", "iptables", "-A", "INPUT", "-i", tap, "-p", "tcp", "-m", "tcp", "-m",
			"multiport", "--dports", "53", "-j", "ACCEPT")
		err = cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("need to provide internet facing device through env")
}
*/

/*
// DeleteIPTable deletes the ip tables rules associated with tap and inetDev
func DeleteIPTable(tap string, inetDev string) {
	cmd := exec.Command("sudo", "iptables", "-t", "nat", "-D", "POSTROUTING", "-o", inetDev, "-j", "MASQUERADE")
	_ = cmd.Run()
	cmd = exec.Command("sudo", "iptables", "-D", "FORWARD", "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	_ = cmd.Run()
	cmd = exec.Command("sudo", "iptables", "-D", "FORWARD", "-i", tap, "-o", inetDev, "-j", "ACCEPT")
	_ = cmd.Run()
	cmd = exec.Command("sudo", "iptables", "-D", "INPUT", "-i", tap, "-p", "udp", "-m", "udp", "-m", "multiport", "--dports", "53", "-j", "ACCEPT")
	_ = cmd.Run()
	cmd = exec.Command("sudo", "iptables", "-D", "INPUT", "-i", tap, "-p", "tcp", "-m", "tcp", "-m", "multiport", "--dports", "53", "-j", "ACCEPT")
	_ = cmd.Run()
}
*/

/*
func createTap(name string, ip string) {
	cmd := exec.Command("sudo", "ip", "tuntap", "add", "dev", name, "mode", "tap")
	err := cmd.Run()
	if err != nil {
		logrus.Fatalf("Fail to create tap device %s: %v", name, err)
	}
	cmd = exec.Command("sudo", "sysctl", "-q", "-w",
		fmt.Sprintf("net.ipv4.conf.%s.proxy_arp=1", name))
	err = cmd.Run()
	if err != nil {
		logrus.Fatalf("Fail to set ipv4: %v", err)
	}
	cmd = exec.Command("sudo", "sysctl", "-q", "-w",
		fmt.Sprintf("net.ipv6.conf.%s.disable_ipv6=1", name))
	err = cmd.Run()
	if err != nil {
		logrus.Fatalf("Fail to disable ipv6: %v", err)
	}
	cmd = exec.Command("sudo", "ip", "addr", "add", fmt.Sprintf("%s/30", ip), "dev", name)
	err = cmd.Run()
	if err != nil {
		logrus.Fatalf("Fail to set address to tap device: %v", err)
	}
	cmd = exec.Command("sudo", "ip", "link", "set", "dev", name, "up")
	err = cmd.Run()
	if err != nil {
		logrus.Fatalf("Fail to up tap device: %v", err)
	}
}
*/

// DeleteTap deletes the tap device with given name
func DeleteTap(name string) {
	cmd := exec.Command("sudo", "ip", "link", "set", name, "down")
	err := cmd.Run()
	if err != nil {
		logrus.Warnf("Fail to down tap device %s: %v", name, err)
	}
	cmd = exec.Command("sudo", "ip", "link", "delete", name)
	err = cmd.Run()
	if err != nil {
		logrus.Warnf("Fail to delete tap device %s: %v", name, err)
	}
}

/*
func generateMacAddress() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		logrus.Fatalf("Fail to generate mac address: %v", err)
	}
	buf[0] = (buf[0] | 2) & 0xfe // set local bit, ensure unicast address
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
}
*/

// GetMinReplicaCount gets the minimum replica count
func GetMinReplicaCount(labels map[string]string) int {
	if value, exists := labels["com.openfaas.scale.min"]; exists {
		minReplicas, err := strconv.Atoi(value)
		if err == nil && minReplicas >= 0 {
			return minReplicas
		}
		logrus.Println(err)
	}
	return 0
}

// GetMaxReplicaCount gets the maximum replica count
func GetMaxReplicaCount(labels map[string]string) int {
	if value, exists := labels["com.openfaas.scale.max"]; exists {
		maxReplicas, err := strconv.Atoi(value)
		if err == nil && maxReplicas > 0 {
			return maxReplicas
		}
		logrus.Println(err)
	}
	return 20
}

// GetExecEnvironmentCount gets the EE count
func GetExecEnvironmentCount(labels map[string]string) int {
	if value, exists := labels["execenvironment.count"]; exists {
		eecount, err := strconv.Atoi(value)
		if err == nil && eecount > 0 {
			return eecount
		}
		logrus.Println(err)
	}
	return 5
}

func getKernelImagePath(labels map[string]string) (string, error) {
	if kernelPath, exists := labels["KERNEL_IMAGE_PATH"]; exists {
		return kernelPath, nil
	}
	return "", fmt.Errorf("need to provide kernel image path through env")
}

func getInitrdPath(labels map[string]string) string {
	if initrdPath, exists := labels["INITRD_PATH"]; exists {
		return initrdPath
	} else {
		return ""
	}
}

func getNameServer(labels map[string]string) string {
	if nameserver, exists := labels["NAMESERVER"]; exists {
		return nameserver
	}
	return "8.8.8.8"
}

func getKernelArgs(labels map[string]string, envs map[string]string, vmType string) string {
	if kernelArgs, exists := labels["KERNEL_ARGS"]; exists {
		//fmt.Printf("getKernelArgs %v", kernelArgs)
		var b strings.Builder
		if vmType == "linux" {
			b.WriteString(kernelArgs)
		}
		for k, v := range envs {
			b.WriteString("--env=")
			b.WriteString(k)
			b.WriteString("=")
			b.WriteString(v)
			b.WriteString(" ")
		}
		if vmType == "osv" {
			b.WriteString(kernelArgs)
		}
		return b.String()
	}
	if vmType == "osv" {
		return "--no-pci"
	}
	return ""
}

func getAwsKeys(envs map[string]string) (string, string, error) {
	if awsAccessKeyID, exists := envs["AWS_ACCESS_KEY_ID"]; exists {
		if awsSecretAccessKey, exists := envs["AWS_SECRET_ACCESS_KEY"]; exists {
			return awsAccessKeyID, awsSecretAccessKey, nil
		}
		return "", "", fmt.Errorf("need to provide both AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY. Only AWS_ACCESS_KEY_ID is provided")
	}
	return "", "", fmt.Errorf("need to provide both AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY")
}

func getCPU(labels map[string]string) uint32 {
	if value, exists := labels["com.openfaas.cpu"]; exists {
		cpu, err := strconv.Atoi(value)
		if err == nil && cpu > 0 {
			return uint32(cpu)
		}
		logrus.Println(err)
	}
	return defaultCPU
}

func getMem(labels map[string]string) uint32 {
	if value, exists := labels["com.openfaas.memory"]; exists {
		memory, err := strconv.Atoi(value)
		if err == nil && memory > 0 {
			return uint32(memory)
		}
		logrus.Println(err)
	}
	return defaultMem
}

// getActualImageName standardize the image name by remove the default image tag if presents
func getActualImageName(imageName string) string {
	if strings.HasSuffix(imageName, DefaultImageSuffix) {
		return imageName[0 : len(imageName)-len(DefaultImageSuffix)]
	}
	return imageName
}
