package runner

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	gosync "sync"
)

var dangerousPatterns = []string{
	"rm -rf /", "rm -rf ~", "mkfs", "dd if=", "> /dev/",
	"chmod -R 777",
}

func warnIfDangerous(cmdStr string) {
	lower := strings.ToLower(cmdStr)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lower, pattern) {
			fmt.Fprintf(os.Stderr, "Warning: potentially dangerous pattern '%s' in command: %s\n", pattern, cmdStr)
			break
		}
	}

	if (strings.Contains(lower, "curl") || strings.Contains(lower, "wget")) && strings.Contains(lower, "|") {
		fmt.Fprintf(os.Stderr, "Warning: piped download detected, review carefully: %s\n", cmdStr)
	}
}

func getShell() (string, string) {
	if runtime.GOOS == "windows" {
		return "cmd", "/c"
	}
	return "sh", "-c"
}

// BackgroundGroup tracks background commands for lifecycle management
var BackgroundGroup gosync.WaitGroup

// RunCommands executes a list of commands in the specified directory
func RunCommands(dir string, commands []string, background bool) error {
	shell, flag := getShell()

	for _, cmdStr := range commands {
		if strings.TrimSpace(cmdStr) == "" {
			continue
		}

		warnIfDangerous(cmdStr)

		cmd := exec.Command(shell, flag, cmdStr)
		cmd.Dir = dir

		if background {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("failed to start command '%s': %w", cmdStr, err)
			}
			BackgroundGroup.Add(1)
			go func(c *exec.Cmd, cmdStr string) {
				defer BackgroundGroup.Done()
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

// WaitBackground blocks until all background commands complete
func WaitBackground() {
	BackgroundGroup.Wait()
}
