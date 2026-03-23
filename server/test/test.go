package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/http"
	"io"
)

type clientInfo struct {
	guid     string
	username string
	hostname string
	PID      uint32
	arch     uint32
	version  string
}

// helper function to convert a string to length-prefixed bytes
func CraftRegisterBytes() []byte {
	guid := "123e4567-e89b12d3a456-426614174000"
	username := "testuser"
	hostname := "myhost"
	PID := uint32(1234)
	arch := uint32(1)
	version := "1.0.0"

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(guid)))

	guidBytes := []byte(guid)
	buf.Write(guidBytes)

	binary.Write(&buf, binary.LittleEndian, uint32(len(username)))
	userBytes := []byte(username)
	buf.Write(userBytes)

	binary.Write(&buf, binary.LittleEndian, uint32(len(hostname)))
	hostBytes := []byte(hostname)
	buf.Write(hostBytes)

	binary.Write(&buf, binary.LittleEndian, PID)
	binary.Write(&buf, binary.LittleEndian, arch)

	binary.Write(&buf, binary.LittleEndian, uint32(len(version)))
	versionBytes := []byte(version)
	buf.Write(versionBytes)

	return buf.Bytes()
}



func ParseRegisterResponse(data []byte) (string, string) {
	reader := bytes.NewReader(data)

    var guidLen uint32
    if err := binary.Read(reader, binary.LittleEndian, &guidLen); err != nil {
            fmt.Println("Failed Reading guidLen");
            return "", ""
    }

    guidStr := make([]byte, guidLen)
    if _, err := reader.Read(guidStr); err != nil {
            fmt.Println("Failed Reading Guid");
            return "", ""
    }

     AuthToken := string(guidStr)
     

    var token2Len uint32
    if err := binary.Read(reader, binary.LittleEndian, &token2Len); err != nil {
            fmt.Println("Failed Reading guidLen");
            return "", ""
    }

    guid2 := make([]byte, token2Len)
    if _, err := reader.Read(guid2); err != nil {
            fmt.Println("Failed Reading Guid");
            return "", ""
    }
    RefreshToken := string(guid2)

    return AuthToken, RefreshToken
}

 /*
           [commandID 4 bytes]
           [CommandCode 4 bytes]
           [param1 Length 4 bytes]
           [param1 N bytes]
           [param2 Length 4 bytes]
           [param2 N bytes]
        */

type commandInfo struct {
	commandId uint32
	commandCode uint32
	param1 string
	param2 string
}

func parseCommand(data []byte) (*commandInfo) {
	reader := bytes.NewReader(data)

	cmd := &commandInfo{}

	var SuccessCheck uint32
	if err := binary.Read(reader, binary.LittleEndian, &SuccessCheck); err != nil {
		fmt.Println("[!] Failed Retreing command ID")
		return nil
	}
	if SuccessCheck > 0 {
		fmt.Println("[!] No Command")
		fmt.Println(SuccessCheck)
		return nil
	}

	if err := binary.Read(reader, binary.LittleEndian, &cmd.commandId); err != nil {
		fmt.Println("[!] Failed Retreing command ID")
		return nil
	}
	if err := binary.Read(reader, binary.LittleEndian, &cmd.commandCode); err != nil {
		fmt.Println("[!] Failed Retreing command Code")
		return nil
	}

	var param1Len uint32
	if err := binary.Read(reader, binary.LittleEndian, &param1Len); err != nil {
		fmt.Println("[!] Failed Retreing param1 len")
		return nil
	}

	param1Str := make([]byte, param1Len)
	if _, err := reader.Read(param1Str); err != nil {
            fmt.Println("Failed Reading userLen");
            return nil
    }

    cmd.param1 = string(param1Str)

    var param2Len uint32
	if err := binary.Read(reader, binary.LittleEndian, &param2Len); err != nil {
		fmt.Println("[!] Failed Retreing param1 len")
		return nil
	}

	param2str := make([]byte, param2Len)
	if _, err := reader.Read(param2str); err != nil {
            fmt.Println("Failed Reading userLen");
            return nil
    }

    cmd.param2 = string(param2str)

    return cmd



}


func CraftCommandOutput(commanID uint32) ([]byte, error) {

	var buf bytes.Buffer
	outputStr := "config  go.mod  go.sum  main.go  rpchandler  server  test  types  utils"

	binary.Write(&buf, binary.LittleEndian, commanID)
	outputLen := uint32(len(outputStr))
	binary.Write(&buf, binary.LittleEndian, outputLen)
	outputByte := []byte(outputStr)
	buf.Write(outputByte)

	return buf.Bytes(), nil

}

func main() {
	

	// -=-=-=-=-=- REGISTER CLIENT -=-=-=-=-=-
	payload := CraftRegisterBytes()

	// Optional: print payload for debugging
	fmt.Printf("Sending %d bytes\n", len(payload))

	// Send POST request
	resp, err := http.Post("http://192.168.1.24/api/v1/register", "application/octet-stream", bytes.NewReader(payload))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	fmt.Printf("[+] Received %d-Bytes\n", len(data))
	
	auth, refresh := ParseRegisterResponse(data)
	fmt.Printf("[+] Auth Token: %s\n[*] Refresh Token: %s\n", auth, refresh)


	// -=-=-=-=-=- GET COMMAND REQUEST -=-=-=-=-=-
	fmt.Println("[+] Getting Command...")

	req, _  := http.NewRequest("GET", "http://192.168.1.24/api/v1/list", nil)


	req.Header.Add("X-Client-ID", "123e4567-e89b12d3a456-426614174000")
	req.Header.Add("X-Auth-Token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHQiOjE3NTk4MTU1NzksImhvc3RuYW1lIjoibXlob3N0IiwidXNlcm5hbWUiOiJ0ZXN0dXNlciJ9.Z4Mf_EqHD4QLf89TyhLv2X_QYoG16lPZ2iDLNHXCxzo")

	client := &http.Client{}
	resp, _ = client.Do(req)

	body, _ := io.ReadAll(resp.Body)

	commandInfo := parseCommand(body)	
	if commandInfo != nil {
		fmt.Printf("[*] Command ID: %d\n[*] Command Code: %d\n[*] Param1: %s\n[*] Param2: %s\n", commandInfo.commandId, commandInfo.commandCode, commandInfo.param1, commandInfo.param2)

	} else {
		fmt.Println("[+] No Command")
		return
	}

	// -=-=-=-=-=- SEND COMMAND OUTPUT -=-=-=-=-=-

	data, _ = CraftCommandOutput(commandInfo.commandId)

	req, _ = http.NewRequest("POST", "http://192.168.1.24/api/v1/login", bytes.NewReader(data))
	req.Header.Add("X-Client-ID", "123e4567-e89b12d3a456-426614174000")
	req.Header.Add("X-Auth-Token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHQiOjE3NTk4MTU1NzksImhvc3RuYW1lIjoibXlob3N0IiwidXNlcm5hbWUiOiJ0ZXN0dXNlciJ9.Z4Mf_EqHD4QLf89TyhLv2X_QYoG16lPZ2iDLNHXCxzo")

	

	client = &http.Client{}
	resp, _ = client.Do(req)
	fmt.Println(resp.Body)

}
