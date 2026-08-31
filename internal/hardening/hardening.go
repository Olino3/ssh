package hardening

import (
	"fmt"

	"sshutil/internal/system"
)

const (
	SSHDFile = "/etc/ssh/sshd_config.d/99-sshutil.conf"
	JailFile = "/etc/fail2ban/jail.local"
	AutoFile = "/etc/apt/apt.conf.d/20auto-upgrades"
)

var SSHDConfig = `# Managed by sshutil
PasswordAuthentication no
ChallengeResponseAuthentication no
PermitRootLogin no
PubkeyAuthentication yes
MaxAuthTries 3
LoginGraceTime 30
X11Forwarding no
ClientAliveInterval 300
Ciphers chacha20-poly1305@openssh.com,aes256-gcm@openssh.com,aes128-gcm@openssh.com
MACs hmac-sha2-512-etm@openssh.com,hmac-sha2-256-etm@openssh.com
KexAlgorithms curve25519-sha256,curve25519-sha256@libssh.org
`

func ApplySSHD(r system.Runner) error {
	if _, err := r.RunSudo("mkdir", "-p", "/etc/ssh/sshd_config.d"); err != nil {
		return err
	}
	return system.WriteRootFile(r, SSHDFile, SSHDConfig)
}

func ValidateSSHD(r system.Runner) error {
	if _, err := r.RunSudo("sshd", "-t"); err != nil {
		return fmt.Errorf("sshd -t failed: %w", err)
	}
	return nil
}

func RestartSSHD(r system.Runner) error {
	_, err := r.RunSudo("systemctl", "restart", "ssh")
	return err
}

const Fail2BanJail = `# Managed by sshutil
[sshd]
enabled = true
port = ssh
maxretry = 5
bantime = 1h
findtime = 10m
`

func ApplyFail2Ban(r system.Runner) error {
	if err := system.WriteRootFile(r, JailFile, Fail2BanJail); err != nil {
		return err
	}
	_, err := r.RunSudo("systemctl", "enable", "--now", "fail2ban")
	return err
}

const UnattendedUpgradesConf = `APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
`

func ApplyUnattendedUpgrades(r system.Runner) error {
	if err := system.WriteRootFile(r, AutoFile, UnattendedUpgradesConf); err != nil {
		return err
	}
	_, err := r.RunSudo("systemctl", "enable", "--now", "unattended-upgrades")
	return err
}
