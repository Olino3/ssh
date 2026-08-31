package sshkeys

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sshutil/internal/system"
)

const (
	PersonalKeyName = "id_personal"
	GitHubKeyName   = "id_github"
)

type Manager struct {
	Runner system.Runner
	SSHDir string
}

func DefaultSSHDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh"), nil
}

func NewManager(r system.Runner, sshDir string) *Manager {
	return &Manager{Runner: r, SSHDir: sshDir}
}

func NewDefaultManager(r system.Runner) (*Manager, error) {
	dir, err := DefaultSSHDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Manager{Runner: r, SSHDir: dir}, nil
}

func (m *Manager) KeyPath(name string) string { return filepath.Join(m.SSHDir, name) }

func ServiceKeyName(service string) string { return "id_" + service }

func (m *Manager) Exists(name string) bool {
	if _, err := os.Stat(m.KeyPath(name)); err == nil {
		return true
	}
	_, err := os.Stat(m.KeyPath(name) + ".pub")
	return err == nil
}

func (m *Manager) Generate(name, comment, passphrase string) error {
	path := m.KeyPath(name)
	if m.Exists(name) {
		return fmt.Errorf("key %s already exists", path)
	}
	if _, err := m.Runner.Run("ssh-keygen", "-t", "ed25519", "-f", path, "-C", comment, "-N", passphrase); err != nil {
		return fmt.Errorf("ssh-keygen %s: %w", name, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("chmod %s: %w", path, err)
	}
	_ = os.Chmod(path+".pub", 0o644)
	return nil
}

func (m *Manager) PublicKey(name string) (string, error) {
	data, err := os.ReadFile(m.KeyPath(name) + ".pub")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
