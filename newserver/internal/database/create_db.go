package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

func GetDB() (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		homeDir, e := os.UserHomeDir()
		if e != nil {
			err = e
			return
		}
		dbPath := filepath.Join(homeDir, ".scurrier", "database", "scurrier.db")
		db, err = sql.Open("sqlite", dbPath)
	})
	return db, err
}

func CreateScurrierDir(dirpath string) error {
	if err := os.MkdirAll(dirpath, 0755); err != nil {
		return err
	}

	subDirs := []string{"database", "logs", "config"}
	for _, dir := range subDirs {
		path := filepath.Join(dirpath, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		}
	}

	if err := createDefaultConfig(filepath.Join(dirpath, "config", "config.json")); err != nil {
		return err
	}

	if err := createDatabase(filepath.Join(dirpath, "database", "scurrier.db")); err != nil {
		return err
	}

	return nil
}

func ValidateSetup(dirpath string) error {
	checks := []struct {
		path    string
		isDir   bool
		createFn func(string) error
	}{
		{filepath.Join(dirpath, "logs"), true, func(p string) error { return os.MkdirAll(p, 0755) }},
		{filepath.Join(dirpath, "config", "config.json"), false, createDefaultConfig},
		{filepath.Join(dirpath, "database", "scurrier.db"), false, createDatabase},
	}

	for _, c := range checks {
		if _, err := os.Stat(c.path); os.IsNotExist(err) {
			if err := c.createFn(c.path); err != nil {
				return err
			}
		}
	}

	return nil
}

func createDefaultConfig(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}
	defaultConfig := `{
	"host": "0.0.0.0",
	"port": 8080,
	"grpc_port": 50051,
	"getEndpoint": "/api/get",
	"postEndpoint": "/api/post"
}`
	return os.WriteFile(path, []byte(defaultConfig), 0644)
}

func createDatabase(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer db.Close()

	clientTable := `CREATE TABLE IF NOT EXISTS clients(
		guid         TEXT NOT NULL,
		code_name    TEXT NOT NULL,
		username     TEXT NOT NULL,
		hostname     TEXT NOT NULL,
		ip           TEXT NOT NULL,
		arch         INT NOT NULL,
		pid          INT NOT NULL,
		version      TEXT NOT NULL,
		last_checkin TEXT NOT NULL
	)`

	commandsTable := `CREATE TABLE IF NOT EXISTS commands(
		guid       TEXT NOT NULL,
		code       INT NOT NULL,
		param1     TEXT NOT NULL,
		param2     TEXT NOT NULL,
		command_id INT NOT NULL,
		executed   INT NOT NULL,
		tasked_at  TEXT NOT NULL
	)`

	if _, err := db.Exec(clientTable); err != nil {
		return err
	}
	if _, err := db.Exec(commandsTable); err != nil {
		return err
	}

	return nil
}

func CheckAndSetup() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	scurrierDir := filepath.Join(homeDir, ".scurrier")

	if _, err := os.Stat(scurrierDir); os.IsNotExist(err) {
		return CreateScurrierDir(scurrierDir)
	}

	return ValidateSetup(scurrierDir)
}
