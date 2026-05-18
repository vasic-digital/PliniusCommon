// CONST-035 / Article XI §11.9 unit coverage for the PliniusCommon
// i18n translator seam. Tests assert observable behaviour (string
// equality on returned keys, behaviour invariance across nil/empty
// params) rather than presence of constructors.
package i18n_test

import (
	"sync"
	"testing"

	"digital.vasic.pliniuscommon/pkg/i18n"
)

func TestNoopTranslator_ReturnsKeyVerbatim(t *testing.T) {
	tr := i18n.NoopTranslator{}

	cases := []struct {
		key string
	}{
		{"pliniuscommon_config_err_service_name_required"},
		{"pliniuscommon_config_err_address_required"},
		{"pliniuscommon_config_err_timeout_must_be_positive"},
		{"pliniuscommon_config_err_connection_timeout_must_be_positive"},
		{"pliniuscommon_config_err_max_retries_cannot_be_negative"},
		{"pliniuscommon_config_err_retry_backoff_must_be_positive"},
		{"pliniuscommon_config_err_tls_cert_path_required"},
		{"pliniuscommon_config_err_read_file_failed"},
		{"pliniuscommon_config_err_parse_file_failed"},
		{"pliniuscommon_err_connection"},
		{"pliniuscommon_err_timeout"},
		{"pliniuscommon_err_cancelled"},
		{"pliniuscommon_errcode_unavailable"},
		{"pliniuscommon_errcode_invalid_argument"},
		{"pliniuscommon_grpc_err_already_connected"},
		{"pliniuscommon_grpc_err_not_connected"},
		{"pliniuscommon_grpc_err_dial_failed"},
		{""},
		{"unknown_key_with_dots.and.colons:plus-dashes"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.key, func(t *testing.T) {
			got := tr.T(c.key, nil)
			if got != c.key {
				t.Fatalf("NoopTranslator.T(%q, nil) = %q, want key verbatim", c.key, got)
			}
		})
	}
}

func TestNoopTranslator_IgnoresParams(t *testing.T) {
	tr := i18n.NoopTranslator{}
	key := "pliniuscommon_config_err_read_file_failed"
	params := map[string]any{
		"path":    "/etc/plinius/config.yaml",
		"service": "ouroborous",
	}

	got := tr.T(key, params)
	if got != key {
		t.Fatalf("NoopTranslator.T(%q, params) = %q, want %q (params must be ignored)", key, got, key)
	}
}

func TestNoopTranslator_ConcurrentSafe(t *testing.T) {
	tr := i18n.NoopTranslator{}
	const goroutines = 64
	const iterations = 256

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if got := tr.T("pliniuscommon_grpc_err_dial_failed", nil); got != "pliniuscommon_grpc_err_dial_failed" {
					t.Errorf("concurrent T() returned %q", got)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// fakeTranslator is a unit-test-only test double — permitted under
// CONST-050(A) because this file is *_test.go. It proves the Translator
// interface is satisfiable by a non-Noop implementation and demonstrates
// that consuming projects can wire a real translator that loads
// bundles/active.en.yaml + locale overrides without changing
// PliniusCommon's public surface.
type fakeTranslator struct {
	calls map[string]int
	mu    sync.Mutex
}

func (f *fakeTranslator) T(key string, _ map[string]any) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.calls == nil {
		f.calls = map[string]int{}
	}
	f.calls[key]++
	return "translated:" + key
}

func TestTranslator_InterfaceSatisfaction(t *testing.T) {
	var _ i18n.Translator = i18n.NoopTranslator{}
	var _ i18n.Translator = &fakeTranslator{}
}

func TestFakeTranslator_RecordsCallsAndReturnsRewrittenString(t *testing.T) {
	f := &fakeTranslator{}
	got := f.T("pliniuscommon_config_err_service_name_required", nil)
	if got != "translated:pliniuscommon_config_err_service_name_required" {
		t.Fatalf("fakeTranslator.T returned %q, want translated:<key>", got)
	}
	if f.calls["pliniuscommon_config_err_service_name_required"] != 1 {
		t.Fatalf("fakeTranslator did not record call: %+v", f.calls)
	}
}
