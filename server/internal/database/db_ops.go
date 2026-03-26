package database

import (
	"bytes"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	bytehandler "github.com/execute-assembly/c2-proj/server/internal/bytes"

	_ "modernc.org/sqlite"
)

func genCodename() string {
	str1 := nouns[rand.Intn(len(nouns))]
	str2 := verbs[rand.Intn(len(verbs))]
	return fmt.Sprintf("%s_%s", str1, str2)
}

func ArchIntToString(arch byte) string {
	switch arch {
	case 0x1:
		return "x86"
	case 0x2:
		return "x64"
	case 0x3:
		return "ARM"
	default:
		return "UNKNOWN"
	}
}

func RegisterClient(data *bytes.Reader, IpAddress string) (string, string, error) {
	db, err := GetDB()
	if err != nil {
		return "", "", err
	}
	clientData, err := bytehandler.ParseClientRegister(data, IpAddress)
	if err != nil {
		return "", "", err
	}

	codeName := genCodename()
	query := `INSERT INTO clients(guid, code_name, username, hostname, ip, pid, arch, version, last_checkin) VALUES(?,?,?,?,?,?,?,?,?)`

	_, err = db.Exec(query, clientData.Guid, codeName, clientData.Username, clientData.Hostname, clientData.Ip, clientData.Pid, ArchIntToString(clientData.Arch), clientData.WinVersion, time.Now().Unix())
	if err != nil {
		return "", "", err
	}
	slog.Info("New User Registered", "username", clientData.Username, "hostname", clientData.Hostname, "code_name", codeName)

	return clientData.Guid, codeName, nil
}

func CheckIfUserExists_db(guid string) (bool, error) {
	db, err := GetDB()
	if err != nil {
		return false, err
	}
	query := `SELECT EXISTS(SELECT 1 FROM clients WHERE guid = ?)`
	var exists bool
	err = db.QueryRow(query, guid).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func ListClients_db() ([]ClientData, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT * FROM clients")
	if err != nil {
		return nil, err
	}

	var clients []ClientData

	for rows.Next() {
		var c ClientData
		err := rows.Scan(&c.Guid, &c.CodeName, &c.Username, &c.Hostname, &c.Ip, &c.Arch, &c.Pid, &c.Version, &c.LastCheckin)
		if err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return clients, nil
}

func ConvertCodeName_db(codename string) (string, error) {
	db, err := GetDB()
	if err != nil {
		return "", err
	}

	var guid string
	err = db.QueryRow("SELECT guid FROM clients WHERE code_name = ?", codename).Scan(&guid)
	if err != nil {
		return "", err
	}
	return guid, nil
}

func UpdateLastSeen_db(guid string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE clients SET last_checkin = ? WHERE guid = ?", time.Now().Unix(), guid)
	if err != nil {
		return err
	}
	return nil
}
