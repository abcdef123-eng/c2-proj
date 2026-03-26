package rpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/execute-assembly/c2-proj/modules/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var Client pb.C2ServiceClient
var Ready = make(chan struct{})

func RunRpcClient() error {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.Connect()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			break
		}
		if !conn.WaitForStateChange(ctx2, state) {
			return fmt.Errorf("could not connect to server")
		}
	}

	Client = pb.NewC2ServiceClient(conn)
	close(Ready)

	select {}
}
