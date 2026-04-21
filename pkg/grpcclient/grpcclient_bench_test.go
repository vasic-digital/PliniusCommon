package grpcclient

import (
	"context"
	"testing"
	"time"

	"digital.vasic.pliniuscommon/pkg/config"
)

func BenchmarkCalculateBackoff(b *testing.B) {
	base := 100 * time.Millisecond
	max := 10 * time.Second
	for i := 0; i < b.N; i++ {
		_ = calculateBackoff(base, max, (i%10)+1)
	}
}

func BenchmarkContextWithMetadata(b *testing.B) {
	cfg := config.New("svc", config.WithAuthToken("tok"), config.WithMetadata("k", "v"))
	c := New(cfg)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_ = c.ContextWithMetadata(ctx)
	}
}
