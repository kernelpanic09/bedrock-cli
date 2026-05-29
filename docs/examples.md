# Examples

## Day-to-day prompts

```sh
# Quick question, default model (sonnet)
bedrock-cli prompt "What's the difference between a goroutine and a thread?"

# Use haiku for cheap, fast answers
bedrock-cli prompt --model haiku "Write a one-liner to find large files in /var/log"

# Force fresh response even if it's cached
bedrock-cli prompt --no-cache --model sonnet "Current best practices for EKS node groups"

# Pipe input
cat error.log | bedrock-cli prompt "Explain this error and suggest a fix"
```

## Code review workflow

```sh
# Review staged changes
bedrock-cli template run code-review --var Diff="$(git diff --cached)"

# Focus on a specific area
bedrock-cli template run code-review \
  --var Diff="$(git diff HEAD~1)" \
  --var Focus="performance and memory allocation"

# Use opus for thorough reviews
bedrock-cli template run code-review \
  --model opus \
  --var Diff="$(git diff origin/main...HEAD)"
```

## Commit message generation

```sh
# From staged diff
bedrock-cli template run commit-message --var Diff="$(git diff --cached)"

# Pipe directly to clipboard (macOS)
bedrock-cli template run commit-message --var Diff="$(git diff --cached)" | pbcopy
```

## Comparing models

```sh
# Is haiku good enough for this use case?
bedrock-cli compare --models haiku,sonnet "Write a regex that validates IPv4 addresses"

# Three-way comparison
bedrock-cli compare --models haiku,sonnet,opus "Explain the CAP theorem"

# Get structured output for scripting
bedrock-cli compare --models haiku,sonnet "Summarize TLS handshake" --json | jq '.[].cost_usd'
```

## Cost tracking

```sh
# See everything you've spent
bedrock-cli cost summary

# Last 7 days
bedrock-cli cost summary --since 7d

# Which models are costing the most?
bedrock-cli cost by-model --since 30d

# Export for analysis
bedrock-cli cost summary --json > usage.json
```

## Knowledge base queries

```sh
# List your knowledge bases
bedrock-cli kb list

# Query one by ID
bedrock-cli kb query KB1234ABCD "What's our on-call rotation policy?"

# More results
bedrock-cli kb query KB1234ABCD "Incident response steps" --max-results 5

# JSON output for scripts
bedrock-cli kb query KB1234ABCD "deployment process" --json | jq '.[0].content'
```

## Scripting / CI

```sh
# Machine-readable output
bedrock-cli prompt "Summarize this PR" --json | jq -r '.response'

# Disable cost display in scripts
bedrock-cli config set show-cost false

# Disable color for log-friendly output
bedrock-cli prompt "..." --no-stream | tee output.txt
```

## Config management

```sh
# See what's configured
bedrock-cli config

# Switch to a cheaper default for experimentation
bedrock-cli config set default-model haiku

# Use a different region where a model is available
bedrock-cli config set region us-west-2

# Turn off streaming if your terminal doesn't handle it
bedrock-cli config set no-stream true
```
