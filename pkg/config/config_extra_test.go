package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateBoundaries covers negative/zero boundary values on numeric fields.
func TestValidateBoundaries(t *testing.T) {
	t.Run("negative max_retries", func(t *testing.T) {
		c := New("svc")
		c.MaxRetries = -1
		assert.Error(t, c.Validate())
	})
	t.Run("zero retry_backoff", func(t *testing.T) {
		c := New("svc")
		c.RetryBackoff = 0
		assert.Error(t, c.Validate())
	})
	t.Run("zero connection_timeout", func(t *testing.T) {
		c := New("svc")
		c.ConnectionTimeout = 0
		assert.Error(t, c.Validate())
	})
}

// TestFromEnvPrecedence verifies that env vars override defaults
// AND that malformed durations/numbers are silently ignored (docs say best-effort parse).
func TestFromEnvPrecedence(t *testing.T) {
	const prefix = "PRECEDENCE_"
	t.Setenv(prefix+"TIMEOUT", "not-a-duration")
	t.Setenv(prefix+"MAX_RETRIES", "also-not-a-number")
	t.Setenv(prefix+"KEEPALIVE_TIME", "3s")
	t.Setenv(prefix+"MAX_RECV_MSG_SIZE", "99")

	cfg := FromEnv("precedence")
	// Bad values must leave defaults intact.
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.MaxRetries)
	// Good values must be applied.
	assert.Equal(t, 3*time.Second, cfg.KeepaliveTime)
	assert.Equal(t, 99, cfg.MaxRecvMsgSize)
}

// TestFromFileMissingFile surfaces a wrapped read error.
func TestFromFileMissingFile(t *testing.T) {
	_, err := FromFile("/no/such/file.yaml", "svc")
	assert.Error(t, err)
}

// TestFromFileInvalidYAML surfaces a wrapped parse error.
func TestFromFileInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(p, []byte("::::not valid yaml\n\t\tbroken"), 0o644))
	_, err := FromFile(p, "svc")
	assert.Error(t, err)
}

// TestFromFileServiceFallback: file without the named service key must still
// parse top-level fields and stamp the service name.
func TestFromFileServiceFallback(t *testing.T) {
	yaml := `address: flat.example.com:5555
timeout: 7s
max_retries: 9
`
	dir := t.TempDir()
	p := filepath.Join(dir, "flat.yaml")
	require.NoError(t, os.WriteFile(p, []byte(yaml), 0o644))
	cfg, err := FromFile(p, "fallback")
	require.NoError(t, err)
	assert.Equal(t, "fallback", cfg.ServiceName)
	assert.Equal(t, "flat.example.com:5555", cfg.Address)
	assert.Equal(t, 7*time.Second, cfg.Timeout)
	assert.Equal(t, 9, cfg.MaxRetries)
}

// TestMetadataNilSafety: WithMetadata on a config whose map was nil'd manually.
func TestMetadataNilSafety(t *testing.T) {
	c := New("svc")
	c.Metadata = nil
	opt := WithMetadata("k", "v")
	opt(c)
	assert.Equal(t, "v", c.Metadata["k"])
}
