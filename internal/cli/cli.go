// Package cli provides command-line interface functionality for switcher.
// It handles parsing commands and dispatching them to appropriate handlers.
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lucas-stellet/switcher/internal/config"
	"github.com/lucas-stellet/switcher/internal/runner"
	"github.com/lucas-stellet/switcher/internal/updater"
)

// Run processes command-line arguments and executes the appropriate command.
// It is the main entry point for the CLI dispatcher.
func Run(args []string, version string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	if args[0] != "update" {
		if msg := updater.CheckForUpdate(version); msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
	}

	switch args[0] {
	case "init":
		if hasHelpFlag(args[1:]) {
			printInitHelp()
			return nil
		}
		return cmdInit()
	case "list", "ls":
		return cmdList()
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: switcher add <name>")
		}
		return cmdAdd(args[1])
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("usage: switcher remove <name>")
		}
		return cmdRemove(args[1])
	case "edit":
		if len(args) < 2 {
			return fmt.Errorf("usage: switcher edit <name>")
		}
		return cmdEdit(args[1])
	case "run":
		return cmdRunCommand(args[1:])
	case "update":
		return updater.SelfUpdate(version)
	case "version", "--version", "-v":
		fmt.Printf("switcher %s\n", version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return cmdRun(args)
	}
}

func printUsage() {
	fmt.Print(`switcher - launch Claude Code with different AI providers

Usage:
  switcher init                          Create config with default providers
  switcher <provider> claude [args...]   Launch Claude with a provider
  switcher run <provider> -- <cmd>        Run command with provider's env vars
  switcher list                          List configured providers
  switcher add <name>                    Add a provider interactively
  switcher remove <name>                 Remove a provider
  switcher edit <name>                   Edit a provider in $EDITOR
  switcher update                        Update switcher to the latest version
  switcher version                       Print current version

Examples:
  switcher moonshot claude --dangerously-skip-permissions
  switcher zai claude -p "hello world"
  switcher openrouter claude
  switcher deepseek -- env | grep ANTHROPIC
`)
}

func hasHelpFlag(args []string) bool {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}

func printInitHelp() {
	fmt.Print(`switcher init - create config file with default providers

Usage:
  switcher init

Creates ~/.switcher.json with all built-in providers pre-configured
(API keys left blank for you to fill in).

If the config file already exists, it is not modified.

Default providers included:
  deepseek       DeepSeek            [deepseek-chat]
  minimax        MiniMax             [MiniMax-M2.5]
  moonshot       Moonshot AI         [kimi-k2.5]
  openrouter     OpenRouter
  zai            ZhipuAI             [glm-5]

After running init, set your API keys with:
  switcher edit <provider>
`)
}

func cmdInit() error {
	path, created, err := config.Init()
	if err != nil {
		return err
	}

	if !created {
		fmt.Printf("Config already exists at %s\n", path)
		return nil
	}

	fmt.Printf("Config created at %s\n", path)
	fmt.Println("Use 'switcher edit <provider>' to set your API keys.")
	return nil
}

func cmdList() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	providers := cfg.List()
	if len(providers) == 0 {
		fmt.Println("No providers configured. Use 'switcher add <name>' to add one.")
		return nil
	}

	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		p := providers[name]
		status := "  (no api key)"
		if p.APIKey != "" {
			status = ""
		}
		desc := ""
		if p.Description != "" {
			desc = " - " + p.Description
		}
		model := ""
		if p.Model != "" {
			model = " [" + p.Model + "]"
		}
		fmt.Printf("  %-14s%s%s%s\n", name, desc, model, status)
	}

	return nil
}

