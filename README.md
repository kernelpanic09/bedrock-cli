# bedrock-cli

Send prompts to Bedrock without writing JSON. Compare models. Track costs. Manage templates. All from your terminal.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](go.mod)
[![Release](https://img.shields.io/github/v/release/kernelpanic09/bedrock-cli)](https://github.com/kernelpanic09/bedrock-cli/releases)

> **Status:** alpha. Core commands work. Rough edges exist. PRs welcome.

---

The `aws bedrock-runtime invoke-model` command wants a JSON body, a content-type header, and a full model ID every time. `bedrock-cli` skips that.

```sh
bedrock-cli "What is HCL?"
```

<!-- screencast coming -->

---

## Install

**Homebrew** (tap coming after first release):

```sh
brew install kernelpanic09/tap/bedrock-cli
```

**go install** (Go 1.22+ required):

```sh
go install github.com/kernelpanic09/bedrock-cli/cmd/bedrock-cli@latest
```

**Download a binary** from the [releases page](https://github.com/kernelpanic09/bedrock-cli/releases) for Linux, macOS, or Windows.

---

## Quick start

```sh
# Set your region and default model (one-time)
bedrock-cli config set region us-east-1
bedrock-cli config set default-model sonnet

# Or just run something - it'll prompt you on first use
bedrock-cli "Explain the difference between OIDC and SAML"
```

You need AWS credentials in the standard places (`~/.aws/credentials`, env vars, IAM role). The IAM principal needs `bedrock:InvokeModel` and `bedrock:ConverseStream`. See [docs/configuration.md](docs/configuration.md) for the exact policy.

---

## Commands

| Command | Description |
|---------|-------------|
| `bedrock-cli <prompt>` | Send a prompt (shorthand for `prompt`) |
| `bedrock-cli prompt <text>` | Send a prompt with full flags |
| `bedrock-cli compare --models a,b <text>` | Side-by-side model comparison |
| `bedrock-cli template create <name>` | Save a prompt template from a file |
| `bedrock-cli template list` | List saved templates |
| `bedrock-cli template run <name>` | Render a template and send it |
| `bedrock-cli template delete <name>` | Delete a template |
| `bedrock-cli models list` | List models from the built-in catalog |
| `bedrock-cli models info <alias>` | Show pricing and details for a model |
| `bedrock-cli cost summary` | Total spend across tracked invocations |
| `bedrock-cli cost by-model` | Spend broken down by model |
| `bedrock-cli kb list` | List Bedrock Knowledge Bases |
| `bedrock-cli kb query <id> <text>` | Query a Knowledge Base |
| `bedrock-cli config` | Show current config |
| `bedrock-cli config set <key> <val>` | Change a config value |
| `bedrock-cli version` | Print version |

---

## Model aliases

You don't need the full model ID. Use short aliases:

| Alias | Model ID |
|-------|----------|
| `haiku` | `anthropic.claude-haiku-4-5-20251001-v1:0` |
| `sonnet` | `anthropic.claude-sonnet-4-6-20250514-v1:0` |
| `opus` | `anthropic.claude-opus-4-7-20250219-v1:0` |
| `llama-3-70b` | `meta.llama3-70b-instruct-v1:0` |
| `mistral-7b` | `mistral.mistral-7b-instruct-v0:2` |
| `titan-text` | `amazon.titan-text-express-v1` |

Full model IDs work too. Run `bedrock-cli models list` for the full catalog.

---

## Streaming

Responses stream by default. Use `--no-stream` to get the full response at once.

```sh
bedrock-cli prompt --no-stream "Write a haiku about YAML"
```

---

## Caching

Responses are cached in `~/.cache/bedrock-cli/responses/` keyed by SHA256 of the model + prompt + parameters. Identical calls don't hit the API twice.

```sh
# Bypass the cache
bedrock-cli prompt --no-cache "Latest news in Kubernetes"

# Disable caching globally
bedrock-cli config set cache-ttl -1
```

---

## Templates

Templates are text files with optional YAML frontmatter. Variables use Go template syntax.

```
---
description: Review a code diff for issues
variables:
  - name: Diff
    required: true
  - name: Focus
    default: security and bugs
defaultModel: sonnet
---
Review this code change with a focus on {{.Focus}}:

{{.Diff}}
```

```sh
# Save a template
bedrock-cli template create code-review --file templates/code-review.txt

# Run it
bedrock-cli template run code-review --var Diff="$(git diff)" --var Focus=performance

# Generate commit messages
bedrock-cli template run commit-message --var Diff="$(git diff --cached)"
```

The repo includes four starter templates in `templates/`. See [docs/templates.md](docs/templates.md).

---

## Cost tracking

Every invocation is logged to a SQLite database at `~/.cache/bedrock-cli/usage.db`.

```sh
bedrock-cli cost summary
bedrock-cli cost summary --since 7d
bedrock-cli cost by-model
```

Pricing data is embedded in the catalog. It's accurate as of the last catalog update but Bedrock pricing changes, so verify against the AWS pricing page for anything financial.

---

## Configuration

```toml
# ~/.config/bedrock-cli/config.toml
default-model = "sonnet"
region        = "us-east-1"
max-tokens    = 4096
temperature   = 0.7
cache-ttl     = 0       # 0 = cache forever; -1 = disable
show-cost     = true
no-stream     = false
no-color      = false
```

All keys can be overridden via env vars: `BEDROCK_CLI_DEFAULT_MODEL`, `BEDROCK_CLI_REGION`, etc.

See [docs/configuration.md](docs/configuration.md) for the full reference including IAM policy.

---

## Related projects

- [agents-platform](https://github.com/kernelpanic09/agents-platform) - multi-agent scheduling and orchestration
- [mcp-server-aws](https://github.com/kernelpanic09/mcp-server-aws) - MCP server for AWS service access
- [terraform-aws-modules](https://github.com/kernelpanic09/terraform-aws-modules) - Terraform modules for common AWS patterns
- [k8s-ai-operator](https://github.com/kernelpanic09/k8s-ai-operator) - Kubernetes operator that integrates Bedrock models into cluster workflows

---

## Contributing

Issues and PRs welcome. Run `make test` before submitting. Run `make lint` if you have golangci-lint installed.

## License

MIT - see [LICENSE](LICENSE).
