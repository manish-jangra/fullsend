package security

import (
	"github.com/fullsend-ai/fullsend/internal/harness"
)

// ClaudeSandboxHooks configures Claude Code PreToolUse/PostToolUse security hooks.
// A nil SandboxHooks field uses the same defaults as an unset harness security block.
type ClaudeSandboxHooks struct {
	SandboxHooks *harness.SandboxHooks
}

// ClaudeSandboxHooksFromHarness extracts sandbox hook settings from a harness.
func ClaudeSandboxHooksFromHarness(h *harness.Harness) ClaudeSandboxHooks {
	if h == nil || h.Security == nil || h.Security.SandboxHooks == nil {
		return ClaudeSandboxHooks{}
	}
	return ClaudeSandboxHooks{SandboxHooks: h.Security.SandboxHooks}
}

func (c ClaudeSandboxHooks) sandboxHooks() *harness.SandboxHooks {
	return c.SandboxHooks
}
