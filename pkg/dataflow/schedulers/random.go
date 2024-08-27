package schedulers

import (
	"faas-memory/pkg/dataflow"
	"faas-memory/pkg/models"
	"faas-memory/pkg/resmngr"
	"faas-memory/pkg/resmngr/policy/gpu"
	"fmt"
	"math/rand"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

// RandomScheduler is a struct that implements random selection of VM
type RandomScheduler struct {
	dataflow.BaseScheduler
}

// NewRandomScheduler creates a new random Scheduler
func NewRandomScheduler(rescStates *resmngr.ResourceManagerStates) *RandomScheduler {
	rs := &RandomScheduler{
		BaseScheduler: dataflow.NewBaseScheduler(rescStates),
	}

	rs.ChildScheduler = rs
	return rs
}

//Choice contains one possible VM and GPU assignment
type Choice struct {
	vmid, gpuUUID string
	gpuNode       *models.GPUNode
	workerPort    uint32
	gpuIdx        int
}

//RandomChoiceWithLocks returns vmid, gpuUUID, gpu worker port, error
func (s *RandomScheduler) RandomChoiceWithLocks(fnMeta *models.FunctionMeta,
	gpuNodes []*models.GPUNode, numWorkerPerGPU uint32) (string, string, uint32, int, *models.GPUNode, error) {
	fnName := fnMeta.Name
	var reps []Choice

	for vmid, rep := range fnMeta.Replicas {
		//if this replica isn't available, dont waste time
		if rep.IdleExecutionEnvs == 0 {
			continue
		}
		//if no GPU is needed, at this point we know the rep is available, so add
		candidate := Choice{vmid: vmid}
		reps = append(reps, candidate)
	}

	//none available, return nothing
	if len(reps) == 0 {
		//log.Error().Str("sched", "random").Msg("No availalbe VM for fn")
		return "", "", 0, 0, nil, fmt.Errorf("no vms on random choice, total (%v)", len(fnMeta.Replicas))
	}

	//choose random one
	chosen := reps[rand.Intn(len(reps))]
	log.Debug().Str("sched", "random").
		Int("idleReplicas", len(reps)).
		Int("totalReplicas", len(fnMeta.Replicas)).
		Str("vmid", chosen.vmid).
		Str("task", fnName).
		Msg("random VM was chosen")

	//if we need GPU, need to update the memory used
	if fnMeta.GPUMemReq > 0 {
		rescs := s.GetRescStates().Resources
		rescs.GPUMappingLock.RLock()
		defer rescs.GPUMappingLock.RUnlock()

		bestFitPolicy := gpu.BestFitMemory{}
		bestUUID, workerPort, gpuIdx, bestGPUNode := bestFitPolicy.OnChooseGPU(fnMeta.GPUMemReq, gpuNodes,
			numWorkerPerGPU)
		log.Debug().Str("gpu uuid", bestUUID).Msg("best gpu is chosen")
		//if we can't find a GPU
		if bestUUID == "" {
			return "", "", 0, 0, nil, fmt.Errorf("no vms")
		}
		chosen.gpuNode = bestGPUNode
		chosen.gpuUUID = bestUUID
		chosen.workerPort = workerPort
		chosen.gpuIdx = gpuIdx
	}

	//mark it as used
	fnMeta.Replicas[chosen.vmid].IdleExecutionEnvs--
	atomic.AddUint32(&fnMeta.Replicas[chosen.vmid].DeployNode.TimesUsed, 1)

	return chosen.vmid, chosen.gpuUUID, chosen.workerPort, chosen.gpuIdx, chosen.gpuNode, nil
}

//ChooseReplica ..
func (s *RandomScheduler) ChooseReplica(task *dataflow.TaskState) (string, string, uint32, int, *models.GPUNode, error) {
	fnName := task.GetFunctionName()
	rescs := s.GetRescStates().Resources
	//get function meta and lock
	fnMeta, ok := rescs.Functions[fnName]
	if !ok {
		return "", "", 0, 0, nil, fmt.Errorf("function not deployed")
	}
	fnMeta.Lock()
	defer fnMeta.Unlock()

	//if there are no replicas, return err so that Schedule can talk to resmngr
	if len(fnMeta.Replicas) == 0 {
		return "", "", 0, 0, nil, fmt.Errorf("no vms")
	}

	//if this function requires a GPU, we need to lock GPUs
	if fnMeta.GPUMemReq != 0 {
		rescs.GPULock.Lock()
		defer rescs.GPULock.Unlock()
	}

	//do a random choice, in different function for reusability
	return s.RandomChoiceWithLocks(fnMeta, rescs.GPUNodes,
		rescs.NumWorkerPerGPU)
}

//RunScheduler ..
func (s *RandomScheduler) RunScheduler(schedulerChan <-chan dataflow.SchedulableMessage) error {
	for {
		msg := <-schedulerChan
		//if it's over let's quit
		if msg.End {
			break
		}

		//go async to schedule more than one fn at a time
		go func() {
			//we only get task states for now, so lets convert
			task := msg.State.(*dataflow.TaskState)
			//choose where the function will be executed and fill the three required fields
			s.Schedule(task)
			//tell the function it can continue
			msg.Signal.L.Lock()
			task.IsReady = true
			msg.Signal.Signal()
			msg.Signal.L.Unlock()
		}()
	}
	return nil
}
