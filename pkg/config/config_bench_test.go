package config

import "testing"

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New("svc")
	}
}

func BenchmarkFromEnv(b *testing.B) {
	b.Setenv("BENCHSRV_ADDRESS", "x:1")
	b.Setenv("BENCHSRV_TIMEOUT", "1s")
	for i := 0; i < b.N; i++ {
		_ = FromEnv("benchsrv")
	}
}
