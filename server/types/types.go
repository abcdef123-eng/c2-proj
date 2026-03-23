package types


type CommandReply struct {
	CommandID uint32
	CommandCode uint32
	Param1 string
	Param2 string
}

type ClientInfo struct {
        Guid string
        Username string
        Hostname string
        PID uint32
        Arch uint32
        Version string
}

