package rpc

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"

	pb "github.com/execute-assembly/c2-proj/modules/pb"
	"github.com/execute-assembly/c2-proj/server/internal/database"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CommandData struct {
	TaskID      int32
	CommandCode int
	Param1      string
	Param2      string
}

type PendingTasksMap struct {
	CommandCode int
	Guid        string
	tasked_at   string
	Param1      string
	Param2      string
}

var (
	CommandMap = map[string][]CommandData{}
	CommandMu  sync.RWMutex

	PendingTasks = map[int32]PendingTasksMap{}
	PendingMu    sync.RWMutex
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

func BroadcastEvent(eventType int32, guid, data string, taskId int32) {
	subMu.Lock()
	defer subMu.Unlock()
	for _, s := range subscribers {
		s.Send(&pb.ServerEvent{EventType: eventType, Guid: guid, Data: data, TaskId: taskId})
	}
}

func (s *Server) ListClients(ctx context.Context, req *emptypb.Empty) (*pb.ListClientResp, error) {
	clients, err := database.ListClients_db()
	if err != nil {
		return nil, err
	}

	var entries []*pb.ClientEntry
	for _, c := range clients {
		entries = append(entries, &pb.ClientEntry{
			Guid:        c.Guid,
			CodeName:    c.CodeName,
			Username:    c.Username,
			Hostname:    c.Hostname,
			Ip:          c.Ip,
			Arch:        c.Arch,
			Pid:         c.Pid,
			Version:     c.Version,
			LastCheckin: c.LastCheckin,
		})
	}
	return &pb.ListClientResp{Clients: entries}, nil
}

var counter int32

func GenerateTaskId() int32 {
	return atomic.AddInt32(&counter, 1)
}

func (s *Server) SendCommand(ctx context.Context, req *pb.CommandReqData) (*pb.CommandRespData, error) {
	exists, err := database.CheckIfUserExists_db(req.Guid)
	if err != nil {
		return &pb.CommandRespData{Status: 1, Message: err.Error()}, err
	}
	if !exists {
		return &pb.CommandRespData{Status: 1, Message: "User Doesnt Exist!"}, nil
	}

	TaskId := GenerateTaskId()
	CommandMu.Lock()
	CommandMap[req.Guid] = append(CommandMap[req.Guid], CommandData{
		TaskID:      TaskId,
		CommandCode: int(req.CommandCode),
		Param1:      req.Param,
		Param2:      req.Param2,
	})
	CommandMu.Unlock()

	return &pb.CommandRespData{Status: 0, Message: "Command Queued!", TaskId: TaskId}, nil
}

func (s *Server) ConvertCodeName(ctx context.Context, req *pb.ConvertCodeMessage) (*pb.ConvertCodeResp, error) {
	guid, err := database.ConvertCodeName_db(req.CodeName)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.ConvertCodeResp{Status: 3, ErrorMsg: "User Not Found"}, nil
		}
		return &pb.ConvertCodeResp{Status: 1, ErrorMsg: err.Error()}, err
	}

	return &pb.ConvertCodeResp{Guid: guid, Status: 0}, nil
}
