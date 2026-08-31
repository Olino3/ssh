package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DirMode  = 0o700
	FileMode = 0o600
)

type Key struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Comment string `yaml:"comment"`
}

type Tailscale struct {
	Role          string `yaml:"role"`
	Authenticated bool   `yaml:"authenticated"`
	Hostname      string `yaml:"hostname"`
}

type State struct {
	Role           string    `yaml:"role"`
	Keys           []Key     `yaml:"keys"`
	Services       []string  `yaml:"services"`
	Tailscale      Tailscale `yaml:"tailscale"`
	CompletedSteps []string  `yaml:"completed_steps"`
	GeneratedAt    time.Time `yaml:"generated_at"`
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "sshutil"), nil
}

func Path() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "state.yaml"), nil
}

func Load() (*State, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	return &s, nil
}

func Save(s *State) error {
	d, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, DirMode); err != nil {
		return err
	}
	s.GeneratedAt = time.Now()
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	p := filepath.Join(d, "state.yaml")
	if err := os.WriteFile(p, data, FileMode); err != nil {
		return err
	}
	return os.Chmod(p, FileMode)
}

func (s *State) MarkComplete(step string) {
	if !s.IsComplete(step) {
		s.CompletedSteps = append(s.CompletedSteps, step)
	}
}

func (s *State) IsComplete(step string) bool {
	for _, c := range s.CompletedSteps {
		if c == step {
			return true
		}
	}
	return false
}

func (s *State) UpsertKey(k Key) {
	for i := range s.Keys {
		if s.Keys[i].Name == k.Name {
			s.Keys[i] = k
			return
		}
	}
	s.Keys = append(s.Keys, k)
}

func (s *State) AddService(name string) {
	if !s.HasService(name) {
		s.Services = append(s.Services, name)
	}
}

func (s *State) HasService(name string) bool {
	for _, x := range s.Services {
		if x == name {
			return true
		}
	}
	return false
}
