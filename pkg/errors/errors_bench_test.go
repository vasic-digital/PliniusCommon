package errors

import (
	"fmt"
	"testing"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(ErrCodeUnavailable, "svc", "msg")
	}
}

func BenchmarkIsRetryableError(b *testing.B) {
	e := Wrap(ErrCodeInternal, "svc", "top", fmt.Errorf("root"))
	for i := 0; i < b.N; i++ {
		_ = IsRetryableError(e)
	}
}
