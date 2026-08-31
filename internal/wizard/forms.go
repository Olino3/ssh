package wizard

import (
	"errors"
	"os/user"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
)

var serviceNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func CurrentUser() string {
	u, err := user.Current()
	if err != nil || u.Username == "" {
		return "user"
	}
	return u.Username
}

func AskRole() (string, error) {
	var role string
	err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("What is this device?").
			Description("Server: accepts inbound SSH over your tailnet. Client: outbound only.").
			Options(
				huh.NewOption("Server — reachable over tailnet", "server"),
				huh.NewOption("Client — outbound only", "client"),
			).Value(&role),
	)).Run()
	return role, err
}

func AskResume() (bool, error) {
	var ok bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Found previous sshutil state. Resume where you left off?").
			Value(&ok),
	)).Run()
	return ok, err
}

func AskPassphrase(name string) (string, error) {
	var pw, confirm string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Passphrase for "+name+" (empty = no passphrase)").
			EchoMode(huh.EchoModePassword).
			Value(&pw),
		huh.NewInput().
			Title("Confirm passphrase").
			EchoMode(huh.EchoModePassword).
			Value(&confirm).
			Validate(func(s string) error {
				if s != pw {
					return errors.New("passphrases do not match")
				}
				return nil
			}),
	)).Run()
	if err != nil {
		return "", err
	}
	return pw, nil
}

func AskServiceName() (string, error) {
	var name string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Service name (lowercase letters, digits, dashes — e.g. nas, gitlab)").
			Value(&name).
			Validate(func(s string) error {
				if !serviceNameRe.MatchString(strings.TrimSpace(s)) {
					return errors.New("invalid name: use lowercase letters, digits, dashes")
				}
				return nil
			}),
	)).Run()
	return strings.TrimSpace(name), err
}

func AskMoreServices() (bool, error) {
	var ok bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title("Add a per-service key?").Value(&ok),
	)).Run()
	return ok, err
}

func AskTailscaleAuth() (string, error) {
	var useKey bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Use a Tailscale pre-auth key? (No = browser login)").
			Value(&useKey),
	)).Run()
	if err != nil {
		return "", err
	}
	if !useKey {
		return "", nil
	}
	var key string
	err = huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Tailscale pre-auth key").
			EchoMode(huh.EchoModePassword).
			Value(&key).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("key cannot be empty")
				}
				return nil
			}),
	)).Run()
	return strings.TrimSpace(key), err
}

func ConfirmGhAdd() (bool, error) {
	var ok bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title("Add this key to GitHub via the gh CLI?").Value(&ok),
	)).Run()
	return ok, err
}

func ConfirmRestart() (bool, error) {
	var ok bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Restart sshd now? Keep this session open until you verified a second login.").
			Value(&ok),
	)).Run()
	return ok, err
}

func AskRetry(step string) (string, error) {
	var act string
	err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Step failed: "+step).
			Options(
				huh.NewOption("Retry", "retry"),
				huh.NewOption("Skip (you can fix later and re-run)", "skip"),
				huh.NewOption("Abort", "abort"),
			).Value(&act),
	)).Run()
	return act, err
}
