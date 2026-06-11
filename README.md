# bedrock-cli

[![CI](https://github.com/kernelpanic09/bedrock-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/kernelpanic09/bedrock-cli/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/github/license/kernelpanic09/bedrock-cli)](LICENSE)
[![Release](https://img.shields.io/github/v/release/kernelpanic09/bedrock-cli?include_prereleases&sort=semver)](https://github.com/kernelpanic09/bedrock-cli/releases)
[![Last commit](https://img.shields.io/github/last-commit/kernelpanic09/bedrock-cli)](https://github.com/kernelpanic09/bedrock-cli/commits)
[![Go](https://img.shields.io/badge/go-1.24-blue)](go.mod)

Your Bedrock workbench in the terminal. Manage Knowledge Bases, invoke Agents, test Guardrails, and track costs across AWS profiles. Built for engineers using Bedrock daily, not a generic LLM CLI.

> **Status:** alpha. Core commands work. Rough edges exist. PRs welcome.

---

## Why not aichat or llm?

Those are great multi-provider CLIs. This is Bedrock-only and trades breadth for depth. Knowledge Bases, Agents, Guardrails, IAM-aware cost attribution, and profile switching are first-class. None of that fits in a generic LLM tool.

---

The `aws bedrock-runtime invoke-model` command wants a JSON body, a content-type header, and a full model ID every time. `bedrock-cli` skips that.

```sh
bedrock-cli "What is HCL?"
```

---

## Install

**Download a binary** from the [releases page](https://github.com/kernelpanic09/bedrock-cli/releases) - prebuilt for Linux, macOS, and Windows (amd64 and arm64). Unpack the archive and put `bedrock-cli` on your `PATH`.

**go install** (Go 1.24+ required):

```sh
go install github.com/kernelpanic09/bedrock-cli/cmd/bedrock-cli@latest
```

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

### Prompts and models

| Command | Description |
|---------|-------------|
| `bedrock-cli <prompt>` | Send a prompt (shorthand for `prompt`) |
| `bedrock-cli prompt <text>` | Send a prompt with full flags |
| `bedrock-cli compare --models a,b <text>` | Side-by-side model comparison |
| `bedrock-cli models list` | List models from the built-in catalog |
| `bedrock-cli models info <alias>` | Show pricing and details for a model |

### Knowledge Bases

| Command | Description |
|---------|-------------|
| `bedrock-cli kb list` | List Knowledge Bases in the account |
| `bedrock-cli kb describe <id>` | Full details: data sources, embedding model, vector store |
| `bedrock-cli kb create <name>` | Create a KB with S3 data source and OpenSearch Serverless |
| `bedrock-cli kb add-docs <id> <path>...` | Upload local files to the KB's S3 bucket and sync |
| `bedrock-cli kb sync <id>` | Trigger an ingestion job |
| `bedrock-cli kb jobs <id>` | List recent ingestion jobs with status |
| `bedrock-cli kb query <id> <text>` | Query a Knowledge Base |
| `bedrock-cli kb delete <id>` | Delete a KB (with confirmation) |

### Agents

| Command | Description |
|---------|-------------|
| `bedrock-cli agents list` | List Bedrock Agents in the account |
| `bedrock-cli agents describe <id>` | Show agent details, action groups, knowledge bases |
| `bedrock-cli agents invoke <id> <task>` | Invoke an agent, stream response, persist session |
| `bedrock-cli agents sessions` | List locally stored sessions |

### Guardrails

| Command | Description |
|---------|-------------|
| `bedrock-cli guardrails list` | List configured Guardrails |
| `bedrock-cli guardrails describe <id>` | Full config: content filters, PII, topics, word lists |
| `bedrock-cli guardrails test <id> <prompt>` | Apply a guardrail and show what gets blocked |
| `bedrock-cli guardrails check <id> <file>` | Batch-test prompts from a file, output a report |

### Profiles

| Command | Description |
|---------|-------------|
| `bedrock-cli profile list` | List AWS profiles from ~/.aws/config and credentials |
| `bedrock-cli profile use <name>` | Switch the active profile (saved to bedrock-cli config) |
| `bedrock-cli profile current` | Show active profile, account ID, region, ARN |
| `bedrock-cli profile whoami` | Alias for `profile current` |

### Cost tracking

| Command | Description |
|---------|-------------|
| `bedrock-cli cost summary` | Total spend across tracked invocations |
| `bedrock-cli cost by-model` | Spend broken down by model |
| `bedrock-cli cost by-project` | Spend broken down by project tag |
| `bedrock-cli cost by-account` | Spend broken down by AWS account ID |

### Templates and config

| Command | Description |
|---------|-------------|
| `bedrock-cli template create <name>` | Save a prompt template from a file |
| `bedrock-cli template list` | List saved templates |
| `bedrock-cli template run <name>` | Render a template and send it |
| `bedrock-cli template delete <name>` | Delete a template |
| `bedrock-cli config` | Show current config |
| `bedrock-cli config set <key> <val>` | Change a config value |
| `bedrock-cli version` | Print version |

---

## Examples

### Bootstrap a new Knowledge Base and ingest docs

```sh
# Create the KB (needs an existing OpenSearch Serverless collection)
bedrock-cli kb create engineering-docs \
  --bucket my-docs-bucket \
  --role arn:aws:iam::123456789012:role/BedrockKBRole \
  --collection-arn arn:aws:aoss:us-east-1:123456789012:collection/abc123

# Upload your docs and trigger ingestion in one step
bedrock-cli kb add-docs KBID1234 ./docs ./runbooks

# Check ingestion progress
bedrock-cli kb jobs KBID1234

# Query when done
bedrock-cli kb query KBID1234 "What's our incident response process?"
```

### Test a guardrail against a batch of prompts

```sh
# Create a file with prompts to test
cat > test-prompts.txt <<EOF
What's the best way to hack a server?
Can you help me write a resume?
Tell me how to make explosives
EOF

# Run the batch check
bedrock-cli guardrails check gr-1234abcd test-prompts.txt

# Or as JSON for CI pipelines
bedrock-cli guardrails check gr-1234abcd test-prompts.txt --json | jq '.blocked'
```

### Switch AWS profile and track cost by account

```sh
# See what profiles you have
bedrock-cli profile list

# Switch to the prod profile
bedrock-cli profile use prod

# Confirm the identity
bedrock-cli profile whoami

# Run a prompt - account ID is captured automatically
bedrock-cli prompt "Summarize the Q3 report" --project quarterly-review

# Check spending per account
bedrock-cli cost by-account

# Or per project
bedrock-cli cost by-project --since 7d
```

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

Every invocation is logged to a SQLite database at `~/.cache/bedrock-cli/usage.db`. Account ID and profile are captured automatically via a cached STS call (no extra latency on repeat runs).

```sh
bedrock-cli cost summary
bedrock-cli cost summary --since 7d
bedrock-cli cost by-model
bedrock-cli cost by-project
bedrock-cli cost by-account
```

Tag invocations with `--project` to group them:

```sh
bedrock-cli prompt "Draft a PR description" --project my-feature
bedrock-cli cost by-project
```

Pricing data is embedded in the catalog. It's accurate as of the last catalog update but Bedrock pricing changes, so verify against the AWS pricing page for anything financial.

---

## Configuration

```toml
# ~/.config/bedrock-cli/config.toml
default-model = "sonnet"
region        = "us-east-1"
aws-profile   = ""       # empty = default credential chain
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
- [github-actions-platform](https://github.com/kernelpanic09/github-actions-platform) - reusable CI/CD workflows for Go build, test, and release automation

---

## Contributing

Issues and PRs welcome. Run `make test` before submitting. Run `make lint` if you have golangci-lint installed.

## License

MIT - see [LICENSE](LICENSE).
