#!/bin/sh
# Unit tests for install.sh.
#
# Usage: sh test/install_test.sh
#
# The script under test is sourced with INSTALL_SH_TEST=1 so its functions can
# be called without running the installer. Each test runs in a subshell because
# install.sh enables `set -e` and its failure paths call exit. External
# commands (curl, uname) are shadowed with shell functions where needed.
#
# Set INSTALL_SH to test a different copy of the script (used to verify that
# these tests catch known-bad versions).

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_DIR=$(dirname "$TEST_DIR")
INSTALL_SH="${INSTALL_SH:-$REPO_DIR/install.sh}"
FIXTURE="$TEST_DIR/fixtures/github_release.json"

PASS=0
FAIL=0

pass() { PASS=$((PASS + 1)); printf 'ok - %s\n' "$1"; }
fail() { FAIL=$((FAIL + 1)); printf 'FAIL - %s\n' "$1"; }

assert_eq() {
  # $1 = test name, $2 = expected, $3 = actual
  if [ "$2" = "$3" ]; then
    pass "$1"
  else
    fail "$1 (expected '$2', got '$3')"
  fi
}

assert_contains() {
  # $1 = test name, $2 = needle, $3 = haystack
  case "$3" in
    *"$2"*) pass "$1" ;;
    *)      fail "$1 (output does not contain '$2')" ;;
  esac
}

[ -f "$INSTALL_SH" ] || { echo "install.sh not found at $INSTALL_SH" >&2; exit 1; }
[ -f "$FIXTURE" ] || { echo "fixture not found at $FIXTURE" >&2; exit 1; }

# ── Syntax check ─────────────────────────────────────────────────────────────

if sh -n "$INSTALL_SH" 2>/dev/null; then
  pass "install.sh parses (sh -n)"
else
  fail "install.sh parses (sh -n)"
fi

# ── get_version: parse latest release from GitHub API ────────────────────────
# Regression test for issue #3: the sed pattern must work on BSD sed (macOS)
# as well as GNU sed (Linux).

out=$(
  INSTALL_SH_TEST=1
  . "$INSTALL_SH"
  curl() { cat "$FIXTURE"; }
  get_version >/dev/null 2>&1
  printf '%s|%s' "$VERSION" "$VERSION_NUM"
) || out="(subshell failed: $out)"
assert_eq "get_version parses tag_name from API response" "v0.1.1|0.1.1" "$out"

# ── get_version: explicit version argument ───────────────────────────────────

out=$(
  INSTALL_SH_TEST=1
  . "$INSTALL_SH"
  get_version "v1.2.3" >/dev/null 2>&1
  printf '%s|%s' "$VERSION" "$VERSION_NUM"
) || out="(subshell failed: $out)"
assert_eq "get_version strips leading v from explicit version" "v1.2.3|1.2.3" "$out"

# ── detect_platform: OS/arch mapping ─────────────────────────────────────────

out=$(
  INSTALL_SH_TEST=1
  . "$INSTALL_SH"
  uname() { case "$1" in (-s) echo "Linux" ;; (-m) echo "x86_64" ;; esac; }
  detect_platform >/dev/null 2>&1
  printf '%s|%s' "$OS" "$ARCH"
) || out="(subshell failed: $out)"
assert_eq "detect_platform maps Linux/x86_64" "linux|x64" "$out"

out=$(
  INSTALL_SH_TEST=1
  . "$INSTALL_SH"
  uname() { case "$1" in (-s) echo "Darwin" ;; (-m) echo "arm64" ;; esac; }
  detect_platform >/dev/null 2>&1
  printf '%s|%s' "$OS" "$ARCH"
) || out="(subshell failed: $out)"
assert_eq "detect_platform maps Darwin/arm64" "macos|arm64" "$out"

rc=0
(
  INSTALL_SH_TEST=1
  . "$INSTALL_SH"
  uname() { echo "FreeBSD"; }
  detect_platform
) >/dev/null 2>&1 || rc=$?
assert_eq "detect_platform rejects unsupported OS with exit 1" "1" "$rc"

# ── has_chrome: browser detection ────────────────────────────────────────────

STUB_DIR=$(mktemp -d)
trap 'rm -rf "$STUB_DIR"' EXIT

# Negative case only works where the hardcoded macOS app bundles are absent
# (they exist on developer Macs and on GitHub's macOS runners).
if [ -d "/Applications/Google Chrome.app" ] || [ -d "/Applications/Chromium.app" ]; then
  echo "skip - has_chrome returns 1 when no browser present (browser app installed on this machine)"
else
  rc=0
  (
    INSTALL_SH_TEST=1
    . "$INSTALL_SH"
    PATH="$STUB_DIR"
    has_chrome
  ) >/dev/null 2>&1 || rc=$?
  assert_eq "has_chrome returns 1 when no browser present" "1" "$rc"
fi

printf '#!/bin/sh\n' > "$STUB_DIR/chromium"
chmod +x "$STUB_DIR/chromium"
rc=0
(
  INSTALL_SH_TEST=1
  . "$INSTALL_SH"
  PATH="$STUB_DIR"
  has_chrome
) >/dev/null 2>&1 || rc=$?
assert_eq "has_chrome finds chromium on PATH" "0" "$rc"

# ── browser_required: message and exit code ──────────────────────────────────

rc=0
out=$(
  INSTALL_SH_TEST=1
  . "$INSTALL_SH"
  browser_required "Test reason line." 2>&1
) || rc=$?
assert_eq "browser_required exits 1" "1" "$rc"
assert_contains "browser_required states the requirement" "requires Chrome or Chromium" "$out"
assert_contains "browser_required echoes the reason" "Test reason line." "$out"
assert_contains "browser_required says to re-run the installer" "run this installer again" "$out"

# ── Summary ──────────────────────────────────────────────────────────────────

echo ""
echo "passed: $PASS  failed: $FAIL"
[ "$FAIL" -eq 0 ] || exit 1
