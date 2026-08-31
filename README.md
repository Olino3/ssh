# sshutil

SSH + Tailscale, set up right. One binary that turns a fresh Ubuntu/Debian
device into a properly configured member of your tailnet.

## What it does

- Generates per-device **ed25519 keys**: `id_personal`, `id_github`, and one
  key per service (`id_<service>`)
- Installs and wires **funtoo keychain** + ssh-agent (keys survive across
  shells; prompted once per boot)
- Writes a secure, structured **ssh client config** (`~/.ssh/config.d/` with
  managed drop-ins)
- Joins **Tailscale** as a *server* (sshd reachable over the tailnet, hardened)
  or a *client* (outbound only)
- Server role: sshd hardening drop-in, **fail2ban**, unattended-upgrades
- `sshutil sync` generates ssh Host entries for every peer on your tailnet

## Quickstart

```bash
git clone <this-repo> ssh && cd ssh
./install.sh
sshutil          # full interactive wizard
```

## Commands

| Command | What it does |
|---|---|
| `sshutil` / `sshutil init` | Interactive setup wizard (resumable) |
| `sshutil add-key` | Add a per-service key + Host block |
| `sshutil sync` | Regenerate tailnet Host entries from `tailscale status --json` |
| `sshutil status` | Dashboard: keys, agent, keychain, tailscale, hardening |

## What gets modified on your device

| Path | Purpose |
|---|---|
| `~/.ssh/id_*` | Your generated ed25519 keys (0600) |
| `~/.ssh/config` | Header + `Include ~/.ssh/config.d/*` (existing config backed up to `config.pre-sshutil`) |
| `~/.ssh/config.d/*` | Managed drop-ins: defaults, github, personal, per-service, tailnet |
| `~/.bashrc` | Guarded keychain snippet (marker-delimited) |
| `~/.config/sshutil/state.yaml` | sshutil state (0600) |
| `/etc/ssh/sshd_config.d/99-sshutil.conf` | Server hardening (server role) |
| `/etc/fail2ban/jail.local` | sshd jail (server role) |
| `/etc/apt/apt.conf.d/20auto-upgrades` | Unattended upgrades (server role) |

## Security notes

- Private keys are generated **on the device** and never leave it
- Passphrases and Tailscale pre-auth keys are typed into masked prompts and
  are never logged, echoed, or written to state
- sshd config is validated with `sshd -t` before any restart; the wizard
  asks before restarting and reminds you to verify a second session
- Server role disables password auth, root login, and weak ciphers/MACs/Kex

## Uninstall

```bash
rm -rf ~/.config/sshutil ~/.ssh/config.d ~/.ssh/config.pre-sshutil
# remove the marker-delimited sshutil block from ~/.bashrc
sudo rm -f /etc/ssh/sshd_config.d/99-sshutil.conf   # server role
sudo rm -f /etc/fail2ban/jail.local
sudo rm -f /etc/apt/apt.conf.d/20auto-upgrades
sudo systemctl disable --now fail2ban unattended-upgrades 2>/dev/null || true
sudo systemctl restart ssh
# generated keys (~/.ssh/id_personal, ~/.ssh/id_github, ~/.ssh/id_<service>)
# are deliberately left in place — remove them yourself if unwanted
```

## Development

```bash
go test ./...
go build ./...
```

See `AGENTS.md` for repository conventions.
