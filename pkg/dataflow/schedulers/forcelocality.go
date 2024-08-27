package schedulers

import (
	"faas-memory/pkg/dataflow"
	"faas-memory/pkg/models"
	"faas-memory/pkg/resmngr"
	"faas-memory/pkg/resmngr/policy/gpu"
	"fmt"

	"github.com/rs/zerolog/log"
)

//ForceLocalityScheduler implements force locality scheduling
type ForceLocalityScheduler struct {
	dataflow.BaseScheduler
	Random *RandomScheduler
}

// NewForceLocalityScheduler creates a new Scheduler
func NewForceLocalityScheduler(rescStates *resmngr.ResourceManagerStates) *ForceLocalityScheduler {
	rs := &ForceLocalityScheduler{
		BaseScheduler: dataflow.NewBaseScheduler(rescStates),
		Random:        NewRandomScheduler(rescStates),
	}

	rs.ChildScheduler = rs
	return rs
}

//ChooseReplica ..
func (s *ForceLocalityScheduler) ChooseReplica(task *dataflow.TaskState) (string, string,
	uint32, int, *models.GPUNode, error) {
	fnName := task.GetFunctionName()
	//get function meta and lock
	fnMeta, ok := s.GetRescStates().Resources.Functions[fnName]
	if !ok {
		return "", "", 0, 0, nil, fmt.Errorf("function not deployed")
	}
	fnMeta.Lock()
	defer fnMeta.Unlock()

	//if there are no replicas, return err so that Schedule can talk to resmngr
	if len(fnMeta.Replicas) == 0 {
		return "", "", 0, 0, nil, fmt.Errorf("no vms")
	}

	rescs := s.GetRescStates().Resources
	//if this function requires a GPU, we need to lock GPUs
	if fnMeta.GPUMemReq != 0 {
		rescs.GPULock.Lock()
		defer rescs.GPULock.Unlock()
	}
	//common preamble done

	previousNodeIP := task.SchedulingHints.ScheduledNodeIP
	//if we dont have previous, do random
	if previousNodeIP == "" {
		return s.Random.RandomChoiceWithLocks(fnMeta, rescs.GPUNodes,
			rescs.NumWorkerPerGPU)
	}

	firstFitPolicy := gpu.FirstFitMemory{}
	for vmid, rep := range fnMeta.Replicas {
		//if this replica isn't available, dont waste time
		if rep.IdleExecutionEnvs == 0 {
			continue
		}
		//if ip is same, we good
		if rep.DeployNode.HostIP == previousNodeIP {
			//if we need gpu, check for any that fit and use
			if fnMeta.GPUMemReq > 0 {
				rescs.GPUMappingLock.RLock()
				uuid, workerPort, gpuIdx, gpuNode := firstFitPolicy.OnChooseGPU(fnMeta.GPUMemReq,
					rescs.GPUNodes, rescs.NumWorkerPerGPU)
				rescs.GPUMappingLock.RUnlock()
				if uuid != "" {
					fnMeta.Replicas[vmid].IdleExecutionEnvs--
					return vmid, uuid, workerPort, gpuIdx, gpuNode, nil
				}
			} else {
				log.Debug().
					Str("sched", "forcelocality").
					Str("task", fnName).
					Str("target", rep.DeployNode.HostIP).
					Msg("found replica in same host")
				fnMeta.Replicas[vmid].IdleExecutionEnvs--
				return vmid, "", 0, 0, nil, nil
			}
		}
	}

	return "", "", 0, 0, nil, fmt.Errorf("no vms")
}

//RunScheduler ..
func (s *ForceLocalityScheduler) RunScheduler(schedulerChan <-chan dataflow.SchedulableMessage) error {
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
