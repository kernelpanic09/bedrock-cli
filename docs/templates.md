# Templates

Templates let you save reusable prompts with named variables. They're stored as plain text files
in `~/.config/bedrock-cli/templates/`.

## Creating a template

Write a text file, optionally with YAML frontmatter:

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

Be specific. Cite line numbers where you can. Suggest concrete fixes.
```

Save it:

```sh
bedrock-cli template create code-review --file code-review.txt
```

## Frontmatter fields

| Field         | Type     | Description |
|---------------|----------|-------------|
| `description` | string   | Short description shown in `template list` |
| `variables`   | array    | Variable definitions (see below) |
| `defaultModel`| string   | Model alias or ID to use for this template |

### Variable definition

```yaml
variables:
  - name: Topic       # variable name, used as {{.Topic}} in the body
    required: true    # error if not provided
  - name: Style
    default: technical  # used if --var Style=... is not provided
```

## Running a template

```sh
bedrock-cli template run code-review --var Diff="$(git diff)" --var Focus=performance
```

The `--model` flag overrides the template's `defaultModel`.

## Variables in the body

Templates use Go's `text/template` syntax:

```
{{.VarName}}           # insert a variable
{{if .VarName}} ... {{end}}   # conditional
```

## Listing templates

```sh
bedrock-cli template list
bedrock-cli template list --json
```

## Built-in example templates

The repo ships four example templates in `templates/`. Import them all at once:

```sh
for f in templates/*.txt; do
  name=$(basename "$f" .txt)
  bedrock-cli template create "$name" --file "$f"
done
```

Available:
- `code-review` - review a git diff with configurable focus area
- `commit-message` - generate a conventional commit message from a diff
- `explain` - explain a concept at a senior-engineer level
- `summarize` - summarize long text in a configurable length
