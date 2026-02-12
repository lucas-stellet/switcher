package runner

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/lucas/switch/internal/config"
)

func Run(provider config.Provider, claudeArgs []string) error {
	claudePath, err := exec.LookPath("claude")
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

	args := []string{"claude"}
	if provider.Model != "" {
		args = append(args, "--model", provider.Model)
	}
	args = append(args, claudeArgs...)

	return syscall.Exec(claudePath, args, env)
}

func removeEnv(env []string, key string) []string {
	prefix := key + "="
	for i, e := range env {
		if len(e) >= len(prefix) && e[:len(prefix)] == prefix {
			return append(env[:i], env[i+1:]...)
		}
	}
	return env
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if len(e) >= len(prefix) && e[:len(prefix)] == prefix {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
