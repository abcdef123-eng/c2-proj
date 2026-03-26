package database

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

type ClientData struct {
	Guid         string
	Code_name    string
	Username     string
	Hostname     string
	Ip           string
	Arch         string
	Pid          int32
	Version      string
	Last_checkin string
}
