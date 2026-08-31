package wizard

import (
	"fmt"
	"os"

	"sshutil/internal/hardening"
	"sshutil/internal/sshkeys"
	"sshutil/internal/state"
	"sshutil/internal/system"
	"sshutil/internal/tailscale"
	"sshutil/internal/ui"
)

func Run() error {
	ui.PrintBanner()
	r := system.NewExecRunner()

	st, err := state.Load()
	if err != nil {
		ui.ErrorBox("Cannot read state", err.Error())
		return err
	}
	if st == nil {
		st = &state.State{}
	} else if len(st.CompletedSteps) > 0 {
		resume, err := AskResume()
		if err != nil {
			return err
		}
		if !resume {
			st = &state.State{}
		}
	}

	osInfo, err := system.DetectOS()
	if err != nil {
		ui.ErrorBox("Unsupported OS", err.Error())
		return err
	}
	ui.Info("Detected %s", osInfo.Pretty)

	role, err := AskRole()
	if err != nil {
		return err
	}
	st.Role = role
	ui.Info("Role: %s", role)

	ui.Info("Requesting sudo (may prompt)…")
	if _, err := r.RunSudo("true"); err != nil {
		ui.ErrorBox("Sudo required", "sshutil needs admin rights. Re-run and enter your password.")
		return err
	}

	type step struct {
		name string
		fn   func() error
	}
	var steps []step
	steps = append(steps, step{"Packages", func() error { return installPackages(r, st) }})
	steps = append(steps, step{"SSH keys", func() error { return setupKeys(r, st) }})
	if role == "server" {
		steps = append(steps, step{"Server hardening", func() error { return applyHardening(r, st) }})
	}
	steps = append(steps, step{"Tailscale", func() error { return setupTailscale(r, st) }})
	steps = append(steps, step{"Finalize", func() error { return finalize(r, st) }})

	for i, s := range steps {
		ui.Step(i+1, len(steps), s.name)
		for {
			err := s.fn()
			if err == nil {
				break
			}
			ui.ErrorBox("Step failed: "+s.name, err.Error())
			act, aerr := AskRetry(s.name)
			if aerr != nil {
				return aerr
			}
			switch act {
			case "retry":
				continue
			case "skip":
				ui.Warn("Skipped %s — re-run 'sshutil init' later to finish it", s.name)
			default:
				return fmt.Errorf("aborted at step %s", s.name)
			}
			break
		}
		if err := state.Save(st); err != nil {
			ui.Warn("Could not save state: %v", err)
		}
	}

	ui.Success("Setup complete — role: %s", st.Role)
	ui.Muted("Refresh tailnet host entries anytime with: sshutil sync")
	return nil
}

func installPackages(r system.Runner, st *state.State) error {
	if _, err := r.RunSudo("apt-get", "update"); err != nil {
		return err
	}
	pkgs := []string{"keychain"}
	if st.Role == "server" {
		pkgs = append(pkgs, "openssh-server", "fail2ban", "unattended-upgrades")
	}
	args := append([]string{"install", "-y"}, pkgs...)
	if _, err := r.RunSudo("apt-get", args...); err != nil {
		return err
	}
	if err := tailscale.Install(r); err != nil {
		return err
	}
	st.MarkComplete("packages")
	return nil
}

func setupKeys(r system.Runner, st *state.State) error {
	m, err := sshkeys.NewDefaultManager(r)
	if err != nil {
		return err
	}
	host, _ := os.Hostname()
	username := CurrentUser()

	cores := []struct{ name, purpose string }{
		{sshkeys.PersonalKeyName, "personal"},
		{sshkeys.GitHubKeyName, "github"},
	}
	for _, c := range cores {
		if err := ensureKey(m, st, c.name, username+"@"+host+":"+c.purpose); err != nil {
			return err
		}
		pub, err := m.PublicKey(c.name)
		if err != nil {
			return err
		}
		fmt.Println(ui.KeyPanel(c.name, pub))
		if c.name == sshkeys.GitHubKeyName {
			if err := addGitHubKey(r, m, host); err != nil {
				return err
			}
		}
	}

	for {
		more, err := AskMoreServices()
		if err != nil {
			return err
		}
		if !more {
			break
		}
		name, err := AskServiceName()
		if err != nil {
			return err
		}
		keyName := sshkeys.ServiceKeyName(name)
		if err := ensureKey(m, st, keyName, username+"@"+host+":"+name); err != nil {
			return err
		}
		st.AddService(name)
	}

	if err := sshkeys.WriteConfigFiles(m.SSHDir, st.Services); err != nil {
		return err
	}
	appended, err := sshkeys.EnsureBashrc()
	if err != nil {
		return err
	}
	if appended {
		ui.Info("Added keychain snippet to ~/.bashrc (agent starts in new shells)")
	}
	st.MarkComplete("keys")
	return nil
}

func ensureKey(m *sshkeys.Manager, st *state.State, name, comment string) error {
	if m.Exists(name) {
		ui.Info("Key %s exists — keeping", name)
		st.UpsertKey(state.Key{Name: name, Path: m.KeyPath(name), Comment: comment})
		return nil
	}
	pw, err := AskPassphrase(name)
	if err != nil {
		return err
	}
	if err := m.Generate(name, comment, pw); err != nil {
		return err
	}
	st.UpsertKey(state.Key{Name: name, Path: m.KeyPath(name), Comment: comment})
	return nil
}

func addGitHubKey(r system.Runner, m *sshkeys.Manager, host string) error {
	if _, err := r.LookPath("gh"); err != nil {
		ui.Muted("gh CLI not found — add the key manually at https://github.com/settings/keys")
		return nil
	}
	ok, err := ConfirmGhAdd()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if _, err := r.Run("gh", "ssh-key", "add", m.KeyPath(sshkeys.GitHubKeyName)+".pub", "--title", "sshutil "+host); err != nil {
		return fmt.Errorf("gh ssh-key add: %w", err)
	}
	ui.Success("GitHub key added via gh")
	return nil
}

func applyHardening(r system.Runner, st *state.State) error {
	if err := hardening.ApplySSHD(r); err != nil {
		return err
	}
	if err := hardening.ValidateSSHD(r); err != nil {
		return err
	}
	if err := hardening.ApplyFail2Ban(r); err != nil {
		return err
	}
	if err := hardening.ApplyUnattendedUpgrades(r); err != nil {
		return err
	}
	st.MarkComplete("hardening")
	return nil
}

func setupTailscale(r system.Runner, st *state.State) error {
	if err := tailscale.Install(r); err != nil {
		return err
	}
	authKey, err := AskTailscaleAuth()
	if err != nil {
		return err
	}
	if err := tailscale.Up(r, authKey); err != nil {
		return err
	}
	host, err := tailscale.Hostname(r)
	if err != nil {
		return err
	}
	st.Tailscale = state.Tailscale{Role: st.Role, Authenticated: true, Hostname: host}
	st.MarkComplete("tailscale")
	ui.Success("Joined tailnet as %s (%s)", host, st.Role)
	return nil
}

func finalize(r system.Runner, st *state.State) error {
	if st.Role != "server" {
		st.MarkComplete("finalize")
		return nil
	}
	ok, err := ConfirmRestart()
	if err != nil {
		return err
	}
	if ok {
		if err := hardening.RestartSSHD(r); err != nil {
			return err
		}
		ui.Success("sshd restarted — verify a second session before closing this one")
	} else {
		ui.Warn("sshd restart pending — run: sudo systemctl restart ssh")
	}
	st.MarkComplete("finalize")
	return nil
}
