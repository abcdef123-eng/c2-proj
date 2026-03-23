package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"strconv"
	"time"
)


func formatLastSeen(lastSeenStr string) string {
	lastSeenUnix, _ := strconv.ParseInt(lastSeenStr, 10, 64)
	now := time.Now().UTC()
	lastSeen := time.Unix(lastSeenUnix, 0)
	diff := now.Sub(lastSeen)

	switch {
	case diff < time.Minute:
		return fmt.Sprintf("Last Seen: %ds ago", int(diff.Seconds()))
	case diff < time.Hour:
		return fmt.Sprintf("Last Seen: %dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("Last Seen: %dh ago", int(diff.Hours()))
	default:
		days := int(diff.Hours()) / 24
		return fmt.Sprintf("Last Seen: %dd ago", days)
	}
}





func ListTasks(guid string) {
	db, err := sql.Open("sqlite3", "../rat.db");
	if err != nil {
		fmt.Println(err);
		return;
	}
	defer db.Close()
	var dbTask, dbParam, dbCommandID string
	rows,_  := db.Query("SELECT code, param, command_id FROM commands WHERE guid=?", guid)
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&dbTask, &dbParam, &dbCommandID);
		fmt.Printf("Command Code: %s | Parameter: %s | CommandID: %s\n", dbTask, dbParam, dbCommandID);
	}
}


// handles retrievinf the clients ID from their code_name
// func code_name_to_guid(name string) error {
// 	db, err := sql.Open("sqlite3", "../rat.db");
// 	if err != nil {
// 		fmt.Println("[!] Failed opeing Database");
// 		return err;
// 	}

// 	defer db.Close();



// 	err = db.QueryRow("SELECT guid FROM clients WHERE code_name=?", name).Scan(&client_in_use);
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			fmt.Println("[!] Client Doesnt Exists");
// 			return err;
// 		}
// 	}
// 	fmt.Printf("[+] using: %s\n", name);

// 	return nil;

// }