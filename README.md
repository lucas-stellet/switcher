# switcher

A CLI tool to launch [Claude Code](https://docs.anthropic.com/en/docs/claude-code) with different AI providers.

## Why?

Claude Code is an excellent coding agent, but it's locked to Anthropic's API by default. Several providers now offer OpenAI-compatible endpoints that can work as drop-in replacements — Moonshot, DeepSeek, OpenRouter, ZhipuAI, MiniMax, and others.

Switching between them manually means juggling environment variables every time: `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, model flags... It gets old fast.

**switcher** wraps that into a single command. You configure your providers once and launch Claude Code pointing at any of them.

## Inspiration

This project was inspired by the growing ecosystem of AI providers offering API compatibility and the friction of managing multiple configurations by hand. Instead of shell aliases or `.env` files scattered around, switcher gives you a dedicated tool to manage and switch providers cleanly.

## Install

### Quick install (macOS / Linux)

```sh
curl -sSL https://raw.githubusercontent.com/lucas-stellet/switcher/main/install.sh | sh
```

### Quick install (Windows PowerShell)

```powershell
irm https://raw.githubusercontent.com/lucas-stellet/switcher/main/install.ps1 | iex
```

### Homebrew

```sh
brew install lucas-stellet/switcher/switcher
```

### From source

```sh
go install github.com/lucas-stellet/switcher@latest
```

### Manual build

```sh
git clone https://github.com/lucas-stellet/switcher.git
cd switcher
make install   # installs to ~/.local/bin (no sudo required)
```

## Usage

```
switcher <provider> claude [args...]
```

On first run, switcher creates `~/.switcher.json` with a set of pre-configured providers (API keys left blank for you to fill in).

### Initialize config

```sh
switcher init
```

Creates `~/.switcher.json` with all default providers. Safe to run if the file already exists — it won't overwrite your config.

### Launch Claude with a provider

```sh
switcher minimax claude
switcher moonshot claude
switcher deepseek claude -p "explain this function"
switcher openrouter claude --dangerously-skip-permissions
```

### Run arbitrary commands with provider env vars

```sh
switcher run moonshot -- env | grep ANTHROPIC
switcher run deepseek -- ./test-e2e.sh
```

The `--` separator is required. Everything after `--` is passed to the command. This is useful for running E2E test scripts that need specific `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN` settings.

### List providers

```sh
switcher list
```

```
  deepseek       - DeepSeek [deepseek-chat]  (no api key)
  minimax        - MiniMax [MiniMax-M2.5]  (no api key)
  moonshot       - Moonshot AI [kimi-k2.5]  (no api key)
  openrouter     - OpenRouter  (no api key)
  zai            - ZhipuAI [glm-5]  (no api key)
```

### Add a provider

```sh
switcher add my-provider
```

Prompts interactively for description, base URL, API key, and model.

### Edit a provider

```sh
switcher edit moonshot
```

Opens the provider config in `$EDITOR` as JSON.

### Remove a provider

```sh
switcher rm my-provider
```

## Configuration

The config file lives at `~/.switcher.json`. You can edit it directly or use the CLI commands. Each provider has:

| Field         | Description                              |
|---------------|------------------------------------------|
| `description` | Human-readable name                      |
| `base_url`    | API endpoint (OpenAI-compatible)         |
| `api_key`     | Your API key for the provider            |
| `model`       | Default model to use                     |
| `env`         | Extra environment variables to set       |

> **Note:** Variables in `env` are applied _after_ `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN`. This means if you set either of these inside `env`, they will override the corresponding `base_url` and `api_key` fields. If a provider's default URL has changed, PRs are welcome.

Example:

```json
{
  "providers": {
    "moonshot": {
      "description": "Moonshot AI",
      "base_url": "https://api.moonshot.ai/anthropic",
      "api_key": "sk-...",
      "model": "kimi-k2.5",
      "env": {
        "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
        "ANTHROPIC_SMALL_FAST_MODEL": "kimi-k2.5",
        "ANTHROPIC_DEFAULT_SONNET_MODEL": "kimi-k2.5",
        "ANTHROPIC_DEFAULT_OPUS_MODEL": "kimi-k2.5",
        "ANTHROPIC_DEFAULT_HAIKU_MODEL": "kimi-k2.5",
        "CLAUDE_CODE_SUBAGENT_MODEL": "kimi-k2.5"
      }
    }
  }
}
```

## Default providers

| Name         | Provider            | Model             | Notes                                                              |
|--------------|---------------------|-------------------|--------------------------------------------------------------------|
| `moonshot`   | Moonshot AI         | `kimi-k2.5`      | Anthropic-compatible API, includes subagent and model env vars     |
| `zai`        | ZhipuAI (Z.AI)     | `glm-5`          | 744B MoE, 40B active params, open-source, trained on Huawei Ascend |
| `openrouter` | OpenRouter          | (user's choice)   | Gateway to 400+ models from all major providers                    |
| `deepseek`   | DeepSeek            | `deepseek-chat`   | Points to latest stable (currently V4), strong reasoning & coding  |
| `minimax`    | MiniMax             | `MiniMax-M2.5`   | Anthropic-compatible API, includes model env vars for all tiers    |

## How it works

switcher finds the `claude` binary in your PATH, sets `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN` to point at your chosen provider, and launches Claude. On macOS/Linux it calls `exec` — replacing itself with the Claude process with zero overhead. On Windows it runs Claude as a subprocess.

## License

MIT
