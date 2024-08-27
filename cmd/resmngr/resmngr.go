package main

// assuming each replica only use on gpu

import (
	"bufio"
	"context"
	"faas-memory/pb/gpuresc"
	"faas-memory/pb/mngrinfo"
	pb "faas-memory/pb/resmngr"
	"faas-memory/pkg/dataflow"
	"faas-memory/pkg/dataflow/schedulers"
	"faas-memory/pkg/models"
	"faas-memory/pkg/proxy"
	"faas-memory/pkg/resmngr"
	"faas-memory/pkg/resmngr/policy/placement"
	"faas-memory/pkg/resmngr/policy/scale"
	"faas-memory/pkg/util/consts"
	"faas-memory/pkg/util/filelog"
	"flag"
	"fmt"
	"math/rand"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

func init() {
	logFormat := os.Getenv("LOG_FORMAT")
	logLevel := os.Getenv("LOG_LEVEL")
	if strings.EqualFold(logFormat, "json") {
		logrus.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyMsg:  "message",
				logrus.FieldKeyTime: "@timestamp",
			},
			TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	if level, err := logrus.ParseLevel(logLevel); err == nil {
		logrus.SetLevel(level)
	}
	if level, err := zerolog.ParseLevel(logLevel); err == nil {
		zerolog.SetGlobalLevel(level)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	zerolog.TimestampFieldName = "ts"

	filelog.SetupFileLogger()

	// pretty print
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	rand.Seed(time.Now().UnixNano())
}

//CreateSched ..
func CreateSched(name string, rescStates *resmngr.ResourceManagerStates) dataflow.Scheduler {
	log.Warn().Msgf("Using scheduler: %s", name)
	var scheduler dataflow.Scheduler
	switch name {
	case "random":
		scheduler = schedulers.NewRandomScheduler(rescStates)
	case "roundrobin":
		scheduler = schedulers.NewRoundRobinScheduler(rescStates)
	case "forcelocality":
		scheduler = schedulers.NewForceLocalityScheduler(rescStates)
	case "forcemiss":
		scheduler = schedulers.NewMissLocalityScheduler(rescStates)
	case "best":
		scheduler = schedulers.NewGalvanicScheduler(rescStates)
	default:
		log.Error().Msgf("Scheduler %v not found", name)
		return nil
	}

	return scheduler
}

func setupCtrlCHandler(traceF, cpuProfF *os.File) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		trace.Stop()
		pprof.StopCPUProfile()
		if traceF != nil {
			traceF.Close()
		}
		if cpuProfF != nil {
			cpuProfF.Close()
		}
		os.Exit(0)
	}()
}

