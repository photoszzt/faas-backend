package models

import (
	"faas-memory/pb/mngrinfo"
	"sync"
)

//GPUChosen contains info of chosen gpu and the node that the gpu is on
type GPUChosen struct {
	AvAReply *mngrinfo.AvAManagerInfoReply
	GPUNode  *GPUNode
	GPUIdx   int
}

// Resources contains all resources the resource manager manges
type Resources struct {
	NodesLock sync.RWMutex
	Nodes     []*Node //map of available nodes

	GPUMappingLock sync.RWMutex          //protect the GPUChosen entry
	GpuMapping     map[string]*GPUChosen //map between vmid to gpu uuid

	ReplicasLock sync.RWMutex
	Replicas     map[string]*Replica //maps vmid to replica

	FunctionsLock sync.RWMutex
	Functions     map[string]*FunctionMeta //maps function names to function meta

	Bypass          bool
	NumWorkerPerGPU uint32

	GPULock  sync.RWMutex //protect the gpu nodes array
	GPUNodes []*GPUNode   //map of available gpu nodes
}
