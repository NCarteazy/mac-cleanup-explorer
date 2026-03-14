package executor

import (
	"fmt"
	"os/exec"
	"strings"
)

// Command represents a parsed cleanup command.
type Command struct {
	Raw       string
	Validated bool
	Error     string
	Warning   string
	Output    string
	ExitCode  int
	Executed  bool
	Running   bool
}

// DangerousPaths that should never be targeted.
var DangerousPaths = []string{
	"/System",
	"/usr",
	"/bin",
	"/sbin",
	"/etc",
	"/Library/Apple",
	"/private/var/db",
	"/Applications/Utilities",
}

// ParseCommands splits a block of text into individual commands.
// Skips empty lines and comments (lines starting with #).
func ParseCommands(input string) []Command {
	var commands []Command
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		commands = append(commands, Command{Raw: line})
	}
	return commands
}

// ValidateCommand checks a command against safety rules.
func ValidateCommand(cmd *Command) error {
	raw := cmd.Raw

	// Block sudo
	if strings.HasPrefix(raw, "sudo ") || strings.Contains(raw, " sudo ") {
		cmd.Error = "sudo commands are not allowed"
		return fmt.Errorf("sudo commands are not allowed")
	}

	// Check for dangerous paths
	for _, dp := range DangerousPaths {
		if strings.Contains(raw, dp) {
			cmd.Error = fmt.Sprintf("targets protected path: %s", dp)
			return fmt.Errorf("targets protected path: %s", dp)
		}
	}

	cmd.Validated = true
	return nil
}

// ExecuteCommand runs a single shell command and captures output.
func ExecuteCommand(cmd *Command) error {
	cmd.Running = true
	defer func() { cmd.Running = false }()

	if !cmd.Validated {
		if err := ValidateCommand(cmd); err != nil {
			return err
		}
	}

	out, err := exec.Command("sh", "-c", cmd.Raw).CombinedOutput()
	cmd.Output = string(out)
	cmd.Executed = true

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			cmd.ExitCode = exitErr.ExitCode()
		} else {
			cmd.ExitCode = 1
		}
		cmd.Error = err.Error()
		return err
	}

	cmd.ExitCode = 0
	return nil
}
