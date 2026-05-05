# Running agents locally

This guide walks through running fullsend agents on your machine using released binaries. No Go toolchain or source checkout is needed. Both macOS and Linux are supported with Podman as the container runtime.

> For building fullsend from source or contributing to the CLI, see [Local development](../dev/local-dev.md).

## Prerequisites

| Requirement | macOS | Linux |
|-------------|-------|-------|
| Container runtime | Podman Desktop with a running machine | Podman |
| OpenShell | 0.0.37-dev+ (Podman support) | 0.0.37-dev+ |
| GCP credentials | Service account key (`Vertex AI User` role) | Same |
| GitHub PAT | `repo` scope for the target org | Same |

> **No Go toolchain required.** On macOS, the CLI automatically downloads a Linux binary for the sandbox. See [Cross-platform binary resolution](#cross-platform-binary-resolution) below.

## 1. Download the fullsend CLI

Download the latest release from [GitHub Releases](https://github.com/fullsend-ai/fullsend/releases). Pick the archive matching your platform:

| Platform | Archive |
|----------|---------|
| macOS (Apple Silicon) | `fullsend_{version}_darwin_arm64.tar.gz` |
| macOS (Intel) | `fullsend_{version}_darwin_amd64.tar.gz` |
| Linux (x86_64) | `fullsend_{version}_linux_amd64.tar.gz` |
| Linux (arm64) | `fullsend_{version}_linux_arm64.tar.gz` |

Extract and move to a directory in your PATH:

```bash
tar xzf fullsend_0.4.0_darwin_arm64.tar.gz
sudo mv fullsend_0.4.0_darwin_arm64/fullsend /usr/local/bin/
```

Verify the installation:

```bash
fullsend version
```

## 2. Clone the .fullsend config directory

Each organization has a `.fullsend` config repo containing harness definitions, agent prompts, policies, scripts, and skills. Clone it to a local path:

```bash
gh repo clone <org>/.fullsend /tmp/fullsend-dot
```

## 3. Create an env file

Env files contain secrets and must never be committed. Keep your env file outside the repo (e.g. `/tmp/fullsend.env`).

### Core variables (all agents)

```bash
ANTHROPIC_VERTEX_PROJECT_ID=<gcp-project-with-vertex-ai>
CLOUD_ML_REGION=global
GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa-key.json
GH_TOKEN=<github-pat>
```

### Per-agent variables

Each agent requires additional variables via its harness `runner_env`. Add the ones needed for the agent you want to run:

| Variable | Agent(s) | Description |
|----------|----------|-------------|
| `GITHUB_ISSUE_URL` | triage | Full URL of the GitHub issue to triage |
| `REPO_FULL_NAME` | triage, code, review, fix | `owner/repo` of the target repository |
| `ISSUE_NUMBER` | code | Issue number the code agent should implement |
| `TARGET_BRANCH` | code, fix | Branch to base work on (e.g. `main`) |
| `PUSH_TOKEN` | code, fix | GitHub token with push access for the post-script |
| `PR_NUMBER` | review, fix | Pull request number to review or fix |
| `GITHUB_PR_URL` | review | Full URL of the pull request to review |
| `REVIEW_TOKEN` | review | GitHub token for posting review comments |

See the harness definitions in `harness/*.yaml` within your `.fullsend` config directory for the complete list per agent.

### arm64 image override

On arm64 hosts (Apple Silicon, Graviton), add:

```bash
FULLSEND_SANDBOX_IMAGE=ghcr.io/fullsend-ai/fullsend-sandbox:dev
```

The default `:latest` tag is amd64-only. The `:dev` tag is a multi-arch image that includes arm64 support. On amd64 hosts this line is not needed.

## 4. Run an agent

```bash
fullsend run <agent> \
  --fullsend-dir /tmp/fullsend-dot \
  --target-repo /path/to/repo \
  --env-file /tmp/fullsend.env
```

The `--env-file` flag loads environment variables before the harness is parsed. It is repeatable — later files override earlier ones.

### Example: triage an issue

```bash
fullsend run triage \
  --fullsend-dir /tmp/fullsend-dot \
  --target-repo /path/to/repo \
  --env-file /tmp/fullsend.env
```

### Example: review a pull request

```bash
fullsend run review \
  --fullsend-dir /tmp/fullsend-dot \
  --target-repo /path/to/repo \
  --env-file /tmp/fullsend.env
```

### Example: implement an issue

```bash
fullsend run code \
  --fullsend-dir /tmp/fullsend-dot \
  --target-repo /path/to/repo \
  --env-file /tmp/fullsend.env
```

### Example: fix review feedback on a PR

```bash
fullsend run fix \
  --fullsend-dir /tmp/fullsend-dot \
  --target-repo /path/to/repo \
  --env-file /tmp/fullsend.env
```

## Testing without side effects

To run agent inference without the post-script (which posts PR comments, pushes branches, or creates PRs), use `--no-post-script`:

```bash
fullsend run review \
  --fullsend-dir /tmp/fullsend-dot \
  --target-repo /path/to/repo \
  --env-file /tmp/fullsend.env \
  --no-post-script
```

The agent runs normally inside the sandbox, but the post-script is skipped. This means `PUSH_TOKEN` and `REVIEW_TOKEN` are not needed in your env file — only the core variables (`GH_TOKEN`, GCP credentials) are required.

## Cross-platform binary resolution

The pre-agent security scan runs `fullsend scan context` inside the Linux sandbox. This requires a Linux binary, even when your host is macOS.

On **Linux**, the CLI copies its own executable into the sandbox — no extra steps needed.

On **macOS**, the CLI automatically obtains a Linux binary using the following priority:

| Priority | Strategy | When it applies |
|----------|----------|-----------------|
| 1 | `--fullsend-binary <path>` | Always used if provided — skips auto-resolution |
| 2 | Download from GitHub Release | CLI version matches a release (e.g. `0.4.0`) |
| 3 | Cross-compile from source | Go toolchain available, run from module root |
| 4 | Download latest release | Fallback when cross-compilation unavailable |

When using a released binary (this guide's workflow), **strategy 2 applies automatically** — the CLI downloads the matching Linux release. No Go toolchain or manual steps needed.

## Platform notes

### macOS

- **Podman machine**: ensure the Podman machine is running (`podman machine start`) before invoking fullsend. The CLI does not start it automatically.
- **Podman host-gateway**: if sandbox creation fails with `unable to replace "host-gateway"`, set `host_containers_internal_ip = "192.168.127.254"` under `[containers]` in `~/.config/containers/containers.conf` and restart the Podman machine.
- **Architecture mismatch**: if your sandbox image uses a different CPU architecture than the host (e.g. amd64 image on an arm64 Mac via QEMU emulation), set `FULLSEND_SANDBOX_ARCH=amd64` so the CLI downloads the correct binary. This is not needed in the typical setup where the Podman VM matches the host arch.

### Linux

- **Docker socket permissions**: if using Docker instead of Podman, your user must be in the `docker` group or run with `sudo`. Podman runs rootless by default.
- **SELinux**: on Fedora/RHEL, bind-mounted volumes may need the `:z` suffix for standalone `podman run`. OpenShell handles this automatically.

## Troubleshooting

**Sandbox creation fails immediately**
- Check that `podman machine start` has been run (macOS only)
- Verify OpenShell is installed: `openshell version`

**`Syntax error: "(" unexpected` inside sandbox**
- The macOS Mach-O binary was injected instead of a Linux ELF. Update to fullsend 0.4.0+ which auto-resolves the correct binary, or provide one explicitly with `--fullsend-binary`

**Agent fails with missing environment variable**
- Check your env file contains all variables listed in the agent's harness YAML (`harness/<agent>.yaml` in the `.fullsend` config directory)

**arm64 sandbox image pull fails**
- The default `:latest` tag is amd64-only. Add `FULLSEND_SANDBOX_IMAGE=ghcr.io/fullsend-ai/fullsend-sandbox:dev` to your env file

**`unable to replace "host-gateway"` on macOS**
- Set `host_containers_internal_ip = "192.168.127.254"` under `[containers]` in `~/.config/containers/containers.conf` and restart the Podman machine
