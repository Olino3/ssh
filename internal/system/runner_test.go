package system

import (
	"strings"
	"testing"
)

func TestMockRunnerRecordsCalls(t *testing.T) {
	m := &MockRunner{}
	if _, err := m.Run("foo", "bar"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.RunSudo("apt-get", "update"); err != nil {
		t.Fatal(err)
	}
	if len(m.Calls) != 2 {
		t.Fatalf("want 2 calls, got %d", len(m.Calls))
	}
	if !m.Calls[1].Sudo || m.Calls[1].Name != "apt-get" {
		t.Fatalf("unexpected call: %+v", m.Calls[1])
	}
}

func TestMockRunnerOutputAndErrs(t *testing.T) {
	m := &MockRunner{
		Output: map[string]string{"echo hi": "hi\n"},
		Errs:   map[string]error{"boom": &TestErr{Msg: "boom"}},
	}
	out, err := m.Run("echo", "hi")
	if err != nil || strings.TrimSpace(out) != "hi" {
		t.Fatalf("out=%q err=%v", out, err)
	}
	if _, err := m.Run("boom"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRedactArgs(t *testing.T) {
	got := redactArgs("ssh-keygen", []string{"-t", "ed25519", "-N", "hunter2", "-f", "/tmp/x"})
	if strings.Contains(got, "hunter2") {
		t.Fatalf("passphrase leaked: %s", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Fatalf("expected redaction marker: %s", got)
	}
	got = redactArgs("tailscale", []string{"up", "--auth-key=secret"})
	if strings.Contains(got, "secret") {
		t.Fatalf("auth key leaked: %s", got)
	}
	if !strings.Contains(got, "--auth-key=[REDACTED]") {
		t.Fatalf("expected auth-key redaction: %s", got)
	}
}

func TestMockRunnerLookPath(t *testing.T) {
	m := &MockRunner{Paths: map[string]bool{"gh": true}}
	if _, err := m.LookPath("gh"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.LookPath("nope"); err == nil {
		t.Fatal("expected error")
	}
}
