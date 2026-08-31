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

run_box() {
  env PATH="$PATH" BOX_SCRIPTS_DIR="$BOX_SCRIPTS_DIR" BOX_INSTALL_DIR="$BOX_INSTALL_DIR" "$BOX" "$@"
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
assert_ok "agent-box landed in ~/.local/bin" test -x "$RC_HOME/.local/bin/agent-box"

echo
echo "passed=$PASS failed=$FAIL"
[ "$FAIL" -eq 0 ]
