package ssh

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// AskPassword asks password using the terminal
func AskPassword(prompt string) string {
	fmt.Print(prompt)
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	return string(bytePassword)
}

// Ask something with prompt
func Ask(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.Replace(text, "\n", "", -1)
}
