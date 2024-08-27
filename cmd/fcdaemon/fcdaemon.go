package main

import (
	"bytes"
	"context"
	pb "faas-memory/pb/fcdaemon"
	resmngrpb "faas-memory/pb/resmngr"
	"faas-memory/pkg/fc"
	"faas-memory/pkg/util/consts"
	llog "faas-memory/pkg/util/log"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type fcServiceServer struct {
	pb.UnimplementedFcServiceServer
	sync.RWMutex
	fcTool         *fc.Tool
	hogs           []*os.Process
	newMachineLock sync.Mutex
	ctx            context.Context
	vmSetupPool    *fc.VMSetupPool
	state          fc.ServiceState

	fnCountLock sync.Mutex
	fnCount     map[string]uint32

	numaInfo *NumaInfo
}

//NumaInfo stores information of numa node
type NumaInfo struct {
	nodes  []*NumaNode
	numCPU uint32
}

//NumaNode stores the range of cpus in this numa node
//We assume the cpu in the range doesn't go offline
type NumaNode struct {
	start uint32
	end   uint32
}

var vmCount uint32

func (s *fcServiceServer) CreateReplica(ctx context.Context, req *pb.CreateRequest) (*pb.CreateReply, error) {
	log.Debug().Str("fnName", req.ImageName).Msg("Launching VM...")

	names := strings.Split(req.ImageName, "/")
	name := names[len(names)-2]

	s.fnCountLock.Lock()
	count, exists := s.fnCount[name]
	if !exists {
		s.fnCount[name] = 0
	}
	s.fnCount[name]++
	count = s.fnCount[name]
	s.fnCountLock.Unlock()

	cpuID := (count - 1) % s.numaInfo.numCPU
	numaNode := uint32(0)
	for idx, numa := range s.numaInfo.nodes {
		if cpuID >= numa.start && cpuID <= numa.end {
			numaNode = uint32(idx)
			break
		}
	}
	replica, err := fc.CreateReplica(s.ctx, &s.newMachineLock,
		s.fcTool.FirecrackerBinary, s.vmSetupPool, &s.state, req, cpuID, numaNode)

	if err != nil {
		log.Fatal().Err(err).Msg("err on CreateReplica")
	}

	atomic.AddUint32(&vmCount, 1)
	vmid := replica.Vmid
	ip := replica.MachineConfig.
		NetworkInterfaces[0].
		StaticConfiguration.
		IPConfiguration.IPAddr.IP.String()

	log.Warn().
		Str("ip", ip).
		Str("vmid", vmid).
		Uint32("vmCount", atomic.LoadUint32(&vmCount)).
		Uint32("cpu id", cpuID).
		Uint32("numa node", numaNode).
		Str("fnName", name).
		Msg("VM launched and ready")

	if err == nil {
		return &pb.CreateReply{
			Ip:   ip,
			Vmid: vmid,
		}, nil
	}
	return nil, err
}

func (s *fcServiceServer) ReleaseVMSetup(ctx context.Context, req *pb.ReleaseRequest) (*empty.Empty, error) {
	vmid := req.Vmid
	err := s.vmSetupPool.ReturnVMSetupByVmid(vmid)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

//CreateCPUHogs ..
func (s *fcServiceServer) CreateCPUHogs(ctx context.Context, req *pb.CpuHogs) (*empty.Empty, error) {
	var args []string = make([]string, 1)
	args[0] = fmt.Sprintf("%d", req.N)

	log.Debug().Msgf("Starting %v hogs", req.N)
	for i := 0; int32(i) < req.N; i++ {

		cmd := exec.Command("./cpuhog", "1")
		if err := cmd.Start(); err != nil {
			log.Fatal().Msgf("error on Start: %v", err)
		}

		s.hogs = append(s.hogs, cmd.Process)
	}

	return &empty.Empty{}, nil
}

//StopCPUHogs ..
func (s *fcServiceServer) StopCPUHogs(ctx context.Context, req *pb.CpuHogs) (*empty.Empty, error) {
	log.Debug().Msgf("Killing up to %v hogs", req.N)
	for i := 0; int32(i) < req.N; i++ {
		if len(s.hogs) == 0 {
			break
		}

		s.hogs[0].Kill()
		s.hogs = s.hogs[1:]
	}

	return &empty.Empty{}, nil
}

// getOutboundIP get preferred outbound ip of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal().Err(err).Msg("dial googld dns")
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	if level, err := zerolog.ParseLevel(logLevel); err == nil {
		zerolog.SetGlobalLevel(level)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
	//pretty print
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if llevel, err := llog.ParseLevel(logLevel); err == nil {
		logrus.SetLevel(llevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}
}

func startProxy() {
	//references:
	//https://golang.org/pkg/net/http/#RoundTripper
	//https://www.integralist.co.uk/posts/golang-reverse-proxy/
	director := func(req *http.Request) {
		//log.Printf("headers  %v", req.Header)
		//req.Header.Add("X-Forwarded-Host", req.Host)
		req.URL.Scheme = "http"
		req.URL.Host = req.Header.Get("Redirect-To") + ":8080"
		log.Info().
			Str("from", req.Host).
			Str("Redirect-To", req.URL.Host).
			Msgf("reverse proxy redirecting request to function")
	}

	dialer := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   1500 * time.Second,
			KeepAlive: -1,
		}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          100,
		IdleConnTimeout:       0,
		TLSHandshakeTimeout:   0,
		ExpectContinueTimeout: 0,
	}

	proxy := &httputil.ReverseProxy{Director: director, Transport: dialer}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
	go http.ListenAndServe(consts.PortToString(consts.FcDaemonProxyPort), nil)
}

func getNuma() (*NumaInfo, error) {
	cmd := exec.Command("numactl", "-H")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	var numaInfo NumaInfo
	outStr := out.String()
	splts := strings.Split(outStr, "\n")
	numCPU := uint32(0)
	for _, splt := range splts {
		if strings.Contains(splt, "available") {
			fields := strings.Fields(splt)
			numNuma, err := strconv.ParseUint(fields[1], 10, 32)
			if err != nil {
				return nil, err
			}
			fmt.Printf("num numa node: %v\n", numNuma)
			numaInfo.nodes = make([]*NumaNode, 0, numNuma)
		}
		if strings.Contains(splt, "cpus") {
			fields := strings.Fields(splt)
			start, err := strconv.ParseUint(fields[3], 10, 32)
			if err != nil {
				return nil, err
			}
			length := len(fields)
			end, err := strconv.ParseUint(fields[length-1], 10, 32)
			if err != nil {
				return nil, err
			}
			numaInfo.nodes = append(numaInfo.nodes, &NumaNode{
				start: uint32(start),
				end:   uint32(end),
			})
			numCPU += uint32(length - 3)
		}
	}
	numaInfo.numCPU = numCPU
	return &numaInfo, nil
}

func main() {
	vmCount = 0
	netname := flag.String("netname", "fcnet", "network name of cni")
	resmngrIP := flag.String("resmngraddr", "", "address of resource manager")
	sizeIPPool := flag.Int("ippool", 120, "how many ips to pre allocate")
	listenPort := flag.Int("port", consts.FcDaemonPort, "port to listen")
	cachePort := flag.Uint("cacheport", 6379, "port of the cache each vm will talk to")
	useCustomCache := flag.Bool("useCustomCache", false, "whether to use the custom cache server; default is redis")
	avaLog := flag.String("avaLog", "info", "ava log level")
	flag.Parse()

	numaInfo, err := getNuma()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to find NUMA info")
	}
	log.Info().Bool("useCustomCache", *useCustomCache).Int("vm pool size", *sizeIPPool).Msg("")
	resmngrIPStr := *resmngrIP
	if resmngrIPStr == "" {
		resmngrIPStr = os.Getenv("RESMNGRADDR")
		if resmngrIPStr == "" {
			log.Fatal().Msg("required flag not set: -resmngraddr=<ip>")
		}
	}

	startProxy()

	// get current ip and register node
	outboundIP := getOutboundIP()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	// setup listener
	lis, err := net.Listen("tcp", consts.PortToString(*listenPort))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	fcTool, err := fc.NewTool()
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create firecracker tool")
	}

	s := grpc.NewServer()

	server := &fcServiceServer{
		fcTool: fcTool,
		ctx:    context.Background(),
		state: fc.ServiceState{
			NetName:        *netname,
			UseCustomCache: *useCustomCache,
			CachePort:      *cachePort,
			AvALogLevel:    *avaLog,
		},
		fnCount:  make(map[string]uint32),
		numaInfo: numaInfo,
	}
	vmSetupPool := fc.NewVMSetupPool(server.ctx, server.state.NetName)
	err = vmSetupPool.FillVMSetupPool(*sizeIPPool)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create vm setup pool")
	}
	server.vmSetupPool = vmSetupPool

	go func() {
		resmngrAddr := fmt.Sprintf("%s:%d", resmngrIPStr, consts.ResMngrPort)
		log.Warn().Msgf("Connecting to resmngr at %v", resmngrAddr)
		server.state.ResmngrAddr = resmngrAddr
		server.state.FcdaemonAddr = fmt.Sprintf("%s:%d", outboundIP.String(), consts.FcDaemonPort)
		log.Warn().Msgf("Outbound IP is %s", server.state.FcdaemonAddr)
		conn, err := grpc.Dial(resmngrAddr, opts...)
		if err != nil {
			log.Fatal().Err(err).Msg("fail to dial")
		}
		defer conn.Close()
		client := resmngrpb.NewResMngrServiceClient(conn)
		uuid, _ := uuid.NewV4()
		req := &resmngrpb.RegisterNodeRequest{
			Ip:     outboundIP.String(),
			Port:   uint32(*listenPort),
			NodeId: uuid.String(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err = client.RegisterNode(ctx, req)
		if err != nil {
			log.Fatal().Err(err).Msg("fail to register node")
		}
		log.Warn().Msg("Registered at resmngr")
	}()

	pb.RegisterFcServiceServer(s, server)
	if err := s.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}
