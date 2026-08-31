#!/usr/bin/env bash
# sshutil installer: builds the binary and links it into ~/.local/bin
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${HOME}/.local/bin"

log()  { printf '\033[1;35m==>\033[0m %s\n' "$*"; }
fail() { printf '\033[1;31merror:\033[0m %s\n' "$*" >&2; exit 1; }

command -v git >/dev/null 2>&1 || fail "git is required"

GO_MIN="1.27"

need_go() {
  local ver
  if ! command -v go >/dev/null 2>&1; then
    return 0
  fi
  ver="$(go env GOVERSION 2>/dev/null | sed 's/^go//')" || return 0
  case "${ver}" in
    devel*) return 0 ;;
  esac
  [ "$(printf '%s\n' "${GO_MIN}" "${ver}" | sort -V | head -n1)" != "${GO_MIN}" ]
}

if need_go; then
  if command -v go >/dev/null 2>&1; then
    log "Go $(go env GOVERSION) is too old (need ${GO_MIN}+) — installing official toolchain"
  else
    log "Installing Go ${GO_MIN} toolchain"
  fi
  command -v curl >/dev/null 2>&1 || fail "curl is required to install Go ${GO_MIN}+"
  case "$(uname -m)" in
    x86_64) GO_ARCH=amd64 ;;
    aarch64 | arm64) GO_ARCH=arm64 ;;
    *) fail "unsupported architecture: $(uname -m) (need amd64 or arm64)" ;;
  esac
  sudo rm -rf /usr/local/go
  curl -fsSL "https://go.dev/dl/go${GO_MIN}.0.linux-${GO_ARCH}.tar.gz" | sudo tar -C /usr/local -xz
  export PATH="/usr/local/go/bin:${PATH}"
fi

command -v go >/dev/null 2>&1 || fail "Go install failed"

log "Building sshutil"
mkdir -p "${BIN_DIR}"
( cd "${REPO_DIR}" && go build -o "${BIN_DIR}/sshutil" . )

case ":${PATH}:" in
  *":${BIN_DIR}:"*) ;;
  *)
    log "NOTE: ${BIN_DIR} is not in your PATH"
    log "Add this to ~/.bashrc:  export PATH=\"${BIN_DIR}:\$PATH\""
    ;;
esac

log "Done — run: sshutil"
