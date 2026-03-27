package commander

import (
	"fmt"
	"io"
	"os"
)

var Out io.Writer = os.Stdout

const (
	colorRed    = "\033[1;31m"
	colorOrange = "\033[1;33m"
	colorGreen  = "\033[1;32m"
	colorWhite  = "\033[1;37m"
	colorBlue   = "\033[1;34m"
	colorReset  = "\033[0m"
)

const (
	EVENT_NEW_AGENT      = 1
	EVENT_COMMAND_OUTPUT = 2
)

func PrintErr(msg string) {
	fmt.Fprintf(Out, "%s[!]%s %s\n", colorRed, colorReset, msg)
}

func PrintInfo(msg string) {
	fmt.Fprintf(Out, "%s[+]%s %s\n", colorOrange, colorReset, msg)
}

func PrintOk(msg string) {
	fmt.Fprintf(Out, "%s[*]%s %s\n", colorGreen, colorReset, msg)
}

func BoldWhite(s string) string {
	return fmt.Sprintf("%s%s%s", colorWhite, s, colorReset)
}

func Blue(s string) string {
	return fmt.Sprintf("%s%s%s", colorBlue, s, colorReset)
}

func PrintOutput(taskID int32, guid string, output string) {
	sep := "――――――――――――――――――――――――――――――――――――――――――"
	fmt.Fprintf(Out, "%s[+]%s Task %d output from %s\n", colorOrange, colorReset, taskID, guid)
	fmt.Fprintln(Out, sep)
	fmt.Fprintln(Out, output)
	fmt.Fprintln(Out, sep)
}
