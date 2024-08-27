package schedulers

import (
	"faas-memory/pkg/dataflow"
	"faas-memory/pkg/models"
	"faas-memory/pkg/resmngr"
	"faas-memory/pkg/resmngr/policy/gpu"
	"fmt"

	"github.com/rs/zerolog/log"
)

type GalvanicScheduler struct {
	dataflow.BaseScheduler
	Random *RandomScheduler
}

// NewGalvanicScheduler creates a new Scheduler
func NewGalvanicScheduler(rescStates *resmngr.ResourceManagerStates) *GalvanicScheduler {
	rs := &GalvanicScheduler{
		BaseScheduler: dataflow.NewBaseScheduler(rescStates),
		Random:        NewRandomScheduler(rescStates),
	}

	rs.ChildScheduler = rs
	return rs
}

//ChooseReplica ..
func (s *GalvanicScheduler) ChooseReplica(task *dataflow.TaskState) (string, string, uint32, int,
	*models.GPUNode, error) {
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
	//common preamble done

	previousNodeIP := ""
	uploads := 0
	if task.SchedulingHints != nil {
		previousNodeIP = task.SchedulingHints.ScheduledNodeIP
		uploads = task.SchedulingHints.Uploads
	}
	//if we dont have previous, do random
	if previousNodeIP == "" {
		return s.Random.RandomChoiceWithLocks(fnMeta, rescs.GPUNodes, rescs.NumWorkerPerGPU)
	}

	//log.Debug().Msgf("prev task uploaded %v", task.SchedulingHints.Uploads)

	if uploads == 0 {
		return s.Random.RandomChoiceWithLocks(fnMeta, rescs.GPUNodes, rescs.NumWorkerPerGPU)
	}

	bestFitPolicy := gpu.BestFitMemory{}
	for vmid, rep := range fnMeta.Replicas {
		//if this replica isn't available, dont waste time
		if rep.IdleExecutionEnvs == 0 {
			continue
		}
		//lets be greedy, find first same host and return
		if rep.DeployNode.HostIP == previousNodeIP {
			log.Debug().Str("previous", previousNodeIP).Msgf("got same %v", rep.DeployNode.HostIP)
			//if it requires GPU mem, we need to find those that have room
			if fnMeta.GPUMemReq > 0 {
				rescs.GPUMappingLock.RLock()
				bestUUID, workerPort, gpuIdx, bestGPUNode := bestFitPolicy.OnChooseGPU(fnMeta.GPUMemReq, rescs.GPUNodes,
					rescs.NumWorkerPerGPU)
				rescs.GPUMappingLock.RUnlock()
				//if we find a GPU, mark node and gpu as used
				if bestUUID != "" {
					fnMeta.Replicas[vmid].IdleExecutionEnvs--
					return vmid, bestUUID, workerPort, gpuIdx, bestGPUNode, nil
				}
			} else {
				fnMeta.Replicas[vmid].IdleExecutionEnvs--
				return vmid, "", 0, 0, nil, nil
			}
		}
	}

	//if we cant find any, random
	//log.Debug().Msgf("no match, trying random")
	return s.Random.RandomChoiceWithLocks(fnMeta, rescs.GPUNodes, rescs.NumWorkerPerGPU)
}

//RunScheduler ..
func (s *GalvanicScheduler) RunScheduler(schedulerChan <-chan dataflow.SchedulableMessage) error {
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
