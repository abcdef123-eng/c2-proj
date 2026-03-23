package commander

import (
	"fmt"
	"strings"

	"github.com/google/shlex"
)

type CommandData struct {
	Name string
	Args []string
}

func Parse(args string) (CommandData, error) {
	CommandArgs, err := shlex.Split(args)
	if err != nil {
		return CommandData{}, err
	}

	if len(CommandArgs) == 0 {
		return CommandData{}, fmt.Errorf("Not Enough args")
	}

	return CommandData{
		Name: strings.ToLower(CommandArgs[0]),
		Args: CommandArgs[1:],
	}, nil
}

func Dispatch(Commands CommandData) error {
	switch Commands.Name {
	case "ls":
		fmt.Println("cmd -> ls")
		break
	default:
		fmt.Println("[!] Incorrect Command!")
		break
	}

	return nil
}
