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

func (c *Console) Print(message string) {
	fmt.Println(message)
}

func (c *Console) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Silent implements Prompter with auto-confirm (for -f flag)
type Silent struct {
	autoConfirm bool
}

func NewSilent(autoConfirm bool) *Silent {
	return &Silent{autoConfirm: autoConfirm}
}

func (s *Silent) Confirm(message string) bool {
	return s.autoConfirm
}

func (s *Silent) Print(message string) {
	fmt.Println(message)
}

func (s *Silent) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
