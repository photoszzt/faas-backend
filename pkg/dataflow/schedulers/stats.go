package schedulers

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

//SchedulerStats ..
type SchedulerStats struct {
	nodeDistribution map[string]int
	//n                int
	sync.Mutex
}

//NewSchedulerStats ..
func NewSchedulerStats() *SchedulerStats {
	return &SchedulerStats{
		nodeDistribution: make(map[string]int),
	}
}

//AddNode ..
func (s *SchedulerStats) AddNode(node string) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.nodeDistribution[node]; !ok {
		s.nodeDistribution[node] = 0
	}
	s.nodeDistribution[node]++
	//log.Debug().Msgf("node %v has %v", node, s.nodeDistribution["node"])
}

//PrintNodeDistribution ..
func (s *SchedulerStats) PrintNodeDistribution() {
	var sum float32 = 0
	for _, v := range s.nodeDistribution {
		sum += float32(v)
	}

	out := ""
	for _, v := range s.nodeDistribution {
		vv := float32(v)
		out = fmt.Sprintf("%s/%.2f", out, vv/sum)
	}

	log.Info().Msgf("distribution:   %s", out[1:])
}
