package system

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(name string, args ...string) (string, error)
	RunSudo(name string, args ...string) (string, error)
	RunSudoInteractive(name string, args ...string) error
	LookPath(name string) (string, error)
}

type ExecRunner struct{}

func NewExecRunner() *ExecRunner { return &ExecRunner{} }

func (e *ExecRunner) Run(name string, args ...string) (string, error) {
	return runCommand(false, name, args...)
}

func (e *ExecRunner) RunSudo(name string, args ...string) (string, error) {
	return runCommand(true, name, args...)
}

func (e *ExecRunner) RunSudoInteractive(name string, args ...string) error {
	cmd := exec.Command("sudo", append([]string{name}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func redactArgs(name string, args []string) string {
	redacted := make([]string, len(args))
	redactNext := false
	for i, a := range args {
		switch {
		case redactNext:
			redacted[i] = "[REDACTED]"
			redactNext = false
		case a == "-N":
			redacted[i] = a
			redactNext = true
		case strings.HasPrefix(a, "--auth-key="):
			redacted[i] = "--auth-key=[REDACTED]"
		default:
			redacted[i] = a
		}
	}
	return name + " " + strings.Join(redacted, " ")
}

func runCommand(sudo bool, name string, args ...string) (string, error) {
	if sudo {
		args = append([]string{name}, args...)
		name = "sudo"
	}
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("%s failed: %w\n%s",
			redactArgs(name, args), err, strings.TrimSpace(out.String()))
	}
	return out.String(), nil
}

func (e *ExecRunner) LookPath(name string) (string, error) {
	p, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH", name)
	}
	return p, nil
}

type Call struct {
	Name string
	Args []string
	Sudo bool
}

type MockRunner struct {
	Calls  []Call
	Output map[string]string
	Errs   map[string]error
	Paths  map[string]bool
}

func (m *MockRunner) Run(name string, args ...string) (string, error) {
	return m.exec(false, name, args...)
}

func (m *MockRunner) RunSudo(name string, args ...string) (string, error) {
	return m.exec(true, name, args...)
}

func (m *MockRunner) RunSudoInteractive(name string, args ...string) error {
	_, err := m.exec(true, name, args...)
	return err
}

func (m *MockRunner) exec(sudo bool, name string, args ...string) (string, error) {
	m.Calls = append(m.Calls, Call{Name: name, Args: args, Sudo: sudo})
	key := strings.Join(append([]string{name}, args...), " ")
	if m.Errs != nil {
		if err, ok := m.Errs[key]; ok {
			return "", err
		}
	}
	if m.Output != nil {
		if out, ok := m.Output[key]; ok {
			return out, nil
		}
	}
	return "", nil
}

func (m *MockRunner) LookPath(name string) (string, error) {
	if m.Paths[name] {
		return "/usr/bin/" + name, nil
	}
	return "", fmt.Errorf("%s not found in PATH", name)
}

// TestErr is a tiny error helper for tests in any package that imports system.
type TestErr struct{ Msg string }

func (e *TestErr) Error() string { return e.Msg }
