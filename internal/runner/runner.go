package runner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCommands executes a list of commands in the specified directory
func RunCommands(dir string, commands []string, background bool) error {
	for _, cmdStr := range commands {
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = dir

		if background {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("failed to start command '%s': %w", cmdStr, err)
			}
			// Wait for completion in goroutine, print errors
			go func(c *exec.Cmd, cmdStr string) {
				if err := c.Wait(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: command '%s' failed: %v\n", cmdStr, err)
				}
			}(cmd, cmdStr)
		} else {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("command '%s' failed: %w", cmdStr, err)
			}
		}
	}

	return nil
}
