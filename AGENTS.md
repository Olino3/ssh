# AGENTS.md — repository conventions for coding agents

## Hard rules

- **NEVER commit.** The owner does not want automated commits. Do not run
  `git add`, `git commit`, or `git push`.
- `docs/superpowers/` (specs and plans) is gitignored and must stay
  uncommitted.
- Never log, echo, or persist private key material or Tailscale pre-auth keys.

## Layout

- `main.go` + `cmd/` — cobra commands (root/init wizard, add-key, sync, status)
- `internal/ui` — all output styling (lipgloss + charmbracelet/log). Commands
  must not print raw ANSI themselves; use `ui.*`.
- `internal/system` — Runner interface (ExecRunner/MockRunner), WriteRootFile,
  OS detection. All exec calls go through a Runner.
- `internal/state` — yaml state at `~/.config/sshutil/state.yaml`
- `internal/sshkeys` — keygen manager + ssh config generation
- `internal/hardening` — sshd drop-in, fail2ban, unattended-upgrades
- `internal/tailscale` — install, up, peer parsing
- `internal/wizard` — huh forms + orchestration

## Commands

- Build: `go build ./...`
- Vet: `go vet ./...`
- Test: `go test ./...`
- Install locally: `./install.sh`

## Conventions

- Anything that shells out must be testable: keep command strings in packages
  that accept a `system.Runner`; tests use `system.MockRunner` and assert on
  `Calls`.
- Generated config files start with the `ManagedBy` header; regeneration must
  be idempotent and must preserve unmanaged user content (see
  `internal/sshkeys/config.go` patterns).
- New config content: add/extend a string constant + a `strings.Contains`
  test in the owning package.
- Private keys: `0600`. SSH/config dirs: `0700`. Root-written files go through
  `system.WriteRootFile` only.
