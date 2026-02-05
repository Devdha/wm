package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Spinner provides a simple loading animation
type Spinner struct {
	frames   []string
	message  string
	delay    time.Duration
	writer   io.Writer
	stopChan chan struct{}
	mu       sync.Mutex
	active   bool
}

// NewSpinner creates a spinner with dots animation
func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message:  message,
		delay:    100 * time.Millisecond,
		writer:   os.Stdout,
		stopChan: make(chan struct{}),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() *Spinner {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return s
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		i := 0
		for {
			select {
			case <-s.stopChan:
				return
			default:
				s.mu.Lock()
				frame := s.frames[i%len(s.frames)]
				fmt.Fprintf(s.writer, "\r%s %s", Info.Sprint(frame), s.message)
				s.mu.Unlock()
				time.Sleep(s.delay)
				i++
			}
		}
	}()

	return s
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()

	close(s.stopChan)
	time.Sleep(s.delay) // Wait for last frame

	// Clear the line
	fmt.Fprint(s.writer, "\r\033[K")
}

// Success stops the spinner and prints a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	Success.Print(IconCheck + " ")
	fmt.Println(message)
}

// Fail stops the spinner and prints an error message
func (s *Spinner) Fail(message string) {
	s.Stop()
	Error.Print(IconCross + " ")
	fmt.Println(message)
}

// Update changes the spinner message
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}
