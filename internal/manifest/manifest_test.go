package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()
	testStateDir = dir
	t.Cleanup(func() { testStateDir = "" })

	orig := &PkgManifest{
		Packages: []PackageEntry{
			{Name: "git", Manager: "pacman"},
			{Name: "zsh", Manager: "pacman"},
		},
		Services: []ServiceEntry{
			{Name: "firewalld", Manager: "systemd"},
			{Name: "waybar", Manager: "systemd-user"},
		},
	}

	if err := Save(dir, orig); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(loaded.Packages))
	}
	if loaded.Packages[0].Name != "git" {
		t.Errorf("expected git, got %s", loaded.Packages[0].Name)
	}

	if len(loaded.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(loaded.Services))
	}
	if loaded.Services[0].Name != "firewalld" {
		t.Errorf("expected firewalld, got %s", loaded.Services[0].Name)
	}
	if loaded.Services[1].Manager != "systemd-user" {
		t.Errorf("expected systemd-user manager, got %s", loaded.Services[1].Manager)
	}
}

func TestLoadEmpty(t *testing.T) {
	dir := t.TempDir()
	testStateDir = dir
	t.Cleanup(func() { testStateDir = "" })

	m, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(m.Packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(m.Packages))
	}
	if len(m.Services) != 0 {
		t.Errorf("expected 0 services, got %d", len(m.Services))
	}
}

func TestStateFile(t *testing.T) {
	dir := t.TempDir()
	testStateDir = dir
	t.Cleanup(func() { testStateDir = "" })

	tests := []struct {
		name      string
		configDir string
	}{
		{"absolute", "/home/user/dotfiles"},
		{"different", "/home/user/root_m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := stateFile(tt.configDir)
			if err != nil {
				t.Fatalf("stateFile: %v", err)
			}
			if !filepath.IsAbs(path) {
				t.Errorf("expected absolute path, got %s", path)
			}
		})
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	testStateDir = dir
	t.Cleanup(func() { testStateDir = "" })

	// Write invalid TOML to the state file path.
	path, err := stateFile(dir)
	if err != nil {
		t.Fatalf("stateFile: %v", err)
	}
	if err := os.WriteFile(path, []byte("this is {{{ invalid toml }}}"), 0o644); err != nil {
		t.Fatalf("write invalid toml: %v", err)
	}

	m, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}

	// Must return a non-nil empty manifest, not nil.
	if m == nil {
		t.Fatal("expected non-nil manifest for invalid TOML, got nil")
	}
	if len(m.Packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(m.Packages))
	}
	if len(m.Services) != 0 {
		t.Errorf("expected 0 services, got %d", len(m.Services))
	}
}
