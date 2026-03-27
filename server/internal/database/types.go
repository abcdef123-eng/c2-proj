package database

var (
	verbs = []string{
		"JUMP", "RUN", "WALK", "FLY", "CHASE", "CATCH", "DREAM", "BUILD", "GROW",
		"SWIM", "DRIVE", "RIDE", "SEEK", "DISCOVER", "SHINE", "IGNITE", "TRANSFORM",
		"EXPLORE", "CLIMB", "LEAP", "STRIKE", "SURGE", "DRIFT", "HUNT", "FORGE",
		"BREACH", "PIERCE", "VANISH", "HAUNT", "CARVE", "SMASH", "STALK", "DEVOUR",
		"SCATTER", "LURK", "RECLAIM", "ASCEND", "DESCEND", "INTERCEPT", "SHATTER",
		"BURN", "FREEZE", "CRASH", "SLICE", "ROAM", "PLUNGE", "CONQUER", "EVADE",
		"STORM", "ECHO",
	}

	nouns = []string{
		"WOLF", "EAGLE", "MOUNTAIN", "RIVER", "STAR", "FIRE", "LIGHT", "HEART",
		"BREEZE", "NIGHT", "VISION", "CLOUD", "STORM", "FLAME", "EARTH", "OCEAN",
		"SOUL", "THUNDER", "HORIZON", "PHANTOM", "RAVEN", "CIPHER", "VIPER", "COBRA",
		"FALCON", "SPECTER", "WRAITH", "TITAN", "HYDRA", "DRAGON", "BLADE", "ARROW",
		"GHOST", "SHADOW", "IRON", "STEEL", "FROST", "ASH", "EMBER", "VOID",
		"SENTINEL", "APEX", "NEXUS", "VECTOR", "PULSE", "SIGNAL", "DAGGER", "CRYPT",
		"ORACLE", "ABYSS",
	}
)

type ClientData struct {
	Guid        string
	CodeName    string
	Username    string
	Hostname    string
	Ip          string
	Arch        string
	Pid         int32
	Version     string
	LastCheckin string
}
