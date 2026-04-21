package grpcclient

import (
	"context"
	"testing"
	"time"

	"digital.vasic.pliniuscommon/pkg/config"
	"digital.vasic.pliniuscommon/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

// TestBackoffBoundsExponentialGrowth asserts backoff grows monotonically
// before it gets capped, and that the ±25% jitter window holds.
func TestBackoffBoundsExponentialGrowth(t *testing.T) {
	base := 100 * time.Millisecond
	max := 10 * time.Second
	prevMin := time.Duration(0)
	for attempt := 1; attempt <= 6; attempt++ {
		d := calculateBackoff(base, max, attempt)
		expectedMin := time.Duration(float64(minDur(base<<uint(attempt-1), max)) * 0.75)
		expectedMax := time.Duration(float64(minDur(base<<uint(attempt-1), max)) * 1.25)
		assert.GreaterOrEqual(t, d, expectedMin, "attempt %d", attempt)
		assert.LessOrEqual(t, d, expectedMax, "attempt %d", attempt)
		assert.GreaterOrEqual(t, d, prevMin, "backoff should be monotonic non-decreasing")
		prevMin = expectedMin
	}
}

func minDur(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// TestBackoffCappedAtMax verifies the max-backoff clamp.
func TestBackoffCappedAtMax(t *testing.T) {
	base := time.Second
	max := 2 * time.Second
	d := calculateBackoff(base, max, 20)
	assert.LessOrEqual(t, d, time.Duration(float64(max)*1.25))
}

// TestContextWithMetadataPropagatesAuth verifies the token lands in outgoing metadata.
func TestContextWithMetadataPropagatesAuth(t *testing.T) {
	cfg := config.New("svc", config.WithAuthToken("tok"), config.WithMetadata("x-trace", "abc"))
	c := New(cfg)
	ctx := c.ContextWithMetadata(context.Background())
	md, ok := metadata.FromOutgoingContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "Bearer tok", md.Get("authorization")[0])
	assert.Equal(t, "abc", md.Get("x-trace")[0])
}

// TestContextWithMetadataWithNoMetadataReturnsOriginal.
func TestContextWithMetadataWithNoMetadataReturnsOriginal(t *testing.T) {
	cfg := config.New("svc")
	cfg.Metadata = nil
	c := New(cfg)
	ctx := c.ContextWithMetadata(context.Background())
	assert.NotNil(t, ctx)
	// No outgoing metadata should have been added.
	_, ok := metadata.FromOutgoingContext(ctx)
	assert.False(t, ok)
}

// TestInvokeUnaryAfterCloseReturnsError: invoking after close hits the closed check.
func TestInvokeUnaryAfterCloseReturnsError(t *testing.T) {
	cfg := config.New("svc", config.WithMaxRetries(0))
	c := New(cfg)
	assert.NoError(t, c.Close())
	err := c.Connect(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.ErrCodeFailedPrecondition))
}

// TestWaitForReadyNotConnected surfaces a precondition error.
func TestWaitForReadyNotConnected(t *testing.T) {
	cfg := config.New("svc")
	c := New(cfg)
	err := c.WaitForReady(context.Background())
	assert.Error(t, err)
}

// TestBuildDialOptionsComposition smoke-tests the option-building path
// under both TLS/insecure and compression branches.
func TestBuildDialOptionsComposition(t *testing.T) {
	cfg := config.New("svc",
		config.WithAddress("x:1"),
		config.WithCompression("gzip"),
	)
	c := New(cfg)
	opts := c.buildDialOptions()
	assert.Greater(t, len(opts), 0)

	cfg2 := config.New("svc",
		config.WithTLS("/c", "/k", "/ca"),
		config.WithInsecureSkipVerify(true),
	)
	c2 := New(cfg2)
	opts2 := c2.buildDialOptions()
	assert.Greater(t, len(opts2), 0)
}
