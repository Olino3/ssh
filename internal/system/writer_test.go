package system

import "testing"

func TestWriteRootFile(t *testing.T) {
	m := &MockRunner{}
	err := WriteRootFile(m, "/etc/ssh/sshd_config.d/99-sshutil.conf", "content")
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Calls) != 2 {
		t.Fatalf("want 2 calls (mkdir, install), got %d", len(m.Calls))
	}
	mkdir, install := m.Calls[0], m.Calls[1]
	if !mkdir.Sudo || mkdir.Name != "mkdir" || mkdir.Args[0] != "-p" || mkdir.Args[1] != "/etc/ssh/sshd_config.d" {
		t.Fatalf("unexpected mkdir: %+v", mkdir)
	}
	if !install.Sudo || install.Name != "install" || install.Args[len(install.Args)-1] != "/etc/ssh/sshd_config.d/99-sshutil.conf" {
		t.Fatalf("unexpected install: %+v", install)
	}
}

func TestWriteRootFileNoMkdirForRootPath(t *testing.T) {
	m := &MockRunner{}
	if err := WriteRootFile(m, "/etc/foo.conf", "x"); err != nil {
		t.Fatal(err)
	}
	if len(m.Calls) != 1 {
		t.Fatalf("want 1 call, got %d", len(m.Calls))
	}
}

type failInstallRunner struct{ MockRunner }

func (f *failInstallRunner) RunSudo(name string, args ...string) (string, error) {
	if name == "install" {
		return "", &TestErr{Msg: "install failed"}
	}
	return f.MockRunner.RunSudo(name, args...)
}

func TestWriteRootFileInstallErrorPropagates(t *testing.T) {
	r := &failInstallRunner{}
	if err := WriteRootFile(r, "/etc/ssh/sshd_config.d/99-sshutil.conf", "content"); err == nil {
		t.Fatal("expected install error to propagate")
	}
}

func TestWriteRootFileMkdirErrorPropagates(t *testing.T) {
	m := &MockRunner{Errs: map[string]error{
		"mkdir -p /etc/ssh/sshd_config.d": &TestErr{Msg: "mkdir failed"},
	}}
	if err := WriteRootFile(m, "/etc/ssh/sshd_config.d/99-sshutil.conf", "content"); err == nil {
		t.Fatal("expected mkdir error to propagate")
	}
}
