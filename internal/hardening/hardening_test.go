package hardening

import (
	"strings"
	"testing"

	"sshutil/internal/system"
)

func TestSSHDConfigDirectives(t *testing.T) {
	for _, want := range []string{
		"PasswordAuthentication no", "ChallengeResponseAuthentication no",
		"PermitRootLogin no", "PubkeyAuthentication yes",
		"MaxAuthTries 3", "LoginGraceTime 30", "X11Forwarding no",
		"ClientAliveInterval 300", "Ciphers ", "MACs ",
		"KexAlgorithms curve25519-sha256,curve25519-sha256@libssh.org",
	} {
		if !strings.Contains(SSHDConfig, want) {
			t.Fatalf("SSHDConfig missing %q", want)
		}
	}
}

func TestApplySSHD(t *testing.T) {
	m := &system.MockRunner{}
	if err := ApplySSHD(m); err != nil {
		t.Fatal(err)
	}
	last := m.Calls[len(m.Calls)-1]
	if !last.Sudo || last.Name != "install" || last.Args[len(last.Args)-1] != SSHDFile {
		t.Fatalf("unexpected final call: %+v", last)
	}
}

func TestValidateSSHD(t *testing.T) {
	m := &system.MockRunner{}
	if err := ValidateSSHD(m); err != nil {
		t.Fatal(err)
	}
	if c := m.Calls[0]; !c.Sudo || c.Name != "sshd" || c.Args[0] != "-t" {
		t.Fatalf("unexpected call: %+v", c)
	}
	m2 := &system.MockRunner{Errs: map[string]error{"sshd -t": &system.TestErr{Msg: "bad"}}}
	err := ValidateSSHD(m2)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "sshd -t failed:") {
		t.Fatalf("expected error to mention sshd -t failed, got: %v", err)
	}
}

func TestApplyFail2Ban(t *testing.T) {
	m := &system.MockRunner{}
	if err := ApplyFail2Ban(m); err != nil {
		t.Fatal(err)
	}
	installIdx, enableIdx := -1, -1
	for i, c := range m.Calls {
		if c.Sudo && c.Name == "systemctl" && len(c.Args) == 3 &&
			c.Args[0] == "enable" && c.Args[1] == "--now" && c.Args[2] == "fail2ban" {
			enableIdx = i
		}
		if c.Name == "install" {
			installIdx = i
		}
	}
	if enableIdx == -1 {
		t.Fatal("fail2ban not enabled")
	}
	if installIdx == -1 || installIdx > enableIdx {
		t.Fatalf("jail install must precede systemctl enable: install=%d enable=%d", installIdx, enableIdx)
	}
}

func TestApplyUnattendedUpgrades(t *testing.T) {
	m := &system.MockRunner{}
	if err := ApplyUnattendedUpgrades(m); err != nil {
		t.Fatal(err)
	}
	installIdx, enableIdx := -1, -1
	for i, c := range m.Calls {
		if c.Sudo && c.Name == "systemctl" && len(c.Args) == 3 &&
			c.Args[0] == "enable" && c.Args[1] == "--now" && c.Args[2] == "unattended-upgrades" {
			enableIdx = i
		}
		if c.Name == "install" {
			installIdx = i
		}
	}
	if enableIdx == -1 {
		t.Fatal("unattended-upgrades not enabled")
	}
	if installIdx == -1 || installIdx > enableIdx {
		t.Fatalf("config install must precede systemctl enable: install=%d enable=%d", installIdx, enableIdx)
	}
}

func TestRestartSSHD(t *testing.T) {
	m := &system.MockRunner{}
	if err := RestartSSHD(m); err != nil {
		t.Fatal(err)
	}
	if c := m.Calls[0]; !c.Sudo || c.Name != "systemctl" || len(c.Args) != 2 || c.Args[0] != "restart" || c.Args[1] != "ssh" {
		t.Fatalf("unexpected call: %+v", c)
	}
}

func TestApplySSHDPropagatesMkdirFailure(t *testing.T) {
	m := &system.MockRunner{Errs: map[string]error{
		"mkdir -p /etc/ssh/sshd_config.d": &system.TestErr{Msg: "nope"},
	}}
	if err := ApplySSHD(m); err == nil {
		t.Fatal("expected mkdir failure to propagate")
	}
	for _, c := range m.Calls {
		if c.Name == "install" || c.Name == "systemctl" {
			t.Fatalf("must abort before writing/enabling: %+v", c)
		}
	}
}
