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
trap 'rm -rf "$WORKDIR" /tmp/box_test_out.$$ /tmp/box_test_err.$$' EXIT

for kind in agent-box agy-box opencode-box; do
  cat >"$WORKDIR/$kind" <<EOF
#!/usr/bin/env bash
printf '%s\n' "\$*" >"$WORKDIR/${kind}.invoked"
EOF
  chmod +x "$WORKDIR/$kind"
done

export PATH="$WORKDIR:$PATH"

# add with positional name
assert_ok "add agent-box positional" "$BOX" add agent-box physics
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "add physics" "agent-box invoked with add physics"

# aliases
assert_ok "add agy alias" "$BOX" add agy chaos
assert_eq "$(cat "$WORKDIR/agy-box.invoked")" "add chaos" "agy-box invoked"

assert_ok "add cursor alias" "$BOX" add cursor-box nurulz
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "add nurulz" "cursor-box maps to agent-box"

assert_ok "add opencode" "$BOX" add opencode-box work
assert_eq "$(cat "$WORKDIR/opencode-box.invoked")" "add work" "opencode-box invoked"

# NAME env
assert_ok "add NAME env" env NAME=fromenv "$BOX" add agent-box
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "add fromenv" "NAME env used when no positional"

# missing name
assert_fail "add without name" "$BOX" add agent-box
grep -q "missing box name" /tmp/box_test_err.$$ || {
  FAIL=$((FAIL + 1))
  echo "FAIL: missing-name error should mention 'missing box name'" >&2
}

# unknown kind
assert_fail "unknown kind" "$BOX" add nope-box foo
grep -qi "unknown" /tmp/box_test_err.$$ || {
  FAIL=$((FAIL + 1))
  echo "FAIL: unknown kind should say unknown" >&2
}

# list
assert_ok "list agent-box" "$BOX" list agent-box
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "list" "list invokes kind list"

# rm
assert_ok "rm agent-box" "$BOX" rm agent-box physics
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "rm physics" "rm invokes kind rm"

# kind as first arg means add
assert_ok "implicit add" "$BOX" agent-box implicit
assert_eq "$(cat "$WORKDIR/agent-box.invoked")" "add implicit" "bare kind is add"

echo
echo "passed=$PASS failed=$FAIL"
[ "$FAIL" -eq 0 ]
