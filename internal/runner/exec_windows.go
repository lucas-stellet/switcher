//go:build windows

package runner

import (
	"os"
	"os/exec"
)

func execClaude(claudePath string, args []string, env []string) error {
	cmd := exec.Command(claudePath, args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func lookClaude() (string, error) {
	// On Windows, try "claude" first (LookPath adds .exe automatically),
	// then try "claude.cmd" for npm global installs.
	path, err := exec.LookPath("claude")
	if err != nil {
		path, err = exec.LookPath("claude.cmd")
	}
	return path, err
}
