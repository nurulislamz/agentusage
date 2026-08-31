#!/usr/bin/env bash
# Create / list / remove multi-account box profiles (agy-box, agent-box, opencode-box).
# Installs the bundled CLI into ~/.local/bin when missing, then runs it.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'EOF'
Usage:
  make box <kind> NAME=<name>
  make box <kind> <name>
  make box-list [kind]
  make box-rm <kind> NAME=<name>

Installs the matching CLI into ~/.local/bin (and PATH) if it is not already there.

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

scripts_dir() {
  printf '%s\n' "${BOX_SCRIPTS_DIR:-$ROOT/scripts/boxes}"
}

install_dir() {
  printf '%s\n' "${BOX_INSTALL_DIR:-$HOME/.local/bin}"
}

ensure_login_path() {
  local dest_dir="$1"
  local line rc
  [ "$dest_dir" = "$HOME/.local/bin" ] || return 0
  line='export PATH="$HOME/.local/bin:$PATH"'
  for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
    [ -f "$rc" ] || continue
    grep -q 'HOME/.local/bin' "$rc" && continue
    printf '\n# agentusage box CLIs\n%s\n' "$line" >> "$rc"
  done
}

ensure_cli() {
  local kind="$1"
  local src dest_dir dest
  src="$(scripts_dir)/$kind"
  [ -f "$src" ] || die "bundled $kind not found in $(scripts_dir)"
  dest_dir="$(install_dir)"
  mkdir -p "$dest_dir" || die "cannot create $dest_dir"
  dest="$dest_dir/$kind"
  if [ ! -f "$dest" ] || ! cmp -s "$src" "$dest"; then
    cp "$src" "$dest"
    chmod 755 "$dest"
    printf 'Installed %s -> %s\n' "$kind" "$dest" >&2
  else
    chmod 755 "$dest"
  fi
  if [ "$kind" = agent-box ]; then
    cp "$dest" "$dest_dir/cursor-box"
    chmod 755 "$dest_dir/cursor-box"
  fi
  case ":$PATH:" in
    *":$dest_dir:"*) ;;
    *)
      export PATH="$dest_dir:$PATH"
      if [ "$dest_dir" = "$HOME/.local/bin" ]; then
        printf 'If %s is not found in this shell, run: export PATH="%s:$PATH"\n' "$kind" "$dest_dir" >&2
      fi
      ;;
  esac
  ensure_login_path "$dest_dir"
  printf '%s\n' "$dest"
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
  cli="$(ensure_cli "$kind")"
  exec "$cli" add "$name"
}

cmd_list() {
  local raw="${1:-}" kind cli
  if [ -z "$raw" ]; then
    local any=0
    for kind in agent-box agy-box opencode-box; do
      if [ -f "$(scripts_dir)/$kind" ]; then
        any=1
        echo "==> $kind"
        cli="$(ensure_cli "$kind")"
        "$cli" list || true
        echo
      fi
    done
    [ "$any" -eq 1 ] || die "bundled box CLIs not found in $(scripts_dir)"
    return 0
  fi
  kind="$(need_kind "$raw")"
  cli="$(ensure_cli "$kind")"
  exec "$cli" list
}

cmd_rm() {
  local kind name cli
  kind="$(need_kind "${1:-}")"
  name="$(box_name "${2:-}")"
  cli="$(ensure_cli "$kind")"
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
