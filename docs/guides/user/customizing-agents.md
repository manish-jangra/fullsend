# Customizing Agents

This guide explains how to customize fullsend agents for your organization and repositories through harness configurations and layered content resolution.

## Harness Configuration

Each agent run is configured by a harness YAML file that defines the complete execution environment. These files live in the `.fullsend` config repo (per-org mode) or `.fullsend/customized/harness/` (per-repo mode).

### Harness YAML Structure

```yaml
agent: agents/coder.md           # Agent definition (Claude agent file)
image: ghcr.io/fullsend/sandbox  # Container image for sandbox
model: claude-sonnet-4-6         # Inference model
policy: policies/coder.yaml      # Security policy
timeout_minutes: 30              # Max execution time

skills:                          # Skills loaded into agent
  - skills/git.md
  - skills/github-pr.md

plugins:                         # MCP plugins
  - name: github
    config: plugins/github.json

providers:                       # Inference providers
  - name: vertex
    type: vertex
    credentials: wif
    config:
      project: ${FULLSEND_GCP_PROJECT_ID}
      region: ${FULLSEND_GCP_REGION}

host_files:                      # Files copied into sandbox
  - src: ${GITHUB_WORKSPACE}/.fullsend/customized/scripts/
    dest: /tmp/workspace/scripts/
    optional: true
  - src: configs/tool-config.json
    dest: /tmp/workspace/.config/tool-config.json
    expand: true                 # Expand ${VAR} in file content

pre_script: scripts/pre-code.sh    # Runs on host before agent
post_script: scripts/post-code.sh  # Runs on host after agent

runner_env:                      # Env vars passed to runner
  - GITHUB_TOKEN
  - FULLSEND_OUTPUT_DIR

validation_loop:                 # Iterative validation
  script: scripts/validate-output-schema.sh
  max_iterations: 3
  feedback_mode: stderr          # Feed script stderr as feedback

security:
  enabled: true
  fail_mode: closed              # Block on security failure
  host_scanners:
    - llm-guard                  # ML-based content scanning
    - tirith                     # Terminal security
  sandbox_hooks:
    - injection-detect           # Prompt injection detection
    - ssrf-validate              # SSRF prevention
    - unicode-normalize          # Unicode attack normalization
    - secret-redact              # Secret leak prevention
  escalation: block              # Action on security violation
  trace: true                    # Enable security event tracing
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
│  Secret name: {org}-{role}-github-app-pem                   │
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

Create `.fullsend/customized/agents/coder.md` to override the default coder agent with org-specific instructions:

```markdown
# Code Agent (Customized)

[Custom instructions for your org...]
```

### Customizing Harness Configuration

Create `.fullsend/customized/harness/coder.yaml` to modify the coder agent's execution environment:

```yaml
# Extend the upstream coder harness
agent: agents/coder.md
model: claude-opus-4-6  # Use Opus instead of Sonnet
timeout_minutes: 45     # Increase timeout for complex tasks

# Add org-specific skills
skills:
  - skills/git.md
  - skills/github-pr.md
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
│       ├── agents/coder.md        # Repo-specific agent instructions
│       ├── skills/repo-skill.md   # Repo-specific skill
│       └── harness/coder.yaml     # Repo-specific harness config
```

## See Also

- [Installation Guide](../admin/installation.md) - Initial setup
- [Bugfix Workflow](bugfix-workflow.md) - How agents work together
- [ADR 0035: Layered Content Resolution](../../ADRs/0035-layered-content-resolution.md)
