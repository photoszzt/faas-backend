package resmngr

import (
	"context"
	"encoding/json"
	"errors"
	fpb "faas-memory/pb/fcdaemon"
	"faas-memory/pb/mngrinfo"
	pb "faas-memory/pb/resmngr"
	"faas-memory/pkg/fc"
	"faas-memory/pkg/models"
	"faas-memory/pkg/resmngr/policy/placement"
	"faas-memory/pkg/resmngr/policy/scale"
	"faas-memory/pkg/util/consts"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/openfaas/faas-provider/types"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var (
	ErrRetry = errors.New("no VMs available, retry")
)

// ResourceManagerServer  has the global state of the manager
type ResourceManagerServer struct {
	pb.UnimplementedResMngrServiceServer
	States *ResourceManagerStates
}

// ResourceManagerStates contain global state
type ResourceManagerStates struct {
	Resources       *models.Resources
	ScalePolicy     scale.Interface
	PlacementPolicy placement.Interface
}

//AvaManagerInfoServer ..
type AvaManagerInfoServer struct {
	mngrinfo.UnimplementedAvAManagerInfoServer
	Resources *models.Resources
}

//DeployRequest request to deploy a VM
type DeployRequest struct {
	RequestJSON []byte
}

// DeployReply reply of deploy requests
type DeployReply struct {
	Ok uint32
}

//ResourceRequest request for a VM
type ResourceRequest struct {
	Function string
	Affinity string
	Amount   uint32
}

//GetAvAManagerInfo ..
func (s *AvaManagerInfoServer) GetAvAManagerInfo(ctx context.Context, req *mngrinfo.AvAManagerInfoRequest) (*mngrinfo.AvAManagerInfoReply, error) {
	s.Resources.GPUMappingLock.RLock()
	defer s.Resources.GPUMappingLock.RUnlock()
	reply, exists := s.Resources.GpuMapping[req.Vmid]
	if exists {
		return reply.AvAReply, nil
	}
	return nil, fmt.Errorf("fail to find gpu mapping with %v", req.Vmid)
}

//DeployFunction ..
func (states *ResourceManagerStates) DeployFunction(ctx context.Context, req *DeployRequest) (*DeployReply, error) {
	request := types.FunctionDeployment{}
	if err := json.Unmarshal(req.RequestJSON, &request); err != nil {
		log.Error().Err(err).Msg("error during unmarshal of create function request.")
		return nil, err
	}

	if _, exists := states.Resources.Functions[request.Service]; exists {
		//return nil, fmt.Errorf("function (%v) is already deployed", request.Service)
		log.Warn().Str("function", request.Service).Msg("Function already deployed, overwriting..")
	}

	minReplicas := fc.GetMinReplicaCount(*request.Labels)
	maxReplicas := fc.GetMaxReplicaCount(*request.Labels)
	//eeCount := fc.GetExecEnvironmentCount(*request.Labels)

	log.Printf("Deploying function (%v)", request.Service)

	gpuMemReq, exists := (*request.Labels)["GPU_MEM"]
	gpuMemReqMegU32 := uint32(0)
	if exists {
		gpuMemReqMeg, err := strconv.ParseUint(gpuMemReq, 10, 32)
		if err != nil {
			log.Fatal().
				Err(err).
				Str("GPU Mem Input", gpuMemReq).
				Msg("fail to parse gpu memory input")
		}
		gpuMemReqMegU32 = uint32(gpuMemReqMeg)
	}
	//the mutex for the cond
	//lock := sync.Mutex{}
	fn := &models.FunctionMeta{
		Name:     request.Service,
		Replicas: make(map[string]*models.Replica),
		//WaitingForVM:              sync.NewCond(&lock),
		ExecEnvironmentPerReplica: 1, //force one only
		//fields for fcdaemon grpc
		Image:       request.Image,
		EnvProcess:  request.EnvProcess,
		Labels:      request.Labels,
		Annotations: request.Annotations,
		EnvVars:     request.EnvVars,
		GPUMemReq:   gpuMemReqMegU32,
	}

	states.Resources.FunctionsLock.Lock()
	states.Resources.Functions[request.Service] = fn
	states.Resources.FunctionsLock.Unlock()

	//tell the policy a function was deployed
	reps := states.ScalePolicy.OnFunctionDeploy(minReplicas, maxReplicas)
	go states.scaleUpFunction(fn, reps)

	return &DeployReply{Ok: 1}, nil
}

