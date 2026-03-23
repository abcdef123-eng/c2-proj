package client

import (
	"fmt"
	"io"
	"time"

	"github.com/abcdef123-eng/c2-proj/newClient/internal/commander"
	"github.com/peterh/liner"
)

func RunClient() {

	line := liner.NewLiner()
	defer line.Close()

	for {
		input, err := line.Prompt(fmt.Sprintf("[%s] $> ", time.Now().Format("15:04:05")))
		if err == liner.ErrPromptAborted || err == io.EOF {
			fmt.Println("\n[*] Exiting...")
			return
		} else if err != nil {
			fmt.Println("[!] Failed reading line:", err)
			return
		}

		line.AppendHistory(input)

		cmd, err := commander.Parse(input)
		if err != nil {
			fmt.Println("[!]", err)
			continue
		}

		commander.Dispatch(cmd)
	}

}
