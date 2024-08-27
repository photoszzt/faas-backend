package models

import (
	"sync"
)

//GPUInfo contains the info for a GPU
type GPUInfo struct {
	UUID            string
	Memory          uint32
	FirstWorkerPort uint32
}

//GPUNode contains gpu info for a gpu worker machine
type GPUNode struct {
	sync.RWMutex
	GPUInfos         []*GPUInfo
	AllocatedGPUResc map[string]*GPUResc
	IP               string
}

type GPUResc struct {
	AllocatedMem    uint32
	AllocatedWorker []bool
}
