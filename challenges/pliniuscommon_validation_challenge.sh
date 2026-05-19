#!/usr/bin/env bash
# pliniuscommon_validation_challenge.sh
#
# Round 213 §11.4 deliverable per the verbatim 2026-05-19 operator
# mandate (cascaded under CONST-049 §11.4.17):
#
#   "all existing tests and Challenges do work in anti-bluff manner —
#    they MUST confirm that all tested codebase really works as expected!
#    We had been in position that all tests do execute with success and
#    all Challenges as well, but in reality the most of the features
#    does not work and can't be used! This MUST NOT be the case and
#    execution of tests and Challenges MUST guarantee the quality, the
#    completition and full usability by end users of the product!"
#
# Six gates — each fails LOUDLY on regression. The paired-mutation gate
# (§6) closes CONST-050(B)'s anti-bluff invariant by proving the
# integrity check actually catches a corruption.

set -u
# Note: we intentionally do NOT use `set -e` — gate failures are
# tallied so the script reports every failure rather than stopping at
# the first.

readonly REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly BUNDLE="${REPO_ROOT}/pkg/i18n/bundles/active.en.yaml"
readonly EXPECTED_KEYS=36

FAILED=0
PASSED=0

log()    { printf '%s\n' "$*"; }
header() { printf '\n=== %s ===\n' "$*"; }
ok()     { log "PASS|$*"; PASSED=$((PASSED+1)); }
fail()   { log "FAIL|$*"; FAILED=$((FAILED+1)); }

cd "${REPO_ROOT}" || { fail "cannot cd into ${REPO_ROOT}"; exit 1; }

###############################################################################
header "Gate 1/6 — go build ./... compiles"
###############################################################################
if GOMAXPROCS=2 nice -n 19 go build ./... >/tmp/plinius_build.log 2>&1; then
  ok "go build ./... exits 0"
else
  fail "go build ./... non-zero (see /tmp/plinius_build.log)"
  sed -n '1,40p' /tmp/plinius_build.log
fi

###############################################################################
header "Gate 2/6 — go test ./pkg/... -race passes"
###############################################################################
if GOMAXPROCS=2 nice -n 19 go test -count=1 -p 1 -race ./pkg/... \
     >/tmp/plinius_test.log 2>&1; then
  ok "go test -race ./pkg/... all-green"
  grep -E '^(ok|FAIL)' /tmp/plinius_test.log || true
else
  fail "go test -race ./pkg/... non-zero (see /tmp/plinius_test.log)"
  sed -n '1,60p' /tmp/plinius_test.log
fi

###############################################################################
header "Gate 3/6 — bundle file present + key count >= ${EXPECTED_KEYS}"
###############################################################################
if [ ! -f "${BUNDLE}" ]; then
  fail "i18n bundle missing at ${BUNDLE}"
else
  ACTUAL_KEYS=$(grep -c '^pliniuscommon_' "${BUNDLE}" || echo 0)
  if [ "${ACTUAL_KEYS}" -ge "${EXPECTED_KEYS}" ]; then
    ok "bundle key count = ${ACTUAL_KEYS} (>= ${EXPECTED_KEYS})"
  else
    fail "bundle key count = ${ACTUAL_KEYS} (< ${EXPECTED_KEYS}) — CONST-046 regression"
  fi
fi

###############################################################################
header "Gate 4/6 — config validator surface complete (7 invariants)"
###############################################################################
# Each invariant has a paired bundle key (round-135 mapping); assert all
# 7 keys exist verbatim.
CONFIG_KEYS=(
  pliniuscommon_config_err_service_name_required
  pliniuscommon_config_err_address_required
  pliniuscommon_config_err_timeout_must_be_positive
  pliniuscommon_config_err_connection_timeout_must_be_positive
  pliniuscommon_config_err_max_retries_cannot_be_negative
  pliniuscommon_config_err_retry_backoff_must_be_positive
  pliniuscommon_config_err_tls_cert_path_required
)
MISSING=0
for k in "${CONFIG_KEYS[@]}"; do
  if ! grep -q "^${k}:" "${BUNDLE}"; then
    fail "config validator bundle key missing: ${k}"
    MISSING=$((MISSING+1))
  fi
done
[ "${MISSING}" -eq 0 ] && ok "all 7 config validator bundle keys present"

###############################################################################
header "Gate 5/6 — grpcclient lifecycle diagnostic keys present (5 keys)"
###############################################################################
GRPC_KEYS=(
  pliniuscommon_grpc_err_already_connected
  pliniuscommon_grpc_err_not_connected
  pliniuscommon_grpc_err_dial_failed
  pliniuscommon_grpc_err_close_failed
  pliniuscommon_grpc_err_invocation_failed
)
MISSING=0
for k in "${GRPC_KEYS[@]}"; do
  if ! grep -q "^${k}:" "${BUNDLE}"; then
    fail "grpcclient bundle key missing: ${k}"
    MISSING=$((MISSING+1))
  fi
done
[ "${MISSING}" -eq 0 ] && ok "all 5 grpcclient bundle keys present"

###############################################################################
header "Gate 6/6 — PAIRED MUTATION (CONST-050(B) anti-bluff)"
###############################################################################
# Plant a known corruption, re-run the integrity check, assert it
# DETECTS the corruption. Without this gate the green PASS in §3-§5
# could be a structural bluff (file exists but check is shallow).
TMP_BUNDLE="$(mktemp)"
cp "${BUNDLE}" "${TMP_BUNDLE}"
trap 'rm -f "${TMP_BUNDLE}" "${BUNDLE}.mutant" 2>/dev/null || true' EXIT

# Mutate: delete the FIRST `pliniuscommon_` key line into a separate file
cp "${BUNDLE}" "${BUNDLE}.mutant"
# remove ONLY the first matching line to drop exactly one key
sed -i '0,/^pliniuscommon_/{//d;}' "${BUNDLE}.mutant"

MUTATED_KEYS=$(grep -c '^pliniuscommon_' "${BUNDLE}.mutant" || echo 0)
if [ "${MUTATED_KEYS}" -lt "${EXPECTED_KEYS}" ]; then
  ok "mutation drops bundle key count to ${MUTATED_KEYS} (< ${EXPECTED_KEYS}) — gate 3 would FAIL on this state"
else
  fail "mutation failed to reduce key count (still ${MUTATED_KEYS}) — challenge integrity broken"
fi

# Negative confirmation: simulate gate 3 on the mutant file and assert
# it reports failure (i.e. the integrity check is REAL, not a bluff).
if [ "${MUTATED_KEYS}" -ge "${EXPECTED_KEYS}" ]; then
  fail "anti-bluff mutation gate FAILED — integrity check did not catch corruption"
else
  ok "anti-bluff mutation gate proved integrity check catches regression"
fi

# Restore + final paranoia check the bundle is byte-identical
if cmp -s "${BUNDLE}" "${TMP_BUNDLE}"; then
  ok "post-mutation bundle integrity preserved (byte-identical to original)"
else
  fail "post-mutation bundle DIFFERS from original — file was corrupted"
fi

###############################################################################
header "Summary"
###############################################################################
log "PASSED gates: ${PASSED}"
log "FAILED gates: ${FAILED}"

if [ "${FAILED}" -gt 0 ]; then
  log "RESULT|FAILED"
  exit 1
fi
log "RESULT|PASSED"
exit 0
