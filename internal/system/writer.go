package system

import (
	"os"
	"path/filepath"
)

func WriteRootFile(r Runner, path, content string) error {
	dir := filepath.Dir(path)
	if dir != "/" && filepath.Dir(dir) != "/" {
		if _, err := r.RunSudo("mkdir", "-p", dir); err != nil {
			return err
		}
	}
	tmp, err := os.CreateTemp("", "sshutil-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(name, 0o644); err != nil {
		return err
	}
	_, err = r.RunSudo("install", "-m", "0644", name, path)
	return err
}
