package client

import (
	"fmt"
	"io"
	"time"

	"github.com/chzyer/readline"
	"github.com/execute-assembly/c2-proj/client/internal/commander"
)

var RL *readline.Instance

func Init() error {
	var err error
	RL, err = readline.New("")
	return err
}

func RunClient() {
	defer RL.Close()

	for {
		prompt := fmt.Sprintf("[%s] ", time.Now().Format("15:04:05"))
		if commander.ClientCodeName != "" {
			prompt += commander.Blue(commander.ClientCodeName) + " "
		}
		prompt += "$> "
		RL.SetPrompt(prompt)
		input, err := RL.Readline()
		if err == readline.ErrInterrupt || err == io.EOF {
			fmt.Fprintln(RL.Stdout(), "\n[*] Exiting...")
			return
		} else if err != nil {
			fmt.Fprintln(RL.Stdout(), "[!] Failed reading line:", err)
			return
		}

		if input == "" {
			continue
		}

		cmd, err := commander.Parse(input)
		if err != nil {
			fmt.Fprintln(RL.Stdout(), "[!]", err)
			continue
		}

		commander.Dispatch(cmd)
	}
}
