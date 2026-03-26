package commander

import "fmt"

const (
	colorRed    = "\033[1;31m"
	colorOrange = "\033[1;33m"
	colorGreen  = "\033[1;32m"
	colorWhite  = "\033[1;37m"
	colorReset  = "\033[0m"
)

func PrintErr(msg string) {
	fmt.Printf("%s[!]%s %s\n", colorRed, colorReset, msg)
}

func PrintInfo(msg string) {
	fmt.Printf("%s[+]%s %s\n", colorOrange, colorReset, msg)
}

func PrintOk(msg string) {
	fmt.Printf("%s[*]%s %s\n", colorGreen, colorReset, msg)
}

func BoldWhite(s string) string {
	return fmt.Sprintf("%s%s%s", colorWhite, s, colorReset)
}
