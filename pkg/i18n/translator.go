// Package i18n provides the locale-aware message translation seam for the
// PliniusCommon submodule. It is the CONST-046-compliant indirection
// between production code in pkg/config, pkg/errors, pkg/grpcclient,
// pkg/types and the locale bundles under bundles/.
//
// Per CONST-051(B), this package is fully decoupled from any consuming
// project: callers inject a Translator (or rely on the NoopTranslator
// default), and the PliniusCommon submodule never reaches into a
// parent project to discover its own catalogue.
//
// Usage:
//
//	t := i18n.NoopTranslator{} // production default — returns key verbatim
//	msg := t.T("pliniuscommon_config_err_service_name_required", nil)
//
// Consuming projects wire a real translator that loads
// bundles/active.en.yaml + locale overrides; the PliniusCommon
// submodule remains project-not-aware.
//
// Scope note (round 135 §11.4 kickoff): PliniusCommon's source surface
// is foundational-utility programmatic — config validation errors,
// gRPC client lifecycle errors, and structured PliniusError values —
// consumed by other Go packages (the 8 sibling Plinius modules + their
// downstream consumers), not displayed directly to end users. The
// bundle below seeds locale-aware companion keys (`pliniuscommon_*`)
// that future UI/CLI/diagnostic consumers can resolve when surfacing
// configuration validation failures, gRPC connection diagnostics, and
// structured error codes to end users — without forcing a breaking
// change to the existing programmatic contracts (sentinel errors,
// ErrorCode values, fmt.Errorf wrappings stay verbatim for protocol
// stability and backward compatibility).
package i18n

// Translator is the message-resolution seam. Implementations MUST be
// safe for concurrent use.
type Translator interface {
	// T returns the localised string for key in the active locale,
	// substituting params by name. When the key is unknown the
	// implementation SHOULD return the key verbatim so production
	// surfaces stay actionable rather than blank.
	T(key string, params map[string]any) string
}

// NoopTranslator is the zero-dependency production-safe default returned
// by package consumers that have not yet wired a project-side
// translator. It returns the key verbatim, which keeps the legacy
// surface shape ("pliniuscommon_config_err_service_name_required")
// rather than an empty string — actionable for downstream string
// assertions and visible in logs.
type NoopTranslator struct{}

// T satisfies Translator by returning the key unchanged. Params are
// ignored on purpose: the noop implementation has no template engine.
func (NoopTranslator) T(key string, _ map[string]any) string {
	return key
}
