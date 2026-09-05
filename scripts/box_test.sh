#!/usr/bin/env bash
# Tests for scripts/box.sh. Run from repo root: ./scripts/box_test.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BOX="$ROOT/scripts/box.sh"
PASS=0
FAIL=0

assert_eq() {
  local got="$1" want="$2" msg="$3"
  if [ "$got" = "$want" ]; then
    PASS=$((PASS + 1))
    return 0
  fi
  FAIL=$((FAIL + 1))
  printf 'FAIL: %s\n  got:  %q\n  want: %q\n' "$msg" "$got" "$want" >&2
}

assert_fail() {
  local msg="$1"
  shift
  if "$@" >/tmp/box_test_out.$$ 2>/tmp/box_test_err.$$; then
    FAIL=$((FAIL + 1))
    printf 'FAIL: %s (expected non-zero)\n' "$msg" >&2
    return 0
  fi
  PASS=$((PASS + 1))
}

assert_ok() {
  local msg="$1"
  shift
  if "$@"; then
    PASS=$((PASS + 1))
    return 0
  fi
  FAIL=$((FAIL + 1))
  printf 'FAIL: %s\n' "$msg" >&2
}

if [ ! -x "$BOX" ]; then
  echo "FAIL: $BOX is missing or not executable" >&2
  exit 1
fi

WORKDIR="$(mktemp -d)"
SCRIPTS="$WORKDIR/scripts"
BINDST="$WORKDIR/bin"
trap 'rm -rf "$WORKDIR" /tmp/box_test_out.$$ /tmp/box_test_err.$$' EXIT

mkdir -p "$SCRIPTS" "$BINDST"
for kind in agent-box agy-box opencode-box; do
  cat >"$SCRIPTS/$kind" <<EOF
#!/usr/bin/env bash
printf '%s\n' "\$*" >"$WORKDIR/${kind}.invoked"
printf '%s\n' "\$0" >"$WORKDIR/${kind}.exe"
EOF
  chmod +x "$SCRIPTS/$kind"
done

# Isolated PATH: make box must install the bundled CLI, not require it already on PATH.
export PATH="/usr/bin:/bin"
export BOX_SCRIPTS_DIR="$SCRIPTS"
export BOX_INSTALL_DIR="$BINDST"
unset NAME || true

run_box() {
  env -u NAME PATH="$PATH" BOX_SCRIPTS_DIR="$BOX_SCRIPTS_DIR" BOX_INSTALL_DIR="$BOX_INSTALL_DIR" "$BOX" "$@"
}

# add with positional name — installs CLI then invokes it
assert_ok "add agent-box positional" run_box add agent-box physics
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "add physics" "agent-box invoked with add physics"
assert_ok "installed agent-box to BOX_INSTALL_DIR" test -x "$BINDST/agent-box"
assert_eq "$(cat "$WORKDIR/agent-box.exe")" "$BINDST/agent-box" "invoked the installed copy, not PATH"

# aliases
assert_ok "add agy alias" run_box add agy chaos
assert_eq "$(cat "$WORKDIR/agy-box.invoked")" "add chaos" "agy-box invoked"
assert_ok "installed agy-box" test -x "$BINDST/agy-box"

assert_ok "add cursor alias" run_box add cursor-box nurulz
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "add nurulz" "cursor-box maps to agent-box"

assert_ok "add opencode" run_box add opencode-box work
assert_eq "$(cat "$WORKDIR/opencode-box.invoked")" "add work" "opencode-box invoked"

# NAME env
assert_ok "add NAME env" env PATH="$PATH" BOX_SCRIPTS_DIR="$SCRIPTS" BOX_INSTALL_DIR="$BINDST" NAME=fromenv "$BOX" add agent-box
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "add fromenv" "NAME env used when no positional"

# missing name
assert_fail "add without name" run_box add agent-box
grep -q "missing box name" /tmp/box_test_err.$$ || {
  FAIL=$((FAIL + 1))
  echo "FAIL: missing-name error should mention 'missing box name'" >&2
}

# unknown kind
assert_fail "unknown kind" run_box add nope-box foo
grep -qi "unknown" /tmp/box_test_err.$$ || {
  FAIL=$((FAIL + 1))
  echo "FAIL: unknown kind should say unknown" >&2
}

# missing bundled CLI
assert_fail "missing bundled cli" env PATH="/usr/bin:/bin" BOX_SCRIPTS_DIR="$WORKDIR/empty" BOX_INSTALL_DIR="$BINDST" "$BOX" add agent-box physics
grep -qi "bundled" /tmp/box_test_err.$$ || {
  FAIL=$((FAIL + 1))
  echo "FAIL: missing bundled CLI should mention bundled" >&2
}

# list
assert_ok "list agent-box" run_box list agent-box
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "list" "list invokes kind list"

# rm
assert_ok "rm agent-box" run_box rm agent-box physics
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "rm physics" "rm invokes kind rm"

