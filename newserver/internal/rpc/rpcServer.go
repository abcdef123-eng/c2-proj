package rpc

import (
	"net"

	pb "github.com/abcdef123-eng/c2-proj/modules/pb"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedMyServiceServer
}

func runRpcServer() error {
	lis, err := net.Listen("tcp", ":50001")
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterMyServiceServer

}
