package rpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/execute-assembly/c2-proj/modules/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func RunRpcClient() error {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("ERROR !2")

		return err
	}

	defer conn.Close()

	client := pb.NewC2ServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.InsertIntoDatabase(ctx, &pb.CommandReqData{
		Guid:        "1111-1111-1111-1111",
		CommandCode: 1,
		Param:       "hello",
		Param2:      "hello",
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Message: %s\n", resp.Message)
	return nil

}
