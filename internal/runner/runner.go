// Package runner provides execution of the Claude CLI with provider-specific environment variables.
// It handles setting ANTHROPIC_BASE_URL, ANTHROPIC_AUTH_TOKEN, and other provider configuration
// before launching Claude.
package runner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/lucas-stellet/switcher/internal/config"
)

// Run launches Claude with the given provider configuration and arguments.
// It sets the necessary environment variables and replaces the current process with Claude.
func Run(provider config.Provider, claudeArgs []string) error {
	claudePath, err := lookClaude()
	if err != nil {
		return fmt.Errorf("claude not found in PATH: %w", err)
	}

	env := os.Environ()
	env = setEnv(env, "ANTHROPIC_BASE_URL", provider.BaseURL)
	env = removeEnv(env, "ANTHROPIC_API_KEY")
	env = setEnv(env, "ANTHROPIC_AUTH_TOKEN", provider.APIKey)
	for k, v := range provider.Env {
		env = setEnv(env, k, v)
	}
	env = removeEnv(env, "CLAUDECODE")
	env = removeEnv(env, "CLAUDE_CODE")

	args := []string{"claude"}
	if provider.Model != "" {
		args = append(args, "--model", provider.Model)
	}
	args = append(args, claudeArgs...)

	return execClaude(claudePath, args, env)
}

// RunCommand runs an arbitrary command with the given provider's environment variables.
// It uses exec.Command to run the command as a subprocess, passing through stdin/stdout/stderr.
func RunCommand(provider config.Provider, cmdName string, cmdArgs []string) error {
	env := os.Environ()
	env = setEnv(env, "ANTHROPIC_BASE_URL", provider.BaseURL)
	env = removeEnv(env, "ANTHROPIC_API_KEY")
	env = setEnv(env, "ANTHROPIC_AUTH_TOKEN", provider.APIKey)
	for k, v := range provider.Env {
		env = setEnv(env, k, v)
	}
	env = removeEnv(env, "CLAUDECODE")
	env = removeEnv(env, "CLAUDE_CODE")

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// removeEnv removes an environment variable from the env list.
// Returns a new slice without the specified variable.
func removeEnv(env []string, key string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			return append(env[:i], env[i+1:]...)
		}
	}
	return env
}

// setEnv sets or updates an environment variable in the env list.
// Returns the modified env list (modifying in-place if the variable exists, or appending if it doesn't).
func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
