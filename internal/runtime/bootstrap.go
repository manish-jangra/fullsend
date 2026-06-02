package runtime

import "github.com/fullsend-ai/fullsend/internal/security"

// BootstrapInput is the portable contract every runtime needs to provision
// agent content into the sandbox. Implementations live outside this package
// (runner adapter, tests).
type BootstrapInput interface {
	SandboxName() string
	AgentPath() string
	SkillDirs() []string
	PluginDirs() []string
}

// ClaudeHooksBootstrap is an optional extension for Claude Code sandbox tool hooks.
// ClaudeRuntime.Bootstrap type-asserts for this; other runtimes ignore it.
type ClaudeHooksBootstrap interface {
	ClaudeSandboxHooks() security.ClaudeSandboxHooks
}
