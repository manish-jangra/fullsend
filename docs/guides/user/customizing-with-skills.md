# Customizing Agents with Skills

Fullsend agents use [agent skills](https://agentskills.org/) — self-contained
markdown documents that teach an agent how to perform a specific task. Each OOTB
agent ships with built-in skills, and you can extend or replace them by
committing your own skills to your repository.

For general project-wide instructions (code style, test conventions,
architecture rules), see [Customizing with CLAUDE.md](customizing-with-claude-md.md).
This guide covers skills specifically.

## What is a skill?

A skill is a directory containing a `SKILL.md` file with YAML frontmatter and
structured instructions. The agent loads the skill by name and follows its
instructions during execution.

```
.agents/skills/my-skill/
  SKILL.md           # skill definition (required)
  scripts/           # supporting scripts (optional)
    helper-script.sh
  references/        # reference data (optional)
    data.json
```

For portability across agent runtimes, store skills in `.agents/skills/` and
symlink `.claude/skills` to it:

```bash
ln -s ../.agents/skills .claude/skills
```

This way, skills are discoverable by fullsend's agent runtime and by any local
agent tooling developers use when working on the repo directly.

The `SKILL.md` has frontmatter declaring the skill's name and description,
followed by step-by-step instructions:

```markdown
---
name: my-skill
description: >-
  One-line summary of what this skill does.
---

# My Skill

Instructions the agent follows when this skill is invoked.

## Step 1: Gather context

...

## Step 2: Produce output

...
```

Skills can reference companion scripts and data files in the same directory,
giving agents the ability to dynamically gather information at runtime.

## Adding skills to your repository

Place skills in `.agents/skills/` in your target repository and symlink
`.claude/skills` to `.agents/skills`. All agents operating on your repo will
discover them automatically:

```
your-repo/
  .agents/skills/
    customer-research/
      SKILL.md
      scripts/
        query-salesforce.sh
    deployment-checks/
      SKILL.md
  .claude/skills -> ../.agents/skills
```

## Skill shadowing

Each fullsend agent ships with built-in skills. You can **shadow** any of
these by providing your own skill with the same name. Your version replaces
the built-in one at runtime — no other configuration needed.

This is the most precise way to tune agent behavior. A shadowed skill is only
loaded by the agent that uses it, unlike `CLAUDE.md` instructions which are
loaded by every agent.

### How shadowing works

Fullsend uses a layered content resolution model
([ADR 0035](../../ADRs/0035-layered-content-resolution.md)). At runtime, the
agent's workspace is assembled by copying upstream defaults first, then
overlaying org-level customizations on top. When you provide a skill with the
same name as a built-in one, yours wins.

To shadow a skill, create it in your `.fullsend` config repo at
`customized/skills/<skill-name>/SKILL.md`. The directory name must match the
built-in skill name exactly.

### Built-in skills

These skills ship with fullsend and can be shadowed:

| Agent | Skill | Purpose |
|-------|-------|---------|
| [Triage](../../agents/triage.md) | `issue-labels` | Label discovery and application during triage |
| [Code](../../agents/code.md) | `code-implementation` | Step-by-step implementation procedure |
| [Review](../../agents/review.md) | `code-review`, `pr-review`, `docs-review` | Review evaluation across dimensions |
| [Fix](../../agents/fix.md) | `fix-review` | Review feedback interpretation and fix strategy |
| [Retro](../../agents/retro.md) | `retro-analysis`, `finding-agent-runs` | Workflow analysis and proposal generation |

### Extension points

These skill names are recognized by agents but do not ship with fullsend. You
provide them to unlock additional agent capabilities:

| Agent | Skill | Purpose |
|-------|-------|---------|
| [Prioritize](../../agents/prioritize.md) | `customer-research` | Customer data gathering for RICE scoring |

See each agent's documentation for concrete examples.

## When to use skills vs. CLAUDE.md

Use **skills** when you need to change how a specific agent performs a specific
task — especially when the customization involves domain knowledge, helper
scripts, or external data sources that only one agent needs.

Use **[CLAUDE.md](customizing-with-claude-md.md)** for broad instructions that
apply to all agents and human contributors alike.

## What not to do

- **Don't put secrets in skills.** They're committed to your repo. Use
  environment variables for anything sensitive — skills can reference
  `$ENV_VAR` in their scripts.
- **Don't duplicate CLAUDE.md content in skills.** If an instruction applies
  to all agents, put it in `CLAUDE.md`. Skills are for agent-specific behavior.
