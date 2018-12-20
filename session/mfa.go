package session

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// AskMFAWithPrompt ask for a mfa token using a custom prompt
func AskMFAWithPrompt(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSuffix(text, "\n")
}

// AskMFA ask for a mfa token using a default prompt
func AskMFA() string {
	return AskMFAWithPrompt("Enter MFA:")
}
