package engine

import (
	"os"
	"path/filepath"
	"testing"

	"dotm/internal/config"
	"dotm/internal/prompt"
)

// ─── isValidShell ────────────────────────────────────────────────────────────

func TestIsValidShell(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		want  bool
	}{
		// Bare names.
		{"bare bash", "bash", true},
		{"bare sh", "sh", true},
		{"bare zsh", "zsh", true},
		{"bare fish", "fish", true},
		// Absolute known paths.
		{"abs /bin/bash", "/bin/bash", true},
		{"abs /bin/sh", "/bin/sh", true},
		{"abs /bin/zsh", "/bin/zsh", true},
		{"abs /bin/fish", "/bin/fish", true},
		{"abs /usr/bin/bash", "/usr/bin/bash", true},
		{"abs /usr/bin/zsh", "/usr/bin/zsh", true},
		// Unknown bare name.
		{"unknown bare", "dash", false},
		// Non-absolute unknown path.
		{"relative path", "bin/bash", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidShell(tt.shell)
			if got != tt.want {
				t.Errorf("isValidShell(%q) = %v, want %v", tt.shell, got, tt.want)
			}
		})
	}
}

func TestIsValidShellAbsolutePath(t *testing.T) {
	tmp := t.TempDir()

	// Create an executable file.
	execPath := filepath.Join(tmp, "myshell")
	if err := os.WriteFile(execPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a non-executable file.
	nonExecPath := filepath.Join(tmp, "notshell")
	if err := os.WriteFile(nonExecPath, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a directory.
	dirPath := filepath.Join(tmp, "dirshell")
	os.Mkdir(dirPath, 0o755)

	tests := []struct {
		name  string
		shell string
		want  bool
	}{
		{"existing executable", execPath, true},
		{"existing non-executable", nonExecPath, false},
		{"directory", dirPath, false},
		{"nonexistent", filepath.Join(tmp, "doesnotexist"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidShell(tt.shell)
			if got != tt.want {
				t.Errorf("isValidShell(%q) = %v, want %v", tt.shell, got, tt.want)
			}
		})
	}
}

// ─── applyPerms fallback (no perms file → default 0o644) ────────────────────

func TestApplyPermsFallback(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create dotm.toml.
	cfgContent := `dest = "` + destDir + `"`
	os.WriteFile(filepath.Join(sourceDir, "dotm.toml"), []byte(cfgContent), 0o644)

	// Create files/ with a test file.
	filesDir := filepath.Join(sourceDir, "files")
	os.MkdirAll(filepath.Join(filesDir, ".config"), 0o755)
	os.WriteFile(filepath.Join(filesDir, ".config", "test.conf"), []byte("content"), 0o644)

	// Do NOT create a perms file.
	cfg, err := config.Load(filepath.Join(sourceDir, "dotm.toml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	state := &prompt.State{
		Data:         make(map[string]any),
		ScriptHashes: make(map[string]string),
	}

	eng, err := New(cfg, state, sourceDir, false)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	// Apply files.
	err = eng.Apply(ScopeFiles)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	// Verify the file was written.
	destPath := filepath.Join(destDir, ".config", "test.conf")
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("stat dest: %v", err)
	}

	// Verify permissions: should be 0o644 (fallback), not 0o600 (initial write).
	got := info.Mode().Perm()
	want := os.FileMode(0o644)
	if got != want {
		t.Errorf("permissions = %04o, want %04o", got, want)
	}
}

func TestApplyPermsFallbackDoesNotOverrideExplicitPerms(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create dotm.toml.
	cfgContent := `dest = "` + destDir + `"`
	os.WriteFile(filepath.Join(sourceDir, "dotm.toml"), []byte(cfgContent), 0o644)

	// Create files/ with a test file.
	filesDir := filepath.Join(sourceDir, "files")
	os.MkdirAll(filepath.Join(filesDir, ".config"), 0o755)
	os.WriteFile(filepath.Join(filesDir, ".config", "test.conf"), []byte("content"), 0o644)

	// Create a perms file that sets 0o600.
	permsContent := `.config/** 0600 - -` + "\n"
	os.WriteFile(filepath.Join(sourceDir, "perms"), []byte(permsContent), 0o644)

	cfg, err := config.Load(filepath.Join(sourceDir, "dotm.toml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	state := &prompt.State{
		Data:         make(map[string]any),
		ScriptHashes: make(map[string]string),
	}

	eng, err := New(cfg, state, sourceDir, false)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	err = eng.Apply(ScopeFiles)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	destPath := filepath.Join(destDir, ".config", "test.conf")
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("stat dest: %v", err)
	}

	// Explicit perms file should override the 0o644 fallback.
	got := info.Mode().Perm()
	want := os.FileMode(0o600)
	if got != want {
		t.Errorf("permissions = %04o, want %04o", got, want)
	}
}

func TestApplyPermsFallbackDryRun(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create dotm.toml.
	cfgContent := `dest = "` + destDir + `"`
	os.WriteFile(filepath.Join(sourceDir, "dotm.toml"), []byte(cfgContent), 0o644)

	// Create files/ with a test file.
	filesDir := filepath.Join(sourceDir, "files")
	os.MkdirAll(filepath.Join(filesDir, ".config"), 0o755)
	os.WriteFile(filepath.Join(filesDir, ".config", "test.conf"), []byte("content"), 0o644)

	// No perms file.
	cfg, err := config.Load(filepath.Join(sourceDir, "dotm.toml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	state := &prompt.State{
		Data:         make(map[string]any),
		ScriptHashes: make(map[string]string),
	}

	eng, err := New(cfg, state, sourceDir, true) // dryRun = true
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	err = eng.Apply(ScopeFiles)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	// In dry-run mode, the file should NOT be written at all.
	destPath := filepath.Join(destDir, ".config", "test.conf")
	if _, err := os.Stat(destPath); err == nil {
		t.Error("expected file NOT to exist in dry-run mode")
	}
}

// ─── Initial file permissions (0o600) before applyPerms corrects them ───────

func TestWalkAndWriteInitialPermissions(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create files/ with a test file.
	filesDir := filepath.Join(sourceDir, "files", ".config")
	os.MkdirAll(filesDir, 0o755)
	os.WriteFile(filepath.Join(filesDir, "secret.conf"), []byte("secret=value"), 0o644)

	cfg := &config.Config{
		Dest:  destDir,
		Shell: "bash",
	}
	state := &prompt.State{
		Data:         make(map[string]any),
		ScriptHashes: make(map[string]string),
	}

	eng, err := New(cfg, state, sourceDir, false)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	// walkAndWrite writes files with 0o600, then applyPerms corrects to 0o644.
	written, err := eng.walkAndWrite()
	if err != nil {
		t.Fatalf("walkAndWrite: %v", err)
	}
	if len(written) < 1 {
		t.Fatalf("expected at least 1 written path, got %d", len(written))
	}

	// At this point, before applyPerms, the file should still be 0o600.
	destPath := filepath.Join(destDir, ".config", "secret.conf")
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("stat dest: %v", err)
	}
	got := info.Mode().Perm()
	want := os.FileMode(0o600)
	if got != want {
		t.Errorf("permissions before applyPerms = %04o, want %04o", got, want)
	}

	// Now call applyPerms — it should lift to 0o644 (no perms file).
	if err := eng.applyPerms(written); err != nil {
		t.Fatalf("applyPerms: %v", err)
	}

	info, err = os.Stat(destPath)
	if err != nil {
		t.Fatalf("stat dest after applyPerms: %v", err)
	}
	got = info.Mode().Perm()
	want = os.FileMode(0o644)
	if got != want {
		t.Errorf("permissions after applyPerms = %04o, want %04o", got, want)
	}
}
