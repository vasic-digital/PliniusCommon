package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAllErrorCodesClassified exercises the full 16-code matrix to ensure
// the retry classification is complete (no panic, no default-case drift).
func TestAllErrorCodesClassified(t *testing.T) {
	codes := []ErrorCode{
		ErrCodeUnavailable, ErrCodeInvalidArgument, ErrCodeNotFound, ErrCodeAlreadyExists,
		ErrCodePermissionDenied, ErrCodeUnauthenticated, ErrCodeResourceExhausted,
		ErrCodeFailedPrecondition, ErrCodeAborted, ErrCodeOutOfRange, ErrCodeUnimplemented,
		ErrCodeInternal, ErrCodeUnknown, ErrCodeTimeout, ErrCodeCancelled, ErrCodeConnection,
	}
	seen := make(map[ErrorCode]bool)
	for _, c := range codes {
		assert.False(t, seen[c], "duplicate code %s", c)
		seen[c] = true
		e := New(c, "svc", "m")
		// No panic. Retryable decision must be deterministic.
		_ = e.IsRetryable()
		assert.Equal(t, c, e.Code)
	}
	assert.Len(t, seen, 16)
}

// TestDeepUnwrapChain verifies errors.Is traverses multi-level wraps.
func TestDeepUnwrapChain(t *testing.T) {
	root := stderrors.New("root")
	mid := fmt.Errorf("mid: %w", root)
	pe := Wrap(ErrCodeInternal, "svc", "top", mid)

	// pe wraps mid which wraps root.
	assert.True(t, stderrors.Is(pe, root))
	assert.True(t, stderrors.Is(pe, mid))

	// errors.As walks to the PliniusError.
	var found *PliniusError
	assert.True(t, stderrors.As(pe, &found))
	assert.Equal(t, ErrCodeInternal, found.Code)
}

// TestWithDetailInitialisation verifies Details map lazy-init doesn't
// clobber an existing map.
func TestWithDetailInitialisation(t *testing.T) {
	e := New(ErrCodeInternal, "svc", "m")
	assert.Nil(t, e.Details)
	e.WithDetail("a", 1)
	assert.NotNil(t, e.Details)
	e.WithDetail("b", "two")
	assert.Len(t, e.Details, 2)
}

// TestIsOnNonPliniusError covers negative path for Is.
func TestIsOnNonPliniusError(t *testing.T) {
	plain := stderrors.New("plain")
	assert.False(t, Is(plain, ErrCodeInternal))
	assert.False(t, Is(nil, ErrCodeInternal))
}

// TestIsRetryableOnNil and plain errors returns false.
func TestIsRetryableOnNil(t *testing.T) {
	assert.False(t, IsRetryableError(nil))
	assert.False(t, IsRetryableError(stderrors.New("plain")))
}

// TestWrapNilCause ensures Wrap handles nil cause gracefully.
func TestWrapNilCause(t *testing.T) {
	e := Wrap(ErrCodeInternal, "svc", "m", nil)
	assert.Nil(t, e.Unwrap())
	// Error string should NOT contain a trailing ": <nil>" from the cause path.
	assert.NotContains(t, e.Error(), "<nil>")
}

// TestErrorFormattingShape covers both with- and without-cause format paths.
func TestErrorFormattingShape(t *testing.T) {
	base := New(ErrCodeNotFound, "svc", "missing")
	assert.True(t, strings.HasPrefix(base.Error(), "[svc:NOT_FOUND]"))

	caused := base.WithCause(stderrors.New("root"))
	assert.Contains(t, caused.Error(), ": root")
}

// TestNewfArgumentForwarding covers format-arg application.
func TestNewfArgumentForwarding(t *testing.T) {
	e := Newf(ErrCodeInvalidArgument, "svc", "bad %s at %d", "x", 42)
	assert.Equal(t, "bad x at 42", e.Message)
}

// TestSentinelUniqueness ensures the package-level sentinels are distinct.
func TestSentinelUniqueness(t *testing.T) {
	assert.NotEqual(t, ErrConnection, ErrTimeout)
	assert.NotEqual(t, ErrConnection, ErrCancelled)
	assert.NotEqual(t, ErrTimeout, ErrCancelled)
}
