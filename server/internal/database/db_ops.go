package database

import (
	"bytes"
	"database/sql"
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

func ArchIntToString(arch byte) string {
	var ArchStr string
	switch arch {
	case 0x1:
		ArchStr = "x86"
	case 0x2:
		ArchStr = "x64"
	case 0x3:
		ArchStr = "ARM"
	default:
		ArchStr = "UNKNOWN"
	}
	return ArchStr
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

	ArchStr := ArchIntToString(clientData.Arch)
	query := `INSERT INTO clients(guid, code_name, username, hostname, ip, pid, arch, version, last_checkin) VALUES(?,?,?,?,?,?,?,?,?)`
	code_name := gen_string()

	_, err = db.Exec(query, clientData.Guid, code_name, clientData.Username, clientData.Hostname, clientData.Ip, clientData.Pid, ArchStr, clientData.WinVersion, time.Now().Unix())
	if err != nil {
		return "", "", err
	}
	slog.Info("New User Registered", "username", clientData.Username, "hostname", clientData.Hostname, "code_name", code_name)

	return clientData.Guid, code_name, nil

}

func CheckIfUserExists_db(Guid string) (bool, error) {
	db, err := GetDB()
	if err != nil {
		return false, err
	}
	query := `SELECT EXISTS(SELECT 1 FROM clients WHERE guid = ?)`
	var exists bool
	err = db.QueryRow(query, Guid).Scan(&exists)
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

	var Clients []ClientData

	for rows.Next() {
		var c ClientData
		err := rows.Scan(&c.Guid, &c.Code_name, &c.Username, &c.Hostname, &c.Ip, &c.Arch, &c.Pid, &c.Version, &c.Last_checkin)
		if err != nil {
			return nil, err
		}
		Clients = append(Clients, c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return Clients, nil
}

func ConvertCodeName_db(Codename string) (string, error) {
	db, err := GetDB()
	if err != nil {
		return "", err
	}

	var Guid string
	err = db.QueryRow("SELECT guid FROM clients WHERE code_name = ?", Codename).Scan(&Guid)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", err
		}
		return "", err
	}
	return Guid, nil

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