func main() {
	scalePolicyName := flag.String("scalepol", "maxcount", "scale policy, must match name in policies.yaml")
	placementPolicyName := flag.String("placepol", "loadbalance", "options: loadbalance, sameallnodes, roundrobin")
	schedulerName := flag.String("scheduler", "best", "options: random, roundrobin, forcelocality, forcemiss, best")
	gpuNodesStr := flag.String("gpuserver", "", "ip address list of gpu servers, separated by ,")
	traceFile := flag.String("trace", "", "write trace to `file`")
	cpuProfFile := flag.String("cpuprof", "", "write cpu profile to `file`")
	numWorkerPerGPU := flag.Uint("numWorkerPerGpu", 4, "number of GPU API server per GPU")

	bypass := flag.Bool("bypass", false, "whether to by pass vm management")
	flag.Parse()
	runtime.SetCPUProfileRate(100)
	runtime.SetMutexProfileFraction(1)

	var traceF, cpuProfF *os.File

	if *traceFile != "" {
		traceF, err := os.Create(*traceFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Fail to create trace file")
		}
		defer traceF.Close()
		trace.Start(traceF)
		defer trace.Stop()
	}
	if *cpuProfFile != "" {
		cpuProfF, err := os.Create(*cpuProfFile)
		if err != nil {
			log.Fatal().Err(err).Msg("fail to create cpu profile file")
		}
		defer cpuProfF.Close()
		if err = pprof.StartCPUProfile(cpuProfF); err != nil {
			log.Fatal().Err(err).Msg("fail to start CPU profile")
		}
		defer pprof.StopCPUProfile()
	}
	setupCtrlCHandler(traceF, cpuProfF)

	scalePolicyConfig := os.Getenv("SCALE_POL_CFG")
	if scalePolicyConfig == "" {
		scalePolicyConfig = "policies.yaml"
	}

	var gpuNodes []*models.GPUNode
	gpuNodeAvaMngrIps := strings.Split(*gpuNodesStr, ",")
	if *gpuNodesStr != "" && len(gpuNodeAvaMngrIps) > 0 {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		opts = append(opts, grpc.WithBlock())
		for _, avaMngrIP := range gpuNodeAvaMngrIps {
			avamngrAddr := fmt.Sprintf("%s:%d", avaMngrIP, consts.AvaMngrPort)
			log.Info().Msgf("Connecting to avamngr at %v", avamngrAddr)
			conn, err := grpc.Dial(avamngrAddr, opts...)
			if err != nil {
				log.Fatal().Err(err).Msg("fail to dial")
			}
			defer conn.Close()
			client := gpuresc.NewGPURescClient(conn)
			em := &empty.Empty{}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			rescReply, err := client.GetGPUResource(ctx, em)
			if err != nil {
				log.Fatal().Err(err).Msg("fail to get gpu resource")
			}
			gpuMem := rescReply.Memory
			log.Info().Msg("get gpu resource")
			var gpuinfos []*models.GPUInfo
			var count uint32
			count = 0
			numWorkerPerGPUUint32 := uint32(*numWorkerPerGPU)
			for uuid, mem := range gpuMem {
				firstWorkerPort := consts.AvaWorkerPortBase + count*numWorkerPerGPUUint32 + 1
				log.Warn().Str("UUID", uuid).
					Uint32("mem", mem).
					Uint32("firstWorkerPort", firstWorkerPort).Msg("")
				gpuinfos = append(gpuinfos, &models.GPUInfo{
					UUID:            uuid,
					Memory:          mem,
					FirstWorkerPort: firstWorkerPort,
				})
				count += 1
			}
			gpuNodes = append(gpuNodes, &models.GPUNode{
				GPUInfos:         gpuinfos,
				AllocatedGPUResc: make(map[string]*models.GPUResc),
				IP:               avaMngrIP,
			})
		}
	}

	lis, err := net.Listen("tcp", consts.PortToString(consts.ResMngrPort))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	s := grpc.NewServer()

	//read policies yaml and create policy passed as parameter to us
	cfg := scale.GetConfig(scalePolicyConfig)
	var scalePol scale.Interface
	log.Warn().Msgf("Using policy (%s) with cfg (%v)", *scalePolicyName, cfg[(*scalePolicyName)].(map[interface{}]interface{}))
	switch *scalePolicyName {
	case "maxcount":
		scalePol = scale.NewMaxCountPolicy(cfg[(*scalePolicyName)].(map[interface{}]interface{}))
	case "watermark":
		scalePol = scale.NewWatermarkPolicy(cfg[(*scalePolicyName)].(map[interface{}]interface{}))
	}

	log.Warn().Msgf("Using placement policy (%s) ", *placementPolicyName)
	var placePol placement.Interface
	switch *placementPolicyName {
	case "loadbalance":
		placePol = placement.NewLoadBalance()
	case "sameallnodes":
		placePol = &placement.SameAllNodes{}
	case "roundrobin":
		placePol = placement.NewRoundRobin()
	}

	resc := &models.Resources{
		GpuMapping:      make(map[string]*models.GPUChosen),
		Replicas:        make(map[string]*models.Replica),
		Functions:       make(map[string]*models.FunctionMeta),
		GPUNodes:        gpuNodes,
		Bypass:          *bypass,
		NumWorkerPerGPU: uint32(*numWorkerPerGPU),
	}

	resmngrServer := &resmngr.ResourceManagerServer{
		States: &resmngr.ResourceManagerStates{
			Resources:       resc,
			ScalePolicy:     scalePol,
			PlacementPolicy: placePol,
		},
	}
	avaManagerInfoServer := &resmngr.AvaManagerInfoServer{
		Resources: resc,
	}

	//launch resmngr
	go func() {
		log.Info().Msg("Registering grpcs")
		pb.RegisterResMngrServiceServer(s, resmngrServer)
		mngrinfo.RegisterAvAManagerInfoServer(s, avaManagerInfoServer)
		if err := s.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("failed to serve")
		}
	}()

	scheduler := CreateSched(*schedulerName, resmngrServer.States)
	go func() {
		log.Info().Msg("Launching backend")
		proxy.SetScheduler(scheduler)
		log.Info().Msg("Launching pprof server")
		proxy.RegisterPprof()
		proxy.LaunchBackend(resmngrServer.States)
	}()

	log.Info().Msg("Backend and resmngr launched, looping")

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		//change scheduler
		if strings.HasPrefix(text, "cs") {
			name := strings.Fields(text)[1]
			log.Warn().Msgf("Changing scheduler to %v", name)
			scheduler := CreateSched(name, resmngrServer.States)

			if scheduler != nil {
				proxy.ChangeScheduler(scheduler)
			}
		} else if strings.HasPrefix(text, "starthogs") {
			var cores int32
			var _pref string
			_, err := fmt.Sscanf(text, "%s %d", &_pref, &cores)
			if err != nil {
				log.Warn().Msgf("Cant parse cpuhog string: %v", err)
				continue
			}

			for _, node := range resc.Nodes {
				node.CreateCPUHogs(cores)
			}
		} else if strings.HasPrefix(text, "stophogs") {
			var cores int32
			var _pref string
			_, err := fmt.Sscanf(text, "%s %d", &_pref, &cores)
			if err != nil {
				log.Warn().Msgf("Cant parse cpuhog string: %v", err)
				continue
			}

			for _, node := range resc.Nodes {
				node.StopCPUHogs(cores)
			}
		}
	}
}
