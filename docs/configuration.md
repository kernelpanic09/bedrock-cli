# Configuration

bedrock-cli reads config from `~/.config/bedrock-cli/config.toml`.

## Config file

```toml
default-model = "sonnet"
region        = "us-east-1"
max-tokens    = 4096
temperature   = 0.7
cache-ttl     = 0       # 0 = cache forever; -1 = disable cache
show-cost     = true
no-stream     = false
no-color      = false
```

Run `bedrock-cli config` to see all current values. Change any with:

```sh
bedrock-cli config set default-model haiku
bedrock-cli config set region us-west-2
```

## Environment variables

All config keys can be overridden via env vars using the `BEDROCK_CLI_` prefix:

| Env var                        | Config key      |
|-------------------------------|-----------------|
| `BEDROCK_CLI_DEFAULT_MODEL`   | `default-model` |
| `BEDROCK_CLI_REGION`          | `region`        |
| `BEDROCK_CLI_MAX_TOKENS`      | `max-tokens`    |
| `BEDROCK_CLI_TEMPERATURE`     | `temperature`   |
| `BEDROCK_CLI_CACHE_TTL`       | `cache-ttl`     |
| `BEDROCK_CLI_NO_STREAM`       | `no-stream`     |
| `BEDROCK_CLI_NO_COLOR`        | `no-color`      |
| `BEDROCK_CLI_SHOW_COST`       | `show-cost`     |

## AWS credentials

bedrock-cli uses the standard AWS credential chain. In order of precedence:

1. `AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY` env vars
2. `AWS_PROFILE` env var (selects a named profile from `~/.aws/credentials`)
3. `~/.aws/credentials` and `~/.aws/config` (default profile)
4. EC2/ECS/EKS instance metadata (IAM role attached to the instance/task/pod)
5. AWS SSO (if configured via `aws configure sso`)

The simplest setup for local use:

```sh
aws configure
# or
export AWS_PROFILE=my-sandbox-account
```

The required IAM permissions depend on which commands you use.

**Core (prompt, compare, models)**

```json
{
  "Effect": "Allow",
  "Action": [
    "bedrock:InvokeModel",
    "bedrock:InvokeModelWithResponseStream",
    "bedrock:ConverseStream",
    "bedrock:Converse",
    "bedrock:ListFoundationModels",
    "bedrock:GetFoundationModel"
  ],
  "Resource": "*"
}
```

**Guardrails (guardrails list/describe/check)**

```json
{
  "Effect": "Allow",
  "Action": [
    "bedrock:ListGuardrails",
    "bedrock:GetGuardrail",
    "bedrock:ApplyGuardrail"
  ],
  "Resource": "*"
}
```

**Agents (agents list/describe/invoke)**

```json
{
  "Effect": "Allow",
  "Action": [
    "bedrock-agent:ListAgents",
    "bedrock-agent:GetAgent",
    "bedrock-agent:ListAgentActionGroups",
    "bedrock-agent:ListAgentKnowledgeBases",
    "bedrock-agent-runtime:InvokeAgent"
  ],
  "Resource": "*"
}
```

**Knowledge Bases (kb list/describe/query/create/sync)**

```json
{
  "Effect": "Allow",
  "Action": [
    "bedrock-agent:ListKnowledgeBases",
    "bedrock-agent:GetKnowledgeBase",
    "bedrock-agent:CreateKnowledgeBase",
    "bedrock-agent:DeleteKnowledgeBase",
    "bedrock-agent:ListDataSources",
    "bedrock-agent:GetDataSource",
    "bedrock-agent:CreateDataSource",
    "bedrock-agent:DeleteDataSource",
    "bedrock-agent:StartIngestionJob",
    "bedrock-agent:ListIngestionJobs",
    "bedrock-agent-runtime:Retrieve",
    "s3:GetObject",
    "s3:PutObject",
    "s3:ListBucket"
  ],
  "Resource": "*"
}
```

## Region

Model availability varies by region. Most Claude models are available in `us-east-1` and `us-west-2`.
Check the [Bedrock model availability page](https://docs.aws.amazon.com/bedrock/latest/userguide/models-regions.html) if you get a `ResourceNotFoundException`.
