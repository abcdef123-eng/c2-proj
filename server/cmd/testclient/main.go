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

const (
	COMMAND_REGISTER_CLIENT = 50
	COMMAND_POST_OUTPUT     = 52
	COMMAND_GET_TASK        = 51
)

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

type Task struct {
	TaskID      int32
	CommandCode int32
	Param1      string
}

func writeString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.LittleEndian, uint32(len(s)))
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

func pollTasks(jwt string) []Task {
	req, err := http.NewRequest("GET", "http://localhost:8080/api/get", nil)
	if err != nil {
		fmt.Println("failed to create task request:", err)
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	q := req.URL.Query()
	q.Set("id", fmt.Sprintf("%d", COMMAND_GET_TASK))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("failed to send task request:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Println("[tasks] no tasks queued")
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("[tasks] unexpected status:", resp.Status)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read task response:", err)
		return nil
	}

	var tasks []Task
	r := bytes.NewReader(body)
	taskNum := 0
	for r.Len() > 0 {
		taskNum++
		var taskID, commandCode int32
		if err := binary.Read(r, binary.LittleEndian, &taskID); err != nil {
			fmt.Println("failed to read task ID:", err)
			return tasks
		}
		if err := binary.Read(r, binary.LittleEndian, &commandCode); err != nil {
			fmt.Println("failed to read command code:", err)
			return tasks
		}

		name, ok := CommandNames[commandCode]
		if !ok {
			name = fmt.Sprintf("unknown(0x%x)", commandCode)
		}

		task := Task{TaskID: taskID, CommandCode: commandCode}
		fmt.Printf("[task %d] ID=%d command=%s", taskNum, taskID, name)

		paramCount := CommandParamCount[commandCode]
		if paramCount >= 1 {
			param1, err := readString(r)
			if err != nil {
				fmt.Println("\nfailed to read param1:", err)
				return tasks
			}
			task.Param1 = param1
			fmt.Printf(" param1=%q", param1)
		}
		if paramCount == 2 {
			if _, err := readString(r); err != nil {
				fmt.Println("\nfailed to read param2:", err)
				return tasks
			}
		}
		fmt.Println()
		tasks = append(tasks, task)
	}
	return tasks
}

func sendOutput(jwt string, tasks []Task) {
	if len(tasks) == 0 {
		return
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(COMMAND_POST_OUTPUT))
	for _, t := range tasks {
		output := fmt.Sprint("File1.txt\nfile2.txt\nPasswords/\nwhoami.txt\n)")
		binary.Write(buf, binary.LittleEndian, t.TaskID)
		writeString(buf, output)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/api/post", buf)
	if err != nil {
		fmt.Println("failed to create output request:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("failed to send output:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("[output] sent %d task outputs, status: %s\n", len(tasks), resp.Status)
}

func main() {
	var jwt string
	if len(os.Args) > 1 {
		jwt = os.Args[1]
		fmt.Println("[*] using provided JWT, skipping registration")
	} else {
		jwt = register()
	}

	fmt.Println("Polling for tasks every 3s (Ctrl+C to stop)...")
	for {
		tasks := pollTasks(jwt)
		if len(tasks) > 0 {
			sendOutput(jwt, tasks)
		}
		time.Sleep(3 * time.Second)
	}
}
