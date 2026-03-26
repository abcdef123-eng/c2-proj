package rpc

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"

	pb "github.com/execute-assembly/c2-proj/modules/pb"
	"github.com/execute-assembly/c2-proj/newserver/internal/database"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CommandData struct {
	TaskID      int32
	CommandCode int
	Param1      string
	Param2      string
}

// type DownloadTask struct {
// 	FileName   string

// }

var (
	CommandMap = map[string][]CommandData{}
	CommandMu  sync.RWMutex
)

var (
	subscribers []pb.C2Service_SubscribeServer
	subMu       sync.Mutex
)

func (s *Server) Subscribe(req *emptypb.Empty, stream pb.C2Service_SubscribeServer) error {
	subMu.Lock()
	subscribers = append(subscribers, stream)
	subMu.Unlock()

	<-stream.Context().Done()
	return nil
}

func BroadcastEvent(eventType, guid, data string) {
	subMu.Lock()
	defer subMu.Unlock()
	for _, s := range subscribers {
		s.Send(&pb.ServerEvent{EventType: eventType,
			Guid: guid, Data: data})
	}
}

func (s *Server) ListClients(ctx context.Context, req *emptypb.Empty) (*pb.ListClientResp, error) {
	Clients, err := database.ListClients_db()
	if err != nil {
		return nil, err
	}

	var grpcClient []*pb.ClientEntry
	for _, c := range Clients {
		grpcClient = append(grpcClient, &pb.ClientEntry{
			Guid:        c.Guid,
			CodeName:    c.Code_name,
			Username:    c.Username,
			Hostname:    c.Hostname,
			Ip:          c.Ip,
			Arch:        c.Arch,
			Pid:         c.Pid,
			Version:     c.Version,
			LastCheckin: c.Last_checkin,
		})
	}
	return &pb.ListClientResp{Clients: grpcClient}, nil
}

var counter int32

func GenerateTaskId() int32 {
	return atomic.AddInt32(&counter, 1)
}

func (s *Server) SendCommand(ctx context.Context, req *pb.CommandReqData) (*pb.CommandRespData, error) {

	UserGuid := req.Guid
	exists, err := database.CheckIfUserExists_db(UserGuid)
	if err != nil {
		return &pb.CommandRespData{Status: 1, Message: err.Error()}, err
	}
	if !exists {
		return &pb.CommandRespData{Status: 1, Message: "User Doesnt Exist!"}, nil
	}

	CommandMu.Lock()
	CommandMap[UserGuid] = append(CommandMap[UserGuid], CommandData{
		TaskID:      GenerateTaskId(),
		CommandCode: int(req.CommandCode),
		Param1:      req.Param,
		Param2:      req.Param2})
	CommandMu.Unlock()

	return &pb.CommandRespData{Status: 0, Message: "Command Queued!"}, nil

}

func (s *Server) ConvertCodeName(ctx context.Context, req *pb.ConvertCodeMessage) (*pb.ConvertCodeResp, error) {
	Guid, err := database.ConvertCodeName_db(req.CodeName)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.ConvertCodeResp{Guid: "", Status: 3, ErrorMsg: "User Not Found"}, nil
		}
		return &pb.ConvertCodeResp{Guid: "", Status: 1, ErrorMsg: err.Error()}, err
	}

	return &pb.ConvertCodeResp{Guid: Guid, Status: 0, ErrorMsg: ""}, err
}
