//go:build windows

package cli

import (
	"os"
	"os/exec"
)

func defaultEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if path, err := exec.LookPath("code"); err == nil && path != "" {
		return "code --wait"
	}
	return "notepad"
}