//UpdateStatus ..
func (s *ResourceManagerServer) UpdateStatus(ctx context.Context, req *pb.StatusUpdateRequest) (*pb.StatusUpdateReply, error) {
	vmid := req.Vmid
	var rep *models.Replica = nil
	//we gotta keep trying since we are waiting for fcdaemon to send us the grpc response, then goes through 2 channels
	for i := 0; i < 200; i++ {
		s.States.Resources.ReplicasLock.RLock()
		_, exists := s.States.Resources.Replicas[vmid]
		s.States.Resources.ReplicasLock.RUnlock()
		if !exists {
			//log.Printf("can't update vmid (%v) because it's not in our map, retrying %v", vmid, i)
			//another ugly hack
			time.Sleep(10 * time.Millisecond)
		} else {
			s.States.Resources.ReplicasLock.RLock()
			rep = s.States.Resources.Replicas[vmid]
			s.States.Resources.ReplicasLock.RUnlock()
			break
		}
	}

	if rep == nil {
		log.Fatal().
			Str("vmid", vmid).
			Msg("cant find replica on our map after 20 retries")
	}

	//When we get a VM ready notification, we need to find the FunctionMeta of that VM and signal
	//because there might be someone waiting
	if req.Status == pb.StatusUpdateRequest_READY {
		s.States.Resources.ReplicasLock.Lock()
		defer s.States.Resources.ReplicasLock.Unlock()
		s.States.Resources.FunctionsLock.RLock()
		defer s.States.Resources.FunctionsLock.RUnlock()

		s.States.Resources.Replicas[vmid].Status = models.VMIdle

		//get the replica function's meta and add it to idle vms so someone can get it
		fnmeta := s.States.Resources.Functions[rep.Function]
		fnmeta.Lock()
		//fnmeta.IdleReplicas[vmid] = s.Resources.Replicas[vmid]
		fnmeta.Replicas[vmid] = s.States.Resources.Replicas[vmid]
		fnmeta.Unlock()

		//if it's ready, mark as idle
		log.Debug().
			Str("vmid", vmid).
			Str("ip", s.States.Resources.Replicas[vmid].IP).
			Str("fn", rep.Function).
			Int("idleFns", rep.IdleExecutionEnvs).
			Msg("VM ready, added to pool")

		//we dont need the lock  to signal according to sync's docs
		//fnmeta.WaitingForVM.Signal()
		return &pb.StatusUpdateReply{Action: pb.StatusUpdateReply_KEEP_ALIVE}, nil
	} else if req.Status == pb.StatusUpdateRequest_FINISHED {
		s.States.Resources.ReplicasLock.Lock()
		defer s.States.Resources.ReplicasLock.Unlock()
		s.States.Resources.FunctionsLock.RLock()
		defer s.States.Resources.FunctionsLock.RUnlock()

		fnmeta := s.States.Resources.Functions[rep.Function]

		s.States.Resources.GPUMappingLock.Lock()
		gpuRescUsed, exists := s.States.Resources.GpuMapping[vmid]
		if exists {
			gpuUUID := gpuRescUsed.AvAReply.GpuUuid
			gpuNode := gpuRescUsed.GPUNode
			gpuNode.Lock()
			allocated := gpuNode.AllocatedGPUResc[gpuUUID]
			allocated.AllocatedMem -= fnmeta.GPUMemReq
			allocated.AllocatedWorker[gpuRescUsed.GPUIdx] = false
			gpuNode.Unlock()
			delete(s.States.Resources.GpuMapping, vmid)
		}
		s.States.Resources.GPUMappingLock.Unlock()

		//ask policy what to do
		kill, reps := s.States.ScalePolicy.OnFunctionCompletion()
		go s.States.scaleUpFunction(fnmeta, reps)

		if kill {
			log.Debug().
				Str("vmid", vmid).
				Str("ip", s.States.Resources.Replicas[vmid].IP).
				Msg("VM finished, policy is kill")
			s.States.Resources.Replicas[vmid].DeployNode.Grpc.ReleaseVMSetup(ctx,
				&fpb.ReleaseRequest{Vmid: vmid})
			delete(s.States.Resources.Replicas, vmid)
			return &pb.StatusUpdateReply{Action: pb.StatusUpdateReply_DESTROY}, nil
		}

		//log.Debug().
		//	Str("vmid", vmid).
		//	Str("ip", s.States.Resources.Replicas[vmid].IP).
		//	Msg("VM finished, put back in pool")
		s.States.Resources.Replicas[vmid].Status = models.VMIdle

		fnmeta.Lock()
		//fnmeta.IdleReplicas[vmid] = s.Resources.Replicas[vmid]
		fnmeta.Replicas[vmid].IdleExecutionEnvs++
		fnmeta.Unlock()

		//fnmeta.WaitingForVM.Signal()
		return &pb.StatusUpdateReply{Action: pb.StatusUpdateReply_KEEP_ALIVE}, nil
	}

	log.Fatal().
		Str("vmid", vmid).
		Str("ip", s.States.Resources.Replicas[vmid].IP).
		Msg("RM: VM UNKOWN status update")
	return &pb.StatusUpdateReply{Action: pb.StatusUpdateReply_DESTROY}, nil
}

