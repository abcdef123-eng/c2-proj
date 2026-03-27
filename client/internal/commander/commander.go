package commander

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/shlex"
)

type CommandData struct {
	Name string
	Args []string
}

func Parse(args string) (CommandData, error) {
	commandArgs, err := shlex.Split(args)
	if err != nil {
		return CommandData{}, err
	}

	if len(commandArgs) == 0 {
		return CommandData{}, fmt.Errorf("Not Enough args")
	}

	return CommandData{
		Name: strings.ToLower(commandArgs[0]),
		Args: commandArgs[1:],
	}, nil
}

func Dispatch(cmd CommandData) error {
	switch cmd.Name {
	case "ls":
		HandleLS(cmd.Args)
	case "list":
		HandleListClients()
	case "use":
		UseClient(cmd.Args)
	case "history":
		HandleHistory()
	case "tasks":
		HandleTasks()
	case "back":
		ClientInUse = ""
		ClientCodeName = ""
	case "exit":
		os.Exit(0)
	default:
		PrintErr("Incorrect Command!")
	}

	return nil
}
