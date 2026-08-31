package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"sshutil/internal/sshkeys"
	"sshutil/internal/state"
	"sshutil/internal/system"
	"sshutil/internal/tailscale"
	"sshutil/internal/ui"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sshutil status dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := system.NewExecRunner()

		st, err := state.Load()
		if err != nil {
			ui.Fail("state", err.Error())
			return nil
		}
		if st == nil {
			ui.Fail("state", "no sshutil state — run 'sshutil init' first")
			return nil
		}
		ui.Pass("state present (role: %s)", st.Role)

		m, err := sshkeys.NewDefaultManager(r)
		if err != nil {
			ui.Fail("ssh dir", err.Error())
			return nil
		}
		for _, k := range st.Keys {
			if m.Exists(k.Name) {
				ui.Pass("key %s", k.Name)
			} else {
				ui.Fail("key "+k.Name, "missing on disk")
			}
		}

		if out, err := r.Run("ssh-add", "-l"); err == nil && strings.TrimSpace(out) != "The agent has no identities." {
			n := len(strings.Split(strings.TrimSpace(out), "\n"))
			ui.Pass("agent keys loaded: %d", n)
		} else {
			ui.Fail("agent", "no keys loaded (open a new shell to start keychain)")
		}

		if _, err := r.LookPath("keychain"); err == nil {
			ui.Pass("keychain installed")
		} else {
			ui.Fail("keychain", "not installed")
		}

		if _, err := r.LookPath("tailscale"); err == nil {
			if host, err := tailscale.Hostname(r); err == nil {
				ui.Pass("tailscale: %s (role: %s)", host, st.Tailscale.Role)
			} else {
				ui.Fail("tailscale", "not connected")
			}
		} else {
			ui.Fail("tailscale", "not installed")
		}

		if st.Role == "server" {
			if _, err := r.RunSudo("systemctl", "is-active", "--quiet", "fail2ban"); err == nil {
				ui.Pass("fail2ban active")
			} else {
				ui.Fail("fail2ban", "inactive")
			}
		}
		return nil
	},
}
