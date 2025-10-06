package cli

import (
	"testing"

	"github.com/pezware/samedi.dev/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetConfigValue_String(t *testing.T) {
	cfg := config.DefaultConfig()

	require.NoError(t, setConfigValue(cfg, "user.email", "user@example.com"))
	assert.Equal(t, "user@example.com", cfg.User.Email)
}

func TestSetConfigValue_Bool(t *testing.T) {
	cfg := config.DefaultConfig()

	require.NoError(t, setConfigValue(cfg, "storage.backup_enabled", "false"))
	assert.False(t, cfg.Storage.BackupEnabled)
}

func TestSetConfigValue_Int(t *testing.T) {
	cfg := config.DefaultConfig()

	require.NoError(t, setConfigValue(cfg, "llm.timeout_seconds", "180"))
	assert.Equal(t, 180, cfg.LLM.TimeoutSeconds)
}

func TestSetConfigValue_InvalidBool(t *testing.T) {
	cfg := config.DefaultConfig()

	err := setConfigValue(cfg, "storage.backup_enabled", "notabool")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid boolean")
}

func TestSetConfigValue_InvalidInt(t *testing.T) {
	cfg := config.DefaultConfig()

	err := setConfigValue(cfg, "llm.timeout_seconds", "abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid integer")
}

func TestSetConfigValue_UnknownKey(t *testing.T) {
	cfg := config.DefaultConfig()

	err := setConfigValue(cfg, "unknown.key", "value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestGetConfigValue(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Sync.Enabled = true

	value := getConfigValue(cfg, "sync.enabled")
	assert.Equal(t, true, value)

	missing := getConfigValue(cfg, "does.not.exist")
	assert.Nil(t, missing)
}
