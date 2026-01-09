package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Prompter handles user interaction
type Prompter interface {
	Confirm(message string) bool
	Input(prompt, defaultValue string) string
	Print(message string)
	Printf(format string, args ...interface{})
}

// Console implements Prompter for terminal interaction
type Console struct {
	reader *bufio.Reader
}

// NewConsole creates a new Console prompter
func NewConsole() *Console {
	return &Console{reader: bufio.NewReader(os.Stdin)}
}

func (c *Console) Confirm(message string) bool {
	fmt.Print(message + " [y/N]: ")
	answer, _ := c.reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

func (c *Console) Input(prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	answer, _ := c.reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return defaultValue
	}
	return answer
}

func (c *Console) Print(message string) {
	fmt.Println(message)
}

func (c *Console) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Silent implements Prompter with auto-confirm (for -f flag or tests)
type Silent struct {
	autoConfirm bool
	defaults    map[string]string
}

func NewSilent(autoConfirm bool) *Silent {
	return &Silent{autoConfirm: autoConfirm, defaults: make(map[string]string)}
}

func (s *Silent) Confirm(message string) bool {
	return s.autoConfirm
}

func (s *Silent) Input(prompt, defaultValue string) string {
	if val, ok := s.defaults[prompt]; ok {
		return val
	}
	return defaultValue
}

func (s *Silent) Print(message string) {
	fmt.Println(message)
}

func (s *Silent) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
