package rpc

import (
	"fmt"
	"net"

	"github.com/execute-assembly/c2-proj/server/internal/config"
	pb "github.com/execute-assembly/c2-proj/modules/pb"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedC2ServiceServer
}

func RunRpcServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Cfg.GrpcPort))
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterC2ServiceServer(s, &Server{})
	go func() {
		s.Serve(lis)
	}()
	return nil
}
