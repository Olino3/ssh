package sshkeys

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"sshutil/internal/system"
)

// keyCreatingRunner wraps MockRunner and creates the private key file when
// ssh-keygen succeeds, mirroring what real ssh-keygen does.
type keyCreatingRunner struct {
	*system.MockRunner
	keyPath string
}

func (c *keyCreatingRunner) Run(name string, args ...string) (string, error) {
	out, err := c.MockRunner.Run(name, args...)
	if err == nil && name == "ssh-keygen" {
		if werr := os.WriteFile(c.keyPath, []byte("private"), 0o600); werr != nil {
			return out, werr
		}
	}
	return out, err
}

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	mock := &system.MockRunner{}
	m := NewManager(&keyCreatingRunner{MockRunner: mock, keyPath: filepath.Join(dir, "id_personal")}, dir)
	err := m.Generate("id_personal", "u@h:personal", "secret")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-t", "ed25519", "-f", m.KeyPath("id_personal"), "-C", "u@h:personal", "-N", "secret"}
	if len(mock.Calls) != 1 || mock.Calls[0].Name != "ssh-keygen" || !reflect.DeepEqual(mock.Calls[0].Args, want) {
		t.Fatalf("unexpected call: %+v", mock.Calls)
	}
	info, err := os.Stat(m.KeyPath("id_personal"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("private key mode %v, want 0600", info.Mode().Perm())
	}
}

func TestGenerateChmodFailurePropagates(t *testing.T) {
	mock := &system.MockRunner{}
	m := NewManager(mock, t.TempDir())
	err := m.Generate("id_personal", "c", "pw")
	if err == nil {
		t.Fatal("expected chmod failure to propagate when the key file was not created")
	}
	if !strings.Contains(err.Error(), "chmod") {
		t.Fatalf("error should mention chmod: %v", err)
	}
}

func TestGenerateRefusesExisting(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "id_personal"), []byte("k"), 0o600)
	mock := &system.MockRunner{}
	m := NewManager(mock, dir)
	if err := m.Generate("id_personal", "c", ""); err == nil {
		t.Fatal("expected error for existing key")
	}
	if len(mock.Calls) != 0 {
		t.Fatalf("expected no clobber attempt, got calls: %+v", mock.Calls)
	}
}

func TestGenerateRunnerErrorPropagates(t *testing.T) {
	dir := t.TempDir()
	key := "ssh-keygen -t ed25519 -f " + filepath.Join(dir, "id_personal") + " -C c -N pw"
	mock := &system.MockRunner{Errs: map[string]error{key: &system.TestErr{Msg: "boom"}}}
	m := NewManager(mock, dir)
	err := m.Generate("id_personal", "c", "pw")
	if err == nil {
		t.Fatal("expected ssh-keygen failure to propagate")
	}
	if !strings.Contains(err.Error(), "ssh-keygen") {
		t.Fatalf("error should mention ssh-keygen: %v", err)
	}
}

func TestExistsAndPublicKey(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "id_github.pub"), []byte("ssh-ed25519 AAA github\n"), 0o644)
	m := NewManager(&system.MockRunner{}, dir)
	if !m.Exists("id_github") {
		t.Fatal("Exists false")
	}
	if m.Exists("id_personal") {
		t.Fatal("Exists true for missing key")
	}
	pub, err := m.PublicKey("id_github")
	if err != nil || pub != "ssh-ed25519 AAA github" {
		t.Fatalf("pub=%q err=%v", pub, err)
	}
}

func TestServiceKeyName(t *testing.T) {
	if ServiceKeyName("nas") != "id_nas" {
		t.Fatal("bad name")
	}
}
