//go:build !windows

package cli

import (
	"os"
	"os/exec"
)

// defaultEditor returns the default editor command to use for editing configurations.
// It prioritizes the EDITOR environment variable, then checks for VS Code, and falls back to vi.
func defaultEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if path, err := exec.LookPath("code"); err == nil && path != "" {
		return "code --wait"
	}
	return "vi"
}
