package main

import (
	//"github.com/execute-assembly/c2-proj/client/internal/client"
	"context"
	"fmt"
	"os"

	"github.com/execute-assembly/c2-proj/client/internal/client"
	"github.com/execute-assembly/c2-proj/client/internal/commander"
	"github.com/execute-assembly/c2-proj/client/internal/rpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	errCh := make(chan error, 1)
	go func() {
		errCh <- rpc.RunRpcClient()
	}()

	select {
	case err := <-errCh:
		fmt.Println("Failed to connect to server:", err)
		os.Exit(1)
	case <-rpc.Ready:
	}

	go func() {
		stream, err := rpc.Client.Subscribe(context.Background(), &emptypb.Empty{})
		if err != nil {
			commander.PrintErr("Failed Connecting to Subsribe endpoint")
			return
		}
		for {
			event, err := stream.Recv()
			if err != nil {
				// server disconnected
				return
			}
			switch event.EventType {
			case "new_agent":
				commander.PrintInfo("New Agent Connected!")
				fmt.Println(event.Data)
			}
		}
	}()

	client.RunClient()
}