func cmdAdd(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Get(name); exists {
		return fmt.Errorf("provider %q already exists. Use 'switcher edit %s' to modify it", name, name)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Adding provider %q\n\n", name)

	description, err := prompt(reader, "Description: ")
	if err != nil {
		return err
	}

	baseURL, err := prompt(reader, "Base URL: ")
	if err != nil {
		return err
	}

	apiKey, err := prompt(reader, "API Key: ")
	if err != nil {
		return err
	}

	model, err := prompt(reader, "Model: ")
	if err != nil {
		return err
	}

	p := config.Provider{
		Description: description,
		BaseURL:     baseURL,
		APIKey:      apiKey,
		Model:       model,
		Env:         map[string]string{},
	}

	cfg.Add(name, p)
	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Printf("\nProvider %q added.\n", name)
	return nil
}

func cmdRemove(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if !cfg.Remove(name) {
		return fmt.Errorf("provider %q not found", name)
	}

	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Printf("Provider %q removed.\n", name)
	return nil
}

func cmdEdit(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	p, ok := cfg.Get(name)
	if !ok {
		dp, isDefault := config.DefaultProvider(name)
		if !isDefault {
			return fmt.Errorf("provider %q not found", name)
		}
		p = dp
		cfg.Add(name, p)
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Provider %q added from defaults.\n", name)
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding provider: %w", err)
	}

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("switcher-%s.json", name))
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	editor := defaultEditor()

	parts := strings.Fields(editor)
	cmdArgs := append(parts[1:], tmpFile)
	cmd := exec.Command(parts[0], cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	edited, err := os.ReadFile(tmpFile)
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	var updated config.Provider
	if err := json.Unmarshal(edited, &updated); err != nil {
		return fmt.Errorf("parsing edited provider: %w", err)
	}

	cfg.Add(name, updated)
	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Printf("Provider %q updated.\n", name)
	return nil
}

func cmdRun(args []string) error {
	providerName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	p, ok := cfg.Get(providerName)
	if !ok {
		return fmt.Errorf("unknown provider %q. Use 'switcher list' to see available providers", providerName)
	}

	if p.APIKey == "" {
		return fmt.Errorf("provider %q has no API key configured. Use 'switcher edit %s' to set one", providerName, providerName)
	}

	// Skip "claude" if it's the second argument
	claudeArgs := args[1:]
	if len(claudeArgs) > 0 && claudeArgs[0] == "claude" {
		claudeArgs = claudeArgs[1:]
	}

	printLaunchInfo(providerName, p)

	return runner.Run(p, claudeArgs)
}

func cmdRunCommand(args []string) error {
	// Find the -- separator
	sepIdx := -1
	for i, arg := range args {
		if arg == "--" {
			sepIdx = i
			break
		}
	}

	if sepIdx == -1 {
		return fmt.Errorf("missing '--' separator. Usage: switcher run <provider> -- <command> [args...]")
	}

	if sepIdx == len(args)-1 {
		return fmt.Errorf("no command specified after '--'. Usage: switcher run <provider> -- <command> [args...]")
	}

	providerName := args[:sepIdx][0]
	commandArgs := args[sepIdx+1:]
	cmdName := commandArgs[0]
	cmdArgs := commandArgs[1:]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	p, ok := cfg.Get(providerName)
	if !ok {
		return fmt.Errorf("unknown provider %q. Use 'switcher list' to see available providers", providerName)
	}

	if p.APIKey == "" {
		return fmt.Errorf("provider %q has no API key configured. Use 'switcher edit %s' to set one", providerName, providerName)
	}

	printLaunchInfo(providerName, p)

	return runner.RunCommand(p, cmdName, cmdArgs)
}

func printLaunchInfo(name string, p config.Provider) {
	fmt.Fprintf(os.Stderr, "Launching %s...\n", name)
	if p.Model != "" {
		fmt.Fprintf(os.Stderr, "  Model:   %s\n", p.Model)
	}

	aliases := []struct {
		label string
		key   string
	}{
		{"Opus", "ANTHROPIC_DEFAULT_OPUS_MODEL"},
		{"Sonnet", "ANTHROPIC_DEFAULT_SONNET_MODEL"},
		{"Haiku", "ANTHROPIC_DEFAULT_HAIKU_MODEL"},
	}

	for _, a := range aliases {
		if v, ok := p.Env[a.key]; ok && v != "" {
			fmt.Fprintf(os.Stderr, "  %-8s → %s\n", a.label, v)
		}
	}

	fmt.Fprintln(os.Stderr)
}

func prompt(reader *bufio.Reader, label string) (string, error) {
	fmt.Print(label)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	return strings.TrimSpace(line), nil
}
