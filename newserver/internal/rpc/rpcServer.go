package rpc

import (
	"net"

	pb "github.com/execute-assembly/c2-proj/modules/pb"
	"google.golang.org/grpc"
)

type Server struct {
	pb.C2Service
}

func runRpcServer() error {
	lis, err := net.Listen("tcp", ":50001")
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterMyServiceServer

}
