package schedulers

import (
	"faas-memory/pkg/dataflow"
	"faas-memory/pkg/models"
	"faas-memory/pkg/resmngr"
	"fmt"

	"github.com/rs/zerolog/log"
)

//RoundRobinScheduler ..

//RoundRobinScheduler ..
type RoundRobinScheduler struct {
	dataflow.BaseScheduler
	nodes    []*models.Node
	lastNode int
}

// NewRoundRobinScheduler creates a new Scheduler
func NewRoundRobinScheduler(rescStates *resmngr.ResourceManagerStates) *RoundRobinScheduler {
	//create grpc client to resmngr
	rs := &RoundRobinScheduler{
		BaseScheduler: dataflow.NewBaseScheduler(rescStates),
		lastNode:      0,
	}
	rs.ChildScheduler = rs
	return rs
}

func (s *RoundRobinScheduler) insertIfNotIn(node *models.Node) {
	for _, n := range s.nodes {
		if node == n {
			return
		}
	}
	log.Debug().
		Str("hostIP", node.HostIP).
		Msg("havent seen this node before, adding to node list")
	s.nodes = append(s.nodes, node)
}

//ChooseReplica ..
func (s *RoundRobinScheduler) ChooseReplica(task *dataflow.TaskState) (string, string, uint32, int,
	*models.GPUNode, error) {
	fnName := task.GetFunctionName()
	fnMeta, ok := s.GetRescStates().Resources.Functions[fnName]
	if !ok {
		return "", "", 0, 0, nil, fmt.Errorf("function not deployed")
	}
	fnMeta.Lock()
	defer fnMeta.Unlock()

	//search for nodes we havent seen
	for k := range fnMeta.Replicas {
		node := fnMeta.Replicas[k].DeployNode
		s.insertIfNotIn(node)
	}

	start := s.lastNode
	for {
		s.lastNode = (s.lastNode + 1) % len(s.nodes)
		destNode := s.nodes[s.lastNode]

		//if we did a whole scan, return that we cant find
		if start == s.lastNode {
			//log.Debug().
			//	Str("sched", "roundrobin").
			//	Msg("whole node scan with no idle reps, returning to sched")
			return "", "", 0, 0, nil, fmt.Errorf("no vms")
		}

		for k, r := range fnMeta.Replicas {
			if r.DeployNode == destNode {
				//if there are no EEs on this VM
				if fnMeta.Replicas[k].IdleExecutionEnvs == 0 {
					break
				}

				//TODO: the GPU memory thing
				fnMeta.Replicas[k].IdleExecutionEnvs--
				log.Debug().
					Int("nodeIndex", s.lastNode).
					Str("vmid", k).
					Msg("going round the node list, vm chosen")
				return k, "", 0, 0, nil, nil
			}
		}

		//log.Debug().
		//	Str("sched", "roundrobin").
		//	Msg("couldn't find a replica on next node, going one more")
	}
}

//RunScheduler ..
func (s *RoundRobinScheduler) RunScheduler(schedulerChan <-chan dataflow.SchedulableMessage) error {
	//st := NewSchedulerStats()
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
			//st.AddNode(task.ExecutorAddress.Hostname())
			//tell the function it can continue
			msg.Signal.L.Lock()
			task.IsReady = true
			msg.Signal.Signal()
			msg.Signal.L.Unlock()
		}()
	}
	//st.PrintNodeDistribution()

	return nil
}
