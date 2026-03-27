package db

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"

	_ "modernc.org/sqlite"
)

var (
	database *sql.DB
	dbOnce   sync.Once
)

func GetDB() (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		homeDir, e := os.UserHomeDir()
		if e != nil {
			err = e
			return
		}

		dbPath := filepath.Join(homeDir, ".scurrier", "client_history.db")
		if e := os.MkdirAll(filepath.Dir(dbPath), 0755); e != nil {
			err = e
			return
		}

		database, err = sql.Open("sqlite", dbPath)
		if err != nil {
			return
		}

		_, err = database.Exec(`CREATE TABLE IF NOT EXISTS task_history (
			task_id   INT  PRIMARY KEY,
			guid      TEXT NOT NULL,
			command   TEXT NOT NULL,
			param1    TEXT NOT NULL,
			param2    TEXT NOT NULL,
			tasked_at TEXT NOT NULL,
			executed  INT  NOT NULL DEFAULT 0
		)`)
	})
	return database, err
}

func InsertTask(taskID int32, guid, command, param1, param2 string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	taskedAt := time.Now().Unix()
	_, err = db.Exec(
		`INSERT OR IGNORE INTO task_history (task_id, guid, command, param1, param2, tasked_at, executed) VALUES (?, ?, ?, ?, ?, ?, 0)`,
		taskID, guid, command, param1, param2, taskedAt,
	)
	return err
}

func MarkExecuted(taskID int32) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE task_history SET executed = 1 WHERE task_id = ?`, taskID)
	return err
}

func ListHistory(out io.Writer, onlyNotExecuted bool) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	query := `SELECT task_id, guid, command, param1, param2, tasked_at, executed FROM task_history ORDER BY tasked_at DESC`
	if onlyNotExecuted {
		query = `SELECT task_id, guid, command, param1,param2, tasked_at, executed FROM task_history WHERE executed = 0 ORDER BY tasked_at DESC`
	}

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.Debug)
	fmt.Fprintln(out)
	fmt.Fprintln(w, "Task ID\tGUID\tCommand\tParam1\tParam2\tTasked At\tExecuted")
	fmt.Fprintln(w, "-------\t----\t-------\t------\t------\t---------\t--------")

	for rows.Next() {
		var taskID int32
		var guid, command, param1, param2, taskedAt string
		var executed int
		if err := rows.Scan(&taskID, &guid, &command, &param1, &param2, &taskedAt, &executed); err != nil {
			continue
		}

		ts, _ := strconv.ParseInt(taskedAt, 10, 64)
		t := time.Unix(ts, 0).Format("2006-01-02 15:04:05")

		status := "no"
		if executed == 1 {
			status = "yes"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n", taskID, guid, command, param1, param2, t, status)
	}
	w.Flush()
	fmt.Fprintln(out)
	return nil
}
