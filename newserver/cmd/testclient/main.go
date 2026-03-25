package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"
)

const COMMAND_REGISTER_CLIENT = 50

func writeString(buf *bytes.Buffer, s string) {
	length := uint32(len(s))
	binary.Write(buf, binary.LittleEndian, length)
	buf.WriteString(s)
}

func main() {
	guid := "1111-2222-3333-4444"
	username := "testuser"
	hostname := "DESKTOP-TEST"
	arch := byte(0x02)   // x64
	major := uint16(10)  // Windows 10/11
	minor := uint16(0)
	build := uint16(22621) // Windows 11 22H2
	pid := uint32(1234)

	buf := new(bytes.Buffer)

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
		fmt.Println("failed to create request:", err)
		os.Exit(1)
	}

	req.Header.Set("X-Type", fmt.Sprintf("%d", COMMAND_REGISTER_CLIENT))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("failed to send request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Println("response status:", resp.Status)
}
