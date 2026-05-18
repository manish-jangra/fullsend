# Customizing Agents

This guide explains how to customize fullsend agents for your organization and repositories through harness configurations and layered content resolution.

## Harness Configuration

Each agent run is configured by a harness YAML file that defines the complete execution environment. These files live in the `.fullsend` config repo (per-org mode) or `.fullsend/customized/harness/` (per-repo mode).

### Harness YAML Structure

A minimal harness configuration (based on actual fullsend agent harnesses):

```yaml
agent: agents/code.md
model: opus
image: ghcr.io/fullsend-ai/fullsend-code:latest
policy: policies/code.yaml
timeout_minutes: 35

skills:
  - skills/code-implementation

plugins:
  - plugins/gopls-lsp

host_files:
  - src: env/gcp-vertex.env
    dest: /tmp/workspace/.env.d/gcp-vertex.env
    expand: true
  - src: ${GOOGLE_APPLICATION_CREDENTIALS}
    dest: /tmp/workspace/.gcp-credentials.json
  - src: ${GCP_OIDC_TOKEN_FILE}
    dest: /tmp/workspace/.gcp-oidc-token
    optional: true

pre_script: scripts/pre-code.sh
post_script: scripts/post-code.sh

validation_loop:
  script: scripts/validate-output-schema.sh
  max_iterations: 2

runner_env:
  PUSH_TOKEN: "${PUSH_TOKEN}"
  REPO_FULL_NAME: "${REPO_FULL_NAME}"
  REPO_DIR: "${GITHUB_WORKSPACE}/target-repo"
```

**Optional fields** (all have secure defaults and can be omitted):

```yaml
providers:                       # Inference providers (loaded from providers/ dir)
  - vertex                       # References providers/vertex.yaml

validation_loop:
  feedback_mode: stderr          # "stderr", "stdout", or "exit_code" (optional)

security:                        # Security is enabled by default with fail_mode: closed
  enabled: true                  # All scanners enabled by default
  fail_mode: closed              # "closed" (reject on failure) or "open" (warn only)
  host_scanners:
    unicode_normalizer: true
    context_injection: true
    ssrf_validator: true
    secret_redactor: true
    llm_guard:
      enabled: true
      threshold: 0.92
      match_type: sentence
  sandbox_hooks:
    tirith:
      enabled: true
      fail_on: high              # "critical", "high", or "medium"
    ssrf_pretool: true
    secret_redact_posttool: true
    unicode_posttool: true
    context_suppress_posttool: true
    canary_pretool: true
    canary_posttool: true
  escalation:
    on_critical: halt            # "halt" or "review"
    review_label: requires-manual-review
  trace:
    enabled: true
```

## Layered Configuration Resolution

Fullsend uses a three-tier inheritance model for all configuration: agent definitions, skills, policies, harness definitions, and guardrails. Each tier can extend or override the one below it.

```
┌──────────────────────────────────────────────────────────────┐
│              Configuration Layering (ADR 0035)                │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Priority (highest wins):                                    │
│                                                              │
│  ┌──────────────────────┐                                    │
│  │ Per-repo overrides   │  .fullsend/customized/{dir}/       │
│  │ (target repo)        │  in the target repository          │
│  └──────────┬───────────┘                                    │
│             │ overrides                                      │
│  ┌──────────▼───────────┐                                    │
│  │ Org-level overrides  │  customized/{dir}/                 │
│  │ (.fullsend config    │  in the .fullsend config repo      │
│  │  repo)               │                                    │
│  └──────────┬───────────┘                                    │
│             │ overrides                                      │
│  ┌──────────▼───────────┐                                    │
│  │ Upstream defaults    │  Provided at runtime by reusable   │
│  │ (fullsend-ai/        │  workflow workspace preparation    │
│  │  fullsend)           │                                    │
│  └──────────────────────┘                                    │
│                                                              │
│  Layered directories:                                        │
│    agents/  skills/  schemas/  harness/                      │
│    policies/  scripts/  env/                                 │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### Customization Directory Structure

Orgs customize layered directories by placing overrides in `customized/` subdirectories:

```
.fullsend/                          (config repo)
├── customized/
│   ├── agents/                     → Override agent definitions
│   ├── skills/                     → Add/override skills
│   ├── schemas/                    → Override output schemas
│   ├── harness/                    → Override harness configs
│   ├── policies/                   → Override security policies
│   ├── scripts/                    → Override hook scripts
│   └── env/                        → Override environment files
└── .github/workflows/              → Reusable workflows
```

For per-repo mode, the same structure lives at `.fullsend/customized/` within the target repo.

## Agent Roles

Each agent role has its own identity, permissions, and purpose:

```
┌─────────────────────────────────────────────────────────────┐
│                   Agent Role Architecture                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Role          GitHub App                  Purpose          │
│  ─────         ──────────                  ───────          │
│  fullsend      {org}-fullsend[bot]         Dispatch/control │
│  triage        {org}-triage[bot]           Issue triage     │
│  coder         {org}-coder[bot]            Code generation  │
│  review        {org}-review[bot]           PR review        │
│  fix           (reuses coder app)          Fix failures     │
│  retro         {org}-retro[bot]            Retrospectives   │
│  prioritize    {org}-prioritize[bot]       Backlog priority │
│                                                             │
│  App naming: {org}-{role}                                   │
│  Bot naming: {org}-{role}[bot]                              │
│  PEM storage: GCP Secret Manager                            │
│  Secret name: fullsend-{org}--{role}-app-pem                │
│                                                             │
│  Note: "fix" role reuses the "coder" app — no separate      │
│  GitHub App is created for it.                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Customization Examples

### Adding a Custom Skill

Create `.fullsend/customized/skills/my-skill.md` in your config repo:

```markdown
# My Custom Skill

Custom domain knowledge for this organization.

## Examples

...
```

The skill will be automatically available to all agents that include `skills/my-skill.md` in their harness configuration.

### Overriding an Agent Definition

Create `.fullsend/customized/agents/code.md` to override the default code agent with org-specific instructions:

```markdown
# Code Agent (Customized)

[Custom instructions for your org...]
```

### Customizing Harness Configuration

Create `.fullsend/customized/harness/coder.yaml` to modify the coder agent's execution environment:

```yaml
# Extend the upstream code harness
agent: agents/code.md
model: claude-opus-4-6  # Use Opus instead of default
timeout_minutes: 45     # Increase timeout for complex tasks

# Add org-specific skills
skills:
  - skills/code-implementation
  - skills/my-custom-linting.md  # Org-specific skill

# Custom validation
validation_loop:
  script: scripts/custom-validate.sh
  max_iterations: 5
```

### Per-Repo Overrides

Target repos can override org-level customizations by placing files in `.fullsend/customized/` within the repo:

```
my-repo/
├── .fullsend/
│   └── customized/
│       ├── agents/code.md         # Repo-specific agent instructions
│       ├── skills/repo-skill.md   # Repo-specific skill
│       └── harness/code.yaml      # Repo-specific harness config
```

## See Also

- [Installation Guide](../admin/installation.md) - Initial setup
- [Bugfix Workflow](bugfix-workflow.md) - How agents work together
- [ADR 0035: Layered Content Resolution](../../ADRs/0035-layered-content-resolution.md)
