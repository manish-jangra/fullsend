package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fullsend-ai/fullsend/internal/sandbox"
)

func TestDefaultRuntime(t *testing.T) {
	rt := Default()
	assert.Equal(t, "claude", rt.Name())
	assert.Equal(t, sandbox.SandboxClaudeConfig, rt.ConfigDir())
	assert.Contains(t, rt.EnvExports()[0], "CLAUDE_CONFIG_DIR")
}

func testRunCommand(agentName, model, repoDir string, pluginDirs []string, debug string) string {
	return buildRunCommand(RunParams{
		AgentBaseName: agentName,
		Model:         model,
		RepoDir:       repoDir,
		PluginDirs:    pluginDirs,
		Debug:         debug,
	})
}

func TestBuildRunCommand_Basic(t *testing.T) {
	cmd := testRunCommand("hello-world", "", "/tmp/workspace/repo", nil, "")
	assert.Contains(t, cmd, "cd /tmp/workspace/repo")
	assert.Contains(t, cmd, "--agent 'hello-world'")
	assert.NotContains(t, cmd, "--model")
	assert.NotContains(t, cmd, "--plugin-dir")
}

func TestBuildRunCommand_WithModel(t *testing.T) {
	cmd := testRunCommand("hello-world", "sonnet", "/tmp/workspace/repo", nil, "")
	assert.Contains(t, cmd, "--model 'sonnet'")
	assert.Contains(t, cmd, "--agent 'hello-world'")
}

func TestBuildRunCommand_EscapesQuotes(t *testing.T) {
	cmd := testRunCommand("test'name", "", "/tmp/workspace/repo", nil, "")
	assert.NotContains(t, cmd, "'test'name'")
	assert.Contains(t, cmd, "'test'\\''name'")
}

func TestBuildRunCommand_WithPluginDirs(t *testing.T) {
	cmd := testRunCommand("agent", "", "/tmp/workspace/repo", []string{"/tmp/claude-config/plugins/gopls-lsp"}, "")
	assert.Contains(t, cmd, "--plugin-dir '/tmp/claude-config/plugins/gopls-lsp'")
}

func TestBuildRunCommand_DebugAll(t *testing.T) {
	cmd := testRunCommand("agent", "", "/tmp/workspace/repo", nil, "*")
	assert.Contains(t, cmd, "--debug-file '/tmp/workspace/claude-debug.log'")
	assert.NotContains(t, cmd, "--debug '")
}

func TestBuildRunCommand_DebugFiltered(t *testing.T) {
	cmd := testRunCommand("agent", "", "/tmp/workspace/repo", nil, "api,hooks")
	assert.Contains(t, cmd, "--debug-file '/tmp/workspace/claude-debug.log'")
	assert.Contains(t, cmd, "--debug 'api,hooks'")
}

func TestBuildPluginConfigs_SinglePlugin(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "gopls-lsp")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.json"),
		[]byte(`{"name":"gopls-lsp"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, ".lsp.json"),
		[]byte(`{"go":{"command":"gopls","args":["serve"]}}`), 0o644))

	entries, err := buildPluginConfigs(
		[]string{pluginDir}, "/tmp/plugins", "/tmp/plugins/marketplaces/claude-plugins-official",
		"claude-plugins-official", "1.0.0", "/tmp/claude-config",
	)
	require.NoError(t, err)
	require.Len(t, entries, 4)

	var mkt map[string]any
	require.NoError(t, json.Unmarshal(entries[0].data, &mkt))
	plugins := mkt["plugins"].([]any)
	require.Len(t, plugins, 1)
	p := plugins[0].(map[string]any)
	assert.Equal(t, "gopls-lsp", p["name"])
	assert.NotNil(t, p["lspServers"])
}

func TestBuildPluginConfigs_EmptyPluginList(t *testing.T) {
	entries, err := buildPluginConfigs(
		nil, "/tmp/plugins", "/tmp/plugins/marketplaces/claude-plugins-official",
		"claude-plugins-official", "1.0.0", "/tmp/claude-config",
	)
	require.NoError(t, err)
	require.Len(t, entries, 4)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(entries[3].data, &settings))
	enabled := settings["enabledPlugins"].(map[string]any)
	assert.Len(t, enabled, 0)
}
