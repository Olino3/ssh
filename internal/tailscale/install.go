package tailscale

import (
	"fmt"

	"sshutil/internal/system"
)

const installScript = "curl -fsSL https://tailscale.com/install.sh | sh"

func Install(r system.Runner) error {
	if _, err := r.LookPath("tailscale"); err == nil {
		return nil
	}
	if _, err := r.LookPath("curl"); err != nil {
		return fmt.Errorf("curl is required to install tailscale: %w", err)
	}
	_, err := r.RunSudo("sh", "-c", installScript)
	return err
}

func Up(r system.Runner, authKey string) error {
	if authKey != "" {
		_, err := r.RunSudo("tailscale", "up", "--auth-key="+authKey)
		return err
	}
	return r.RunSudoInteractive("tailscale", "up")
}
