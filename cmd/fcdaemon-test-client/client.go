package main

import (
	"context"
	pb "faas-memory/pb/fcdaemon"
	"flag"
	"io"
	"log"

	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:50051", "The server address in the format of host:port")
)

/*
const (
	firecrackerDefaultPath = "firecracker"
)
*/

func createReplicaRequest(client pb.FcServiceClient) {
	req := &pb.CreateRequest{
		KernelImagePath:    "/home/xxx/.capstan/loader.elf",
		KernelArgs:         "--verbose --nopci /libserver.so",
		ImageName:          "/home/xxx/.capstan/repository/sprocket-tf-cpu/sprocket-tf-cpu.raw",
		ImageStoragePrefix: "/home/xxx/.capstan/storage",
		InetDev:            "enp4s0",
		NameServer:         "128.83.120.181",
		AwsAccessKeyID:     "",
		AwsSecretAccessKey: "",
		Cpu:                4,
		Mem:                1024,
		NumReplica:         10,
	}
	ctx := context.Background()
	stream, err := client.CreateReplica(ctx, req)
	if err != nil {
		log.Fatalf("%v.CreateReplica(_) = _, %v", client, err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			in := stream
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("failed to receive a ip: %v", err)
			}
			log.Printf("Got ip %s", in.Ip)
		}
	}()
	<-waitc
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewFcServiceClient(conn)
	createReplicaRequest(client)
}
