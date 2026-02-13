//go:build windows

package cli

import (
	"os"
	"os/exec"
)

// defaultEditor returns the default editor command to use for editing configurations.
// It prioritizes the EDITOR environment variable, then checks for VS Code, and falls back to notepad.
func defaultEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if _, err := exec.LookPath("code"); err == nil {
		return "code --wait"
	}
	return "notepad"
}
