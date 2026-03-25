package database

import (
	"bytes"
	"fmt"

	bytehandler "github.com/execute-assembly/c2-proj/newserver/internal/bytes"
	_ "modernc.org/sqlite"
)

func RegisterClient(data *bytes.Reader, IpAddress string) error {
	//	db, err := GetDB()
	// if err != nil {
	// 	return err
	// }
	clientData, err := bytehandler.ParseClientRegister(data, IpAddress)
	if err != nil {
		return err
	}
	fmt.Printf("Guid: %s\n", clientData.Guid)
	fmt.Printf("Username: %s\n", clientData.Username)
	fmt.Printf("Hostname: %s\n", clientData.Hostname)
	fmt.Printf("IP: %s\n", clientData.Ip)
	fmt.Printf("Arch: %d\n", clientData.Arch)
	fmt.Printf("WinVersion: %s\n", clientData.WinVersion)
	fmt.Printf("PID: %d\n", clientData.Pid)
	return nil
}
