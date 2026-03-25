package database

import (
	"bytes"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	bytehandler "github.com/execute-assembly/c2-proj/newserver/internal/bytes"

	_ "modernc.org/sqlite"
)

func gen_string() string {
	str1 := nouns[rand.Intn(len(nouns))]
	str2 := verbs[rand.Intn(len(verbs))]
	return fmt.Sprintf("%s_%s", str1, str2)
}

var (
	verbs = []string{"jump", "run", "walk", "fly", "chase", "catch", "dream", "build", "grow",
		"swim", "drive", "ride", "seek", "discover", "shine", "ignite", "transform",
		"explore", "climb", "leap",
	}

	nouns = []string{
		"wolf", "eagle", "mountain", "river", "dream", "star", "fire", "light",
		"heart", "breeze", "night", "vision", "cloud", "storm", "flame", "earth",
		"ocean", "soul", "thunder", "horizon",
	}
)

func RegisterClient(data *bytes.Reader, IpAddress string) (string, error) {
	db, err := GetDB()
	if err != nil {
		return "", err
	}
	clientData, err := bytehandler.ParseClientRegister(data, IpAddress)
	if err != nil {
		return "", err
	}

	fmt.Printf("Guid: %s\n", clientData.Guid)
	fmt.Printf("Username: %s\n", clientData.Username)
	fmt.Printf("Hostname: %s\n", clientData.Hostname)
	fmt.Printf("IP: %s\n", clientData.Ip)
	fmt.Printf("Arch: %d\n", clientData.Arch)
	fmt.Printf("WinVersion: %s\n", clientData.WinVersion)
	fmt.Printf("PID: %d\n", clientData.Pid)
	query := `INSERT INTO clients(guid, code_name, username, hostname, ip, pid, arch, version, last_checkin) VALUES(?,?,?,?,?,?,?,?,?)`
	slog.Info("New User Registered", "username", clientData.Username, "hostname", clientData.Hostname, "code_name", clientData.Hostname)
	code_name := gen_string()

	_, err = db.Exec(query, clientData.Guid, code_name, clientData.Username, clientData.Hostname, clientData.Ip, clientData.Pid, clientData.Arch, clientData.WinVersion, time.Now().Unix())
	if err != nil {
		return "", err
	}

	return clientData.Guid, nil

}
