#!/usr/bin/env bash
# Create / list / remove multi-account box profiles (agy-box, agent-box, opencode-box).
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  make box <kind> NAME=<name>
  make box <kind> <name>
  make box-list [kind]
  make box-rm <kind> NAME=<name>

Kinds:
  agent-box     Cursor agent boxes (~/.agent-containers)
                aliases: agent, cursor-box, cursor
  agy-box       Antigravity boxes (~/.agy-containers)
                aliases: agy, antigravity
  opencode-box  OpenCode boxes (~/.opencode-containers)
                aliases: opencode
EOF
}

die() {
  printf 'Error: %s\n' "$*" >&2
  exit 1
}

resolve_kind() {
  local raw="${1:-}"
  raw="$(printf '%s' "$raw" | tr '[:upper:]' '[:lower:]')"
  case "$raw" in
    agent-box | agent | cursor-box | cursor)
      printf '%s\n' agent-box
      ;;
    agy-box | agy | antigravity)
      printf '%s\n' agy-box
      ;;
    opencode-box | opencode)
      printf '%s\n' opencode-box
      ;;
    *)
      return 1
      ;;
  esac
}

need_kind() {
  local kind
  kind="$(resolve_kind "${1:-}")" || die "unknown box kind '${1:-}' (try agent-box, agy-box, opencode-box)"
  printf '%s\n' "$kind"
}

need_cli() {
  local kind="$1"
  command -v "$kind" >/dev/null 2>&1 || die "$kind not found on PATH. Install it, then retry."
  command -v "$kind"
}

box_name() {
  local positional="${1:-}"
  local name="${positional:-${NAME:-}}"
  [ -n "$name" ] || die "missing box name (make box <kind> NAME=<name>)"
  printf '%s\n' "$name"
}

cmd_add() {
  local kind name cli
  kind="$(need_kind "${1:-}")"
  name="$(box_name "${2:-}")"
  cli="$(need_cli "$kind")"
  exec "$cli" add "$name"
}

cmd_list() {
  local raw="${1:-}" kind cli
  if [ -z "$raw" ]; then
    local any=0
    for kind in agent-box agy-box opencode-box; do
      if command -v "$kind" >/dev/null 2>&1; then
        any=1
        echo "==> $kind"
        "$kind" list || true
        echo
      fi
    done
    [ "$any" -eq 1 ] || die "no box CLIs found on PATH (agent-box, agy-box, opencode-box)"
    return 0
  fi
  kind="$(need_kind "$raw")"
  cli="$(need_cli "$kind")"
  exec "$cli" list
}

cmd_rm() {
  local kind name cli
  kind="$(need_kind "${1:-}")"
  name="$(box_name "${2:-}")"
  cli="$(need_cli "$kind")"
  exec "$cli" rm "$name"
}

main() {
  local cmd="${1:-}"
  if [ $# -gt 0 ]; then
    shift
  fi
  case "$cmd" in
    "" | -h | --help | help)
      usage
      exit 0
      ;;
    add)
      cmd_add "$@"
      ;;
    list | ls)
      cmd_list "$@"
      ;;
    rm | remove | delete)
      cmd_rm "$@"
      ;;
    *)
      if resolve_kind "$cmd" >/dev/null; then
        cmd_add "$cmd" "$@"
        return
      fi
      die "unknown command '$cmd'"
      ;;
  esac
}

main "$@"
