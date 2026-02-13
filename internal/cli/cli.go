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
)

func Run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
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
  switcher list                          List configured providers
  switcher add <name>                    Add a provider interactively
  switcher remove <name>                 Remove a provider
  switcher edit <name>                   Edit a provider in $EDITOR

Examples:
  switcher moonshot claude --dangerously-skip-permissions
  switcher zai claude -p "hello world"
  switcher openrouter claude
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

	description := prompt(reader, "Description: ")
	baseURL := prompt(reader, "Base URL: ")
	apiKey := prompt(reader, "API Key: ")
	model := prompt(reader, "Model: ")

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

	return runner.Run(p, claudeArgs)
}

func prompt(reader *bufio.Reader, label string) string {
	fmt.Print(label)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
