package sshkeys

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBlocksContainExpected(t *testing.T) {
	if !strings.Contains(ClientDefaults, "AddKeysToAgent yes") ||
		!strings.Contains(ClientDefaults, "IdentitiesOnly yes") ||
		!strings.Contains(ClientDefaults, "HashKnownHosts yes") {
		t.Fatalf("bad defaults: %s", ClientDefaults)
	}
	if !strings.Contains(GitHubBlock, "Host github.com") ||
		!strings.Contains(GitHubBlock, "IdentityFile ~/.ssh/id_github") ||
		!strings.Contains(GitHubBlock, "User git") {
		t.Fatalf("bad github block: %s", GitHubBlock)
	}
	if !strings.Contains(PersonalBlock, "Host *") ||
		!strings.Contains(PersonalBlock, "IdentityFile ~/.ssh/id_personal") {
		t.Fatalf("bad personal block: %s", PersonalBlock)
	}
	sb := ServiceBlock("nas")
	if !strings.Contains(sb, "Host nas") || !strings.Contains(sb, "IdentityFile ~/.ssh/id_nas") {
		t.Fatalf("bad service block: %s", sb)
	}
	pb := TailnetPeerBlock("web", "web.tail1234.ts.net", "alice")
	for _, want := range []string{"Host web", "HostName web.tail1234.ts.net", "User alice", "IdentityFile ~/.ssh/id_personal"} {
		if !strings.Contains(pb, want) {
			t.Fatalf("peer block missing %q: %s", want, pb)
		}
	}
}

func TestWriteConfigFiles(t *testing.T) {
	dir := t.TempDir()
	if err := WriteConfigFiles(dir, []string{"nas"}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"00-defaults", "10-github", "10-personal", "20-nas"} {
		if _, err := os.Stat(filepath.Join(dir, "config.d", name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	main, err := os.ReadFile(filepath.Join(dir, "config"))
	if err != nil || !strings.Contains(string(main), "Include ~/.ssh/config.d/*") {
		t.Fatalf("bad main config: %s err=%v", main, err)
	}
	defaults, err := os.ReadFile(filepath.Join(dir, "config.d", "00-defaults"))
	if err != nil || !strings.Contains(string(defaults), "AddKeysToAgent yes") {
		t.Fatalf("bad 00-defaults content: %s err=%v", defaults, err)
	}
	github, err := os.ReadFile(filepath.Join(dir, "config.d", "10-github"))
	if err != nil || !strings.Contains(string(github), "IdentityFile ~/.ssh/id_github") {
		t.Fatalf("bad 10-github content: %s err=%v", github, err)
	}
	svc, err := os.ReadFile(filepath.Join(dir, "config.d", "20-nas"))
	if err != nil || !strings.Contains(string(svc), "Host nas") {
		t.Fatalf("bad 20-nas content: %s err=%v", svc, err)
	}
	if fi, err := os.Stat(filepath.Join(dir, "config.d")); err != nil || fi.Mode().Perm() != 0o700 {
		t.Fatalf("config.d perms wrong: %v err=%v", fi, err)
	}
	if fi, err := os.Stat(filepath.Join(dir, "config.d", "00-defaults")); err != nil || fi.Mode().Perm() != 0o644 {
		t.Fatalf("drop-in perms wrong: %v err=%v", fi, err)
	}
}

func TestEnsureMainConfigRefusesOverwriteOfBackup(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config"), []byte("Host old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.pre-sshutil"), []byte("Host prior\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteConfigFiles(dir, nil); err == nil {
		t.Fatal("expected error when backup already exists")
	}
}

func TestEnsureMainConfigBacksUpUnmanaged(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config"), []byte("Host old\n"), 0o644)
	if err := WriteConfigFiles(dir, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.pre-sshutil")); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
}

func TestEnsureMainConfigOverwritesManaged(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config"), []byte(ManagedBy+"\njunk\n"), 0o644)
	if err := WriteConfigFiles(dir, nil); err != nil {
		t.Fatal(err)
	}
	main, _ := os.ReadFile(filepath.Join(dir, "config"))
	if strings.Contains(string(main), "junk") {
		t.Fatalf("managed file not regenerated: %s", main)
	}
}

func TestEnsureBashrc(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	appended, err := EnsureBashrc()
	if err != nil || !appended {
		t.Fatalf("appended=%v err=%v", appended, err)
	}
	appended, err = EnsureBashrc()
	if err != nil || appended {
		t.Fatalf("second run must not append: appended=%v err=%v", appended, err)
	}
}

func TestWriteTailnetConfig(t *testing.T) {
	dir := t.TempDir()
	err := WriteTailnetConfig(dir, []string{TailnetPeerBlock("a", "a.x.ts.net", "u")})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "config.d", "90-tailnet"))
	if err != nil || !strings.Contains(string(data), "Host a") {
		t.Fatalf("bad tailnet config: %s err=%v", data, err)
	}
}
