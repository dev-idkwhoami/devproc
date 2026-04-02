package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager_CreatesDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	mgr, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if mgr.Config == nil {
		t.Fatal("Config should not be nil")
	}
	if len(mgr.Config.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(mgr.Config.Projects))
	}
	if mgr.Config.LastActiveProject != "" {
		t.Errorf("expected empty last_active_project, got %q", mgr.Config.LastActiveProject)
	}

	// File should exist on disk
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file should have been created on disk")
	}
}

func TestNewManager_LoadsExistingConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte(`projects:
  - name: "test-project"
    path: "/tmp/test"
    commands:
      - name: "Dev"
        cmd: "npm run dev"
command_history:
  - "npm run dev"
last_active_project: "test-project"
`)
	os.WriteFile(configPath, content, 0644)

	mgr, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if len(mgr.Config.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(mgr.Config.Projects))
	}
	if mgr.Config.Projects[0].Name != "test-project" {
		t.Errorf("expected project name 'test-project', got %q", mgr.Config.Projects[0].Name)
	}
	if len(mgr.Config.Projects[0].Commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(mgr.Config.Projects[0].Commands))
	}
	if mgr.Config.CommandHistory[0] != "npm run dev" {
		t.Errorf("expected command history entry 'npm run dev', got %q", mgr.Config.CommandHistory[0])
	}
}

func TestAddProject(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(filepath.Join(dir, "config.yaml"))

	err := mgr.AddProject("my-app", "/path/to/my-app")
	if err != nil {
		t.Fatalf("AddProject failed: %v", err)
	}

	if len(mgr.Config.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(mgr.Config.Projects))
	}
	if mgr.Config.Projects[0].Name != "my-app" {
		t.Errorf("expected name 'my-app', got %q", mgr.Config.Projects[0].Name)
	}
	if mgr.Config.Projects[0].Path != "/path/to/my-app" {
		t.Errorf("expected path '/path/to/my-app', got %q", mgr.Config.Projects[0].Path)
	}
}

func TestAddProject_DuplicatePath(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(filepath.Join(dir, "config.yaml"))

	mgr.AddProject("my-app", "/path/to/my-app")
	err := mgr.AddProject("my-app-2", "/path/to/my-app")

	if err == nil {
		t.Error("expected error for duplicate path")
	}
}

func TestRemoveProject(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(filepath.Join(dir, "config.yaml"))

	mgr.AddProject("my-app", "/path/to/my-app")
	err := mgr.RemoveProject("my-app")
	if err != nil {
		t.Fatalf("RemoveProject failed: %v", err)
	}
	if len(mgr.Config.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(mgr.Config.Projects))
	}
}

func TestAddCommand(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(filepath.Join(dir, "config.yaml"))

	mgr.AddProject("my-app", "/path/to/my-app")
	err := mgr.AddCommand("my-app", "Vite", "npm run dev")
	if err != nil {
		t.Fatalf("AddCommand failed: %v", err)
	}

	proj := mgr.Config.Projects[0]
	if len(proj.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(proj.Commands))
	}
	if proj.Commands[0].Name != "Vite" {
		t.Errorf("expected command name 'Vite', got %q", proj.Commands[0].Name)
	}
	if proj.Commands[0].Cmd != "npm run dev" {
		t.Errorf("expected cmd 'npm run dev', got %q", proj.Commands[0].Cmd)
	}

	// Should also be in command history
	found := false
	for _, h := range mgr.Config.CommandHistory {
		if h == "npm run dev" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'npm run dev' in command_history")
	}
}

func TestRemoveCommand(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(filepath.Join(dir, "config.yaml"))

	mgr.AddProject("my-app", "/path/to/my-app")
	mgr.AddCommand("my-app", "Vite", "npm run dev")

	err := mgr.RemoveCommand("my-app", "Vite")
	if err != nil {
		t.Fatalf("RemoveCommand failed: %v", err)
	}
	if len(mgr.Config.Projects[0].Commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(mgr.Config.Projects[0].Commands))
	}
}

func TestAddToHistory_NoDuplicates(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(filepath.Join(dir, "config.yaml"))

	mgr.AddProject("my-app", "/path/to/my-app")
	mgr.AddCommand("my-app", "Vite", "npm run dev")
	mgr.AddCommand("my-app", "Vite2", "npm run dev") // same cmd, different name

	count := 0
	for _, h := range mgr.Config.CommandHistory {
		if h == "npm run dev" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 'npm run dev' once in history, found %d", count)
	}
}
