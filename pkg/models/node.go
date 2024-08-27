package models

import (
	"context"
	"faas-memory/pb/fcdaemon"
	"sync"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// Node contains the ip address of the node
type Node struct {
	sync.RWMutex
	HostIP    string
	NodeID    string
	Conn      *grpc.ClientConn
	Grpc      fcdaemon.FcServiceClient
	TimesUsed uint32
	//dont trust TotalCreated unless you are in loadbalance.go, since it's the only thing that updates this number
	TotalCreated uint32
}

// RequestCreateReplica send the CreateRequest rpc and returns the replied ips
//this is sent to fcdaemon
func (n *Node) RequestCreateReplica(req *fcdaemon.CreateRequest) (*fcdaemon.CreateReply, error) {
	ctx := context.Background()
	vm, err := n.Grpc.CreateReplica(ctx, req)
	if err != nil {
		log.Fatal().Err(err).Msgf("%v.CreateReplica(_) = _", n.Grpc)
	}

	return vm, nil
}

//CreateCPUHogs ..
func (n *Node) CreateCPUHogs(num int32) {
	ctx := context.Background()
	req := fcdaemon.CpuHogs{N: num}

	_, err := n.Grpc.CreateCPUHogs(ctx, &req)
	if err != nil {
		log.Fatal().Err(err).Msgf("%v.CreateCPUHogs(_) = _", n.Grpc)
	}
}

//StopCPUHogs ..
func (n *Node) StopCPUHogs(num int32) {
	ctx := context.Background()
	req := fcdaemon.CpuHogs{N: num}

	_, err := n.Grpc.StopCPUHogs(ctx, &req)
	if err != nil {
		log.Fatal().Err(err).Msgf("%v.CreateCPUHogs(_) = _", n.Grpc)
	}
}
