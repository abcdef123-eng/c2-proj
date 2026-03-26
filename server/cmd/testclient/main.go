package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const COMMAND_REGISTER_CLIENT = 50

var CommandNames = map[int32]string{
	0x1: "ls",
	0x2: "cd",
	0x3: "rm",
	0x4: "mv",
	0x5: "cat",
	0x6: "get-privs",
}

var CommandParamCount = map[int32]int{
	0x1: 1,
	0x2: 1,
	0x3: 1,
	0x4: 2,
	0x5: 1,
	0x6: 0,
}

func writeString(buf *bytes.Buffer, s string) {
	length := uint32(len(s))
	binary.Write(buf, binary.LittleEndian, length)
	buf.WriteString(s)
}

func readString(r *bytes.Reader) (string, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func register() string {
	guid := "1111-2222-3333-0000"
	username := "testuser"
	hostname := "DESKTOP-TEST"
	arch := byte(0x02)
	major := uint16(10)
	minor := uint16(0)
	build := uint16(22621)
	pid := uint32(1234)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(COMMAND_REGISTER_CLIENT))
	writeString(buf, guid)
	writeString(buf, username)
	writeString(buf, hostname)
	buf.WriteByte(arch)
	binary.Write(buf, binary.LittleEndian, major)
	binary.Write(buf, binary.LittleEndian, minor)
	binary.Write(buf, binary.LittleEndian, build)
	binary.Write(buf, binary.LittleEndian, pid)

	req, err := http.NewRequest("POST", "http://localhost:8080/api/post", buf)
	if err != nil {
		fmt.Println("failed to create register request:", err)
		os.Exit(1)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("failed to send register request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Println("[register] status:", resp.Status)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("[register] unexpected status:", resp.Status)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read register response:", err)
		os.Exit(1)
	}

	r := bytes.NewReader(body)
	var jwtLen uint32
	binary.Read(r, binary.LittleEndian, &jwtLen)
	jwtData := make([]byte, jwtLen)
	io.ReadFull(r, jwtData)
	jwt := string(jwtData)
	fmt.Printf("[register] JWT (%d bytes): %s\n", jwtLen, jwt)
	return jwt
}

func pollTasks(jwt string) {
	req, err := http.NewRequest("GET", "http://localhost:8080/api/get", nil)
	if err != nil {
		fmt.Println("failed to create task request:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("failed to send task request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Println("[tasks] no tasks queued")
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("[tasks] unexpected status:", resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read task response:", err)
		return
	}

	if len(body) == 0 {
		fmt.Println("[tasks] no tasks queued")
		return
	}

	r := bytes.NewReader(body)
	taskNum := 0
	for r.Len() > 0 {
		taskNum++
		var taskID int32
		var commandCode int32
		if err := binary.Read(r, binary.LittleEndian, &taskID); err != nil {
			fmt.Println("failed to read task ID:", err)
			return
		}
		if err := binary.Read(r, binary.LittleEndian, &commandCode); err != nil {
			fmt.Println("failed to read command code:", err)
			return
		}

		name, ok := CommandNames[commandCode]
		if !ok {
			name = fmt.Sprintf("unknown(0x%x)", commandCode)
		}

		fmt.Printf("[task %d] ID=%d command=%s", taskNum, taskID, name)

		paramCount := CommandParamCount[commandCode]
		if paramCount >= 1 {
			param1, err := readString(r)
			if err != nil {
				fmt.Println("\nfailed to read param1:", err)
				return
			}
			fmt.Printf(" param1=%q", param1)
		}
		if paramCount == 2 {
			param2, err := readString(r)
			if err != nil {
				fmt.Println("\nfailed to read param2:", err)
				return
			}
			fmt.Printf(" param2=%q", param2)
		}
		fmt.Println()
	}
}

func main() {
	jwt := register()
	fmt.Println("Polling for tasks every 3s (Ctrl+C to stop)...")
	for {
		pollTasks(jwt)
		time.Sleep(3 * time.Second)
	}
}
