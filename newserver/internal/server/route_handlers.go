package server

import (
	"bytes"

	bytehandler "github.com/execute-assembly/c2-proj/newserver/internal/bytes"
	"github.com/execute-assembly/c2-proj/newserver/internal/database"
)

func NewClientRegisterHandler(Ip string, reader *bytes.Reader) ([]byte, error) {
	ClientGuid, err := database.RegisterClient(reader, Ip)
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
	return JwtBytes, nil
}
