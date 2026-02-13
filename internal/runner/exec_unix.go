//go:build !windows

package runner

import (
	"os/exec"
	"syscall"
)

func execClaude(claudePath string, args []string, env []string) error {
	return syscall.Exec(claudePath, args, env)
}

func lookClaude() (string, error) {
	return exec.LookPath("claude")
}