# kind as first arg means add
assert_ok "implicit add" run_box agent-box implicit
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "add implicit" "bare kind is add"

# persist ~/.local/bin on PATH in shell rc when installing to the default location
RC_HOME="$WORKDIR/home"
mkdir -p "$RC_HOME"
touch "$RC_HOME/.bashrc"
assert_ok "install updates bashrc" \
  env HOME="$RC_HOME" PATH="/usr/bin:/bin" BOX_SCRIPTS_DIR="$SCRIPTS" BOX_INSTALL_DIR="$RC_HOME/.local/bin" \
  "$BOX" add agent-box rcpath
assert_ok "bashrc mentions .local/bin" grep -q 'HOME/.local/bin' "$RC_HOME/.bashrc"
# --- Real box script shell fixture tests (fake tools, temporary HOME) ---
BOX_HOME="$WORKDIR/box_home"
FAKE_BIN="$WORKDIR/fake_bin"
mkdir -p "$BOX_HOME" "$FAKE_BIN"

cat >"$FAKE_BIN/bwrap" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >"$BOX_TEST_LOG"
EOF
chmod +x "$FAKE_BIN/bwrap"

cat >"$FAKE_BIN/agy" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
chmod +x "$FAKE_BIN/agy"

cat >"$FAKE_BIN/agent" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
chmod +x "$FAKE_BIN/agent"

export BOX_TEST_LOG="$WORKDIR/bwrap.log"
REAL_AGY_BOX="$ROOT/scripts/boxes/agy-box"
REAL_AGENT_BOX="$ROOT/scripts/boxes/agent-box"

# 1. Shell syntax check
assert_ok "agy-box syntax" bash -n "$REAL_AGY_BOX"
assert_ok "agent-box syntax" bash -n "$REAL_AGENT_BOX"

# 2. agy-box add does not inject statusline
assert_ok "agy-box add devbox" env HOME="$BOX_HOME" PATH="$FAKE_BIN:$PATH" BWRAP_BIN="$FAKE_BIN/bwrap" "$REAL_AGY_BOX" add devbox
AGY_SETTINGS="$BOX_HOME/.agy-containers/devbox/.gemini/antigravity-cli/settings.json"
if [ -f "$AGY_SETTINGS" ]; then
  assert_fail "agy-box add does not create statusLine" grep -q '"statusLine"' "$AGY_SETTINGS"
else
  PASS=$((PASS + 1))
fi

# 3. agy-box launch preserves custom settings and does not inject statusline
mkdir -p "$(dirname "$AGY_SETTINGS")"
cat >"$AGY_SETTINGS" <<'EOF'
{
  "customKey": "customVal"
}
EOF
assert_ok "agy-box launch devbox" env HOME="$BOX_HOME" PATH="$FAKE_BIN:$PATH" BWRAP_BIN="$FAKE_BIN/bwrap" "$REAL_AGY_BOX" devbox
assert_ok "agy-box invoked fake bwrap" test -s "$BOX_TEST_LOG"
assert_fail "agy-box launch did not inject statusLine" grep -q '"statusLine"' "$AGY_SETTINGS"
assert_ok "agy-box launch preserved customKey" grep -q '"customVal"' "$AGY_SETTINGS"

# 4. agent-box add does not inject statusline
assert_ok "agent-box add devbox" env HOME="$BOX_HOME" PATH="$FAKE_BIN:$PATH" BWRAP_BIN="$FAKE_BIN/bwrap" "$REAL_AGENT_BOX" add devbox
CURSOR_CONFIG="$BOX_HOME/.agent-containers/devbox/.cursor/cli-config.json"
if [ -f "$CURSOR_CONFIG" ]; then
  assert_fail "agent-box add does not create statusLine" grep -q '"statusLine"' "$CURSOR_CONFIG"
else
  PASS=$((PASS + 1))
fi

# 5. agent-box launch preserves custom statusLine command and does not overwrite
mkdir -p "$(dirname "$CURSOR_CONFIG")"
cat >"$CURSOR_CONFIG" <<'EOF'
{
  "statusLine": {
    "command": "custom-status-cmd"
  }
}
EOF
: >"$BOX_TEST_LOG"
assert_ok "agent-box launch devbox" env HOME="$BOX_HOME" PATH="$FAKE_BIN:$PATH" BWRAP_BIN="$FAKE_BIN/bwrap" "$REAL_AGENT_BOX" devbox
assert_ok "agent-box invoked fake bwrap" test -s "$BOX_TEST_LOG"
assert_ok "agent-box launch preserved custom statusLine command" grep -q 'custom-status-cmd' "$CURSOR_CONFIG"
assert_fail "agent-box launch did not inject openusage statusline" grep -q 'openusage cursor statusline' "$CURSOR_CONFIG"

echo
echo "passed=$PASS failed=$FAIL"
[ "$FAIL" -eq 0 ]
