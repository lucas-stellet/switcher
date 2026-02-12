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

	"github.com/lucas/switch/internal/config"
	"github.com/lucas/switch/internal/runner"
)

func Run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "list", "ls":
		return cmdList()
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: switch add <name>")
		}
		return cmdAdd(args[1])
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("usage: switch remove <name>")
		}
		return cmdRemove(args[1])
	case "edit":
		if len(args) < 2 {
			return fmt.Errorf("usage: switch edit <name>")
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
	fmt.Print(`switch - launch Claude Code with different AI providers

Usage:
  switch <provider> claude [args...]    Launch Claude with a provider
  switch list                           List configured providers
  switch add <name>                     Add a provider interactively
  switch remove <name>                  Remove a provider
  switch edit <name>                    Edit a provider in $EDITOR

Examples:
  switch moonshot claude --dangerously-skip-permissions
  switch zhipu claude -p "hello world"
  switch openrouter claude
`)
}

func cmdList() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	providers := cfg.List()
	if len(providers) == 0 {
		fmt.Println("No providers configured. Use 'switch add <name>' to add one.")
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
		return fmt.Errorf("provider %q already exists. Use 'switch edit %s' to modify it", name, name)
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
		return fmt.Errorf("provider %q not found", name)
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding provider: %w", err)
	}

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("switch-%s.json", name))
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, tmpFile)
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
		return fmt.Errorf("unknown provider %q. Use 'switch list' to see available providers", providerName)
	}

	if p.APIKey == "" {
		return fmt.Errorf("provider %q has no API key configured. Use 'switch edit %s' to set one", providerName, providerName)
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
