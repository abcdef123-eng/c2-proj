package rpc

import (
	"sync"
)

type CommandData struct {
	CommandCode int
	Param1      string
	Param2      string
}

var (
	CommandMap = map[string][]CommandData{}
	CommandMu  sync.RWMutex
)

// func (s *Server) SendCommand(ctx context.Context, req *pb.CommandReqData) (*pb.CommandRespData, error) {

// 	UserGuid := req.Guid

// 	CommandMu.Lock()
// 	CommandMap[UserGuid] = append(CommandMap[UserGuid], CommandData{
// 		CommandCode: int(req.CommandCode),
// 		Param1:      req.Param,
// 		Param2:      req.Param2})
// 	CommandMu.Unlock()

// }
