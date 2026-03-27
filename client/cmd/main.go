package main

import (
	"context"
	"fmt"
	"os"

	"github.com/execute-assembly/c2-proj/client/internal/client"
	"github.com/execute-assembly/c2-proj/client/internal/commander"
	clientdb "github.com/execute-assembly/c2-proj/client/internal/db"
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

	if err := client.Init(); err != nil {
		fmt.Println("Failed to init readline:", err)
		os.Exit(1)
	}
	commander.Out = client.RL.Stdout()

	go func() {
		stream, err := rpc.Client.Subscribe(context.Background(), &emptypb.Empty{})
		if err != nil {
			commander.PrintErr("Failed Connecting to Subscribe endpoint")
			return
		}
		for {
			event, err := stream.Recv()
			if err != nil {
				return
			}
			switch event.EventType {
			case commander.EVENT_NEW_AGENT:
				commander.PrintInfo("New Agent Connected!")
				fmt.Fprintln(commander.Out, event.Data)
			case commander.EVENT_COMMAND_OUTPUT:
				commander.PrintOutput(event.TaskId, event.Guid, event.Data)
				clientdb.MarkExecuted(event.TaskId)
			}
		}
	}()

	client.RunClient()
}