//RegisterNode ..
func (s *ResourceManagerServer) RegisterNode(ctx context.Context, req *pb.RegisterNodeRequest) (*empty.Empty, error) {
	//connect to Node gRpc
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	ip := req.Ip + consts.PortToString(int(req.Port))
	log.Info().
		Str("ip", ip).
		Msg("RM: registering node, connecting back..")
	conn, err := grpc.Dial(ip, opts...)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("fail to dial")
	}
	client := fpb.NewFcServiceClient(conn)

	log.Info().Msgf("RM: connected back to node (%v), register done", ip)

	node := models.Node{
		HostIP: req.Ip,
		Grpc:   client,
		Conn:   conn,
		NodeID: req.NodeId,
	}

	s.States.Resources.NodesLock.Lock()
	defer s.States.Resources.NodesLock.Unlock()
	s.States.Resources.Nodes = append(s.States.Resources.Nodes, &node)
	return &empty.Empty{}, nil
}

func (states *ResourceManagerStates) scaleUpFunction(fnmeta *models.FunctionMeta, n uint32) {
	//we must scale up a function, this is where we make decision of which nodes and how many
	//param n is a suggestion if it's 0, otherwise its the amount we want
	if n != 0 {
		log.Debug().
			Uint32("num", n).
			Str("fn", fnmeta.Name).
			Msg("RM: scaling up function")
	}

	c := make(chan *models.Replica)

	stime := time.Now()
	//this is not pretty, but the scale thing can create more than n, so we get as return
	n = states.PlacementPolicy.OnScaleReplicaOnNode(states.Resources.Nodes, fnmeta, n, "", c)

	if n == 0 {
		//dont waste time if not spawning anything
		close(c)
		return
	}

	//update the replica map
	for i := uint32(0); i < n; i++ {
		rep := <-c
		//log.Printf("ScaleUpFunction: from channel, adding to replica map (%v)", rep.Vmid)
		states.Resources.ReplicasLock.Lock()
		states.Resources.Replicas[rep.Vmid] = rep
		states.Resources.ReplicasLock.Unlock()
	}
	duration := time.Since(stime)
	log.Debug().Dur("tototal start time", duration).
		Uint32("num vm", n).
		Int("Rep len", len(states.Resources.Replicas)).
		Str("fn", fnmeta.Name).
		Msg("startup time measured in resource manager")
	close(c)
}

//RequestVM this function is called when the scheduler can't find a vm for a function, so we need to create one or more
//this call is non-blocking, which means that after it returns there is no guarantee a VM is available
func (states *ResourceManagerStates) RequestVM(ctx context.Context, req *ResourceRequest) error {
	fnmeta, exists := states.Resources.Functions[req.Function]
	if !exists {
		return fmt.Errorf("function (%v) is not deployed", req.Function)
	}

	log.Info().Str("name", fnmeta.Name).Msg("RM: Someone requesting new VM")

	// ask scale policy if we should wait or return "retry"
	var reps uint32
	if reps = states.ScalePolicy.OnRequestNoneAvailable(); reps == 0 {
		return errors.New("policy says we cant spawn more VMs, wait or retry")
	}

	//scale up
	states.scaleUpFunction(fnmeta, reps)

	return nil
}
