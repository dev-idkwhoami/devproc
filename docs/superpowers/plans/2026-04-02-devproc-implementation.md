# DevProc Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a system tray desktop app that manages long-running development commands per project.

**Architecture:** Go backend handles process spawning/killing, YAML config persistence, and systray icon management. Svelte 5 frontend renders a borderless popup window with project selector, command list, and add-command dialog. Wails v2 bridges Go ↔ Svelte via bindings and events.

**Tech Stack:** Go 1.21+, Wails v2.11, Svelte 5, TypeScript, `gopkg.in/yaml.v3`, `github.com/energye/systray`

**Spec:** `docs/superpowers/specs/2026-04-02-devproc-design.md`

---

## File Structure

```
devproc/
├── main.go                          # Entry point — Wails app + systray init
├── app.go                           # App struct — Wails-bound methods (frontend API)
├── internal/
│   ├── config/
│   │   ├── config.go                # ConfigManager — load/save YAML, CRUD projects & commands
│   │   └── config_test.go           # Tests for ConfigManager
│   ├── process/
│   │   ├── manager.go               # ProcessManager — spawn, kill, monitor commands
│   │   ├── kill_windows.go          # Windows tree kill (taskkill /T /F)
│   │   ├── kill_unix.go             # Unix process group kill (SIGTERM → SIGKILL)
│   │   └── manager_test.go          # Tests for ProcessManager
│   └── tray/
│       ├── tray.go                  # SystrayManager — icon state, menu, click handlers
│       └── icons.go                 # Embedded icon bytes (grey/green/yellow/red)
├── icons/
│   ├── grey.ico                     # No project selected
│   ├── green.ico                    # All commands running
│   ├── yellow.ico                   # Partial — some stopped/errored
│   └── red.ico                      # All commands down
├── frontend/
│   ├── src/
│   │   ├── main.ts                  # Svelte mount
│   │   ├── App.svelte               # Root — layout shell, routes nothing
│   │   ├── app.css                  # Global dark brutalist theme
│   │   ├── lib/
│   │   │   ├── state.svelte.ts      # Svelte 5 runes store — app state + backend calls
│   │   │   ├── TitleBar.svelte      # DEVPROC label + close button
│   │   │   ├── ProjectSelector.svelte # Dropdown + add project
│   │   │   ├── CommandList.svelte   # Renders CommandItem per command
│   │   │   ├── CommandItem.svelte   # Status dot + name + action button + delete
│   │   │   ├── AddCommandDialog.svelte # Name/cmd fields + history list
│   │   │   └── BottomBar.svelte     # Add CMD + Start All + Kill All
│   │   └── wails.d.ts              # Wails runtime type declarations
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── svelte.config.js
│   └── tsconfig.json
├── build/
│   └── appicon.png
├── wails.json
├── go.mod
└── go.sum
```

---

### Task 1: Scaffold Wails Project

**Files:**
- Create: all project scaffolding via `wails init`
- Modify: `wails.json` (window config)

- [ ] **Step 1: Initialize Wails project in a temp directory and move files**

`wails init` requires an empty directory. Scaffold into a temp location, then move everything to the project root:

```bash
# Scaffold into a temp dir
wails init -n devproc -t svelte-ts -d "$TEMP/devproc-scaffold"

# Move everything from the scaffolded dir into our project root
cp -r "$TEMP/devproc-scaffold/devproc/"* "$PROJECT_ROOT/"
cp "$TEMP/devproc-scaffold/devproc/".* "$PROJECT_ROOT/" 2>/dev/null || true

# Clean up temp
rm -rf "$TEMP/devproc-scaffold"
```

The `docs/` directory already exists in the project root — it will be preserved alongside the Wails files.

- [ ] **Step 3: Configure wails.json for borderless window**

Edit `wails.json` — set these values:

```json
{
  "name": "devproc",
  "outputfilename": "devproc",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "",
    "email": ""
  },
  "info": {
    "companyName": "",
    "productName": "DevProc",
    "productVersion": "0.1.0",
    "copyright": "",
    "comments": "Development Process Manager"
  },
  "wailsdir": ""
}
```

- [ ] **Step 4: Configure main.go for frameless window**

Replace the contents of `main.go` with:

```go
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "DevProc",
		Width:     320,
		Height:    480,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
```

- [ ] **Step 5: Create minimal app.go**

Replace `app.go` with:

```go
package main

import "context"

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
```

- [ ] **Step 6: Install frontend dependencies and verify build**

Run:
```bash
cd frontend && npm install && cd ..
wails build
```

Expected: builds successfully, produces `build/bin/devproc.exe`

- [ ] **Step 7: Commit**

```bash
git init
echo "node_modules/\nbuild/bin/\nfrontend/dist/\n.superpowers/" > .gitignore
git add .
git commit -m "feat: scaffold Wails v2 project with Svelte-TS template"
```

---

### Task 2: ConfigManager — Load and Save

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write failing test for default config creation**

Create `internal/config/config_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -v -run TestNewManager_CreatesDefaultConfig`
Expected: FAIL — package doesn't exist yet

- [ ] **Step 3: Implement ConfigManager with NewManager and Save**

Create `internal/config/config.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Command struct {
	Name string `yaml:"name"`
	Cmd  string `yaml:"cmd"`
}

type Project struct {
	Name     string    `yaml:"name"`
	Path     string    `yaml:"path"`
	Commands []Command `yaml:"commands"`
}

type Config struct {
	Projects          []Project `yaml:"projects"`
	CommandHistory    []string  `yaml:"command_history"`
	LastActiveProject string    `yaml:"last_active_project"`
}

type Manager struct {
	Config   *Config
	filePath string
	mu       sync.Mutex
}

func NewManager(filePath string) (*Manager, error) {
	mgr := &Manager{
		filePath: filePath,
		Config: &Config{
			Projects:       []Project{},
			CommandHistory: []string{},
		},
	}

	data, err := os.ReadFile(filePath)
	if err == nil {
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			if cfg.Projects == nil {
				cfg.Projects = []Project{}
			}
			if cfg.CommandHistory == nil {
				cfg.CommandHistory = []string{}
			}
			mgr.Config = &cfg
		}
	}

	if err := mgr.Save(); err != nil {
		return nil, err
	}

	return mgr, nil
}

func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(m.filePath), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(m.Config)
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}
```

- [ ] **Step 4: Add yaml dependency and run test**

Run:
```bash
go get gopkg.in/yaml.v3
go test ./internal/config/ -v -run TestNewManager_CreatesDefaultConfig
```
Expected: PASS

- [ ] **Step 5: Write failing test for loading existing config**

Add to `internal/config/config_test.go`:

```go
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
```

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./internal/config/ -v -run TestNewManager_LoadsExistingConfig`
Expected: PASS (implementation already handles this)

- [ ] **Step 7: Commit**

```bash
git add internal/config/ go.mod go.sum
git commit -m "feat: add ConfigManager with load/save YAML support"
```

---

### Task 3: ConfigManager — Project & Command CRUD

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

- [ ] **Step 1: Write failing tests for AddProject, RemoveProject, AddCommand, RemoveCommand, AddToHistory**

Add to `internal/config/config_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/config/ -v -run "TestAdd|TestRemove"`
Expected: FAIL — methods don't exist

- [ ] **Step 3: Implement CRUD methods**

Add to `internal/config/config.go`:

```go
import (
	"fmt"
	// ... existing imports
)

func (m *Manager) AddProject(name, path string) error {
	for _, p := range m.Config.Projects {
		if p.Path == path {
			return fmt.Errorf("project with path %q already exists", path)
		}
	}

	m.Config.Projects = append(m.Config.Projects, Project{
		Name:     name,
		Path:     path,
		Commands: []Command{},
	})
	return m.Save()
}

func (m *Manager) RemoveProject(name string) error {
	for i, p := range m.Config.Projects {
		if p.Name == name {
			m.Config.Projects = append(m.Config.Projects[:i], m.Config.Projects[i+1:]...)
			if m.Config.LastActiveProject == name {
				m.Config.LastActiveProject = ""
			}
			return m.Save()
		}
	}
	return fmt.Errorf("project %q not found", name)
}

func (m *Manager) AddCommand(projectName, cmdName, cmd string) error {
	for i, p := range m.Config.Projects {
		if p.Name == projectName {
			m.Config.Projects[i].Commands = append(m.Config.Projects[i].Commands, Command{
				Name: cmdName,
				Cmd:  cmd,
			})
			m.addToHistory(cmd)
			return m.Save()
		}
	}
	return fmt.Errorf("project %q not found", projectName)
}

func (m *Manager) RemoveCommand(projectName, cmdName string) error {
	for i, p := range m.Config.Projects {
		if p.Name == projectName {
			for j, c := range p.Commands {
				if c.Name == cmdName {
					m.Config.Projects[i].Commands = append(p.Commands[:j], p.Commands[j+1:]...)
					return m.Save()
				}
			}
			return fmt.Errorf("command %q not found in project %q", cmdName, projectName)
		}
	}
	return fmt.Errorf("project %q not found", projectName)
}

func (m *Manager) addToHistory(cmd string) {
	for _, h := range m.Config.CommandHistory {
		if h == cmd {
			return
		}
	}
	m.Config.CommandHistory = append(m.Config.CommandHistory, cmd)
}

func (m *Manager) GetProject(name string) (*Project, error) {
	for i, p := range m.Config.Projects {
		if p.Name == name {
			return &m.Config.Projects[i], nil
		}
	}
	return nil, fmt.Errorf("project %q not found", name)
}

func (m *Manager) SetLastActiveProject(name string) error {
	m.Config.LastActiveProject = name
	return m.Save()
}
```

- [ ] **Step 4: Run all config tests**

Run: `go test ./internal/config/ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add project and command CRUD to ConfigManager"
```

---

### Task 4: ProcessManager — Spawn and Monitor

**Files:**
- Create: `internal/process/manager.go`
- Create: `internal/process/kill_windows.go`
- Create: `internal/process/kill_unix.go`
- Create: `internal/process/manager_test.go`

- [ ] **Step 1: Write failing test for starting and monitoring a process**

Create `internal/process/manager_test.go`:

```go
package process

import (
	"runtime"
	"testing"
	"time"
)

func TestStartCommand_RunsProcess(t *testing.T) {
	mgr := NewManager()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "ping -n 10 127.0.0.1"
	} else {
		cmd = "sleep 10"
	}

	err := mgr.Start("test-cmd", cmd, t.TempDir())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	status := mgr.Status("test-cmd")
	if status != Running {
		t.Errorf("expected Running, got %v", status)
	}

	// Cleanup
	mgr.KillAll()
}

func TestStartCommand_DetectsExit(t *testing.T) {
	mgr := NewManager()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "cmd /C echo hello"
	} else {
		cmd = "echo hello"
	}

	err := mgr.Start("test-cmd", cmd, t.TempDir())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for process to finish
	time.Sleep(500 * time.Millisecond)

	status := mgr.Status("test-cmd")
	if status != Stopped {
		t.Errorf("expected Stopped, got %v", status)
	}
}

func TestStartCommand_DetectsError(t *testing.T) {
	mgr := NewManager()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "cmd /C exit 1"
	} else {
		cmd = "sh -c 'exit 1'"
	}

	err := mgr.Start("test-cmd", cmd, t.TempDir())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	status := mgr.Status("test-cmd")
	if status != Errored {
		t.Errorf("expected Errored, got %v", status)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/process/ -v -run TestStartCommand`
Expected: FAIL — package doesn't exist

- [ ] **Step 3: Implement ProcessManager**

Create `internal/process/manager.go`:

```go
package process

import (
	"fmt"
	"os/exec"
	"runtime"
	"sync"
)

type Status int

const (
	Stopped Status = iota
	Running
	Errored
)

func (s Status) String() string {
	switch s {
	case Stopped:
		return "stopped"
	case Running:
		return "running"
	case Errored:
		return "errored"
	}
	return "unknown"
}

type CommandState struct {
	Name      string
	Cmd       string
	Status    Status
	ExitCode  int
	StatusMsg string
	process   *exec.Cmd
	pid       int
}

type StatusChangeFunc func(name string, status Status, exitCode int)

type Manager struct {
	commands map[string]*CommandState
	mu       sync.Mutex
	OnChange StatusChangeFunc
}

func NewManager() *Manager {
	return &Manager{
		commands: make(map[string]*CommandState),
	}
}

func (m *Manager) Start(name, cmd, workDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cs, exists := m.commands[name]; exists && cs.Status == Running {
		return fmt.Errorf("command %q is already running", name)
	}

	var proc *exec.Cmd
	if runtime.GOOS == "windows" {
		proc = exec.Command("cmd", "/C", cmd)
	} else {
		proc = exec.Command("sh", "-c", cmd)
	}
	proc.Dir = workDir

	if err := proc.Start(); err != nil {
		m.commands[name] = &CommandState{
			Name:      name,
			Cmd:       cmd,
			Status:    Errored,
			StatusMsg: fmt.Sprintf("failed to start: %v", err),
		}
		m.notifyChange(name, Errored, -1)
		return nil
	}

	cs := &CommandState{
		Name:    name,
		Cmd:     cmd,
		Status:  Running,
		process: proc,
		pid:     proc.Process.Pid,
	}
	m.commands[name] = cs

	go m.monitor(name, proc)

	return nil
}

func (m *Manager) monitor(name string, proc *exec.Cmd) {
	err := proc.Wait()

	m.mu.Lock()
	cs, exists := m.commands[name]
	if !exists {
		m.mu.Unlock()
		return
	}

	if err != nil {
		cs.Status = Errored
		if exitErr, ok := err.(*exec.ExitError); ok {
			cs.ExitCode = exitErr.ExitCode()
			cs.StatusMsg = fmt.Sprintf("exited (code %d)", cs.ExitCode)
		} else {
			cs.StatusMsg = err.Error()
		}
	} else {
		cs.Status = Stopped
		cs.ExitCode = 0
		cs.StatusMsg = "exited (code 0)"
	}
	status := cs.Status
	exitCode := cs.ExitCode
	m.mu.Unlock()

	m.notifyChange(name, status, exitCode)
}

func (m *Manager) notifyChange(name string, status Status, exitCode int) {
	if m.OnChange != nil {
		m.OnChange(name, status, exitCode)
	}
}

func (m *Manager) Kill(name string) error {
	m.mu.Lock()
	cs, exists := m.commands[name]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("command %q not found", name)
	}
	if cs.Status != Running {
		m.mu.Unlock()
		return nil
	}
	pid := cs.pid
	m.mu.Unlock()

	err := killProcessTree(pid)
	if err != nil {
		return fmt.Errorf("failed to kill %q: %w", name, err)
	}

	return nil
}

func (m *Manager) KillAll() {
	m.mu.Lock()
	names := make([]string, 0)
	for name, cs := range m.commands {
		if cs.Status == Running {
			names = append(names, name)
		}
	}
	m.mu.Unlock()

	for _, name := range names {
		m.Kill(name)
	}
}

func (m *Manager) Status(name string) Status {
	m.mu.Lock()
	defer m.mu.Unlock()

	cs, exists := m.commands[name]
	if !exists {
		return Stopped
	}
	return cs.Status
}

func (m *Manager) GetState(name string) *CommandState {
	m.mu.Lock()
	defer m.mu.Unlock()

	cs, exists := m.commands[name]
	if !exists {
		return nil
	}
	// Return a copy
	copy := *cs
	copy.process = nil
	return &copy
}

func (m *Manager) AllStates() []CommandState {
	m.mu.Lock()
	defer m.mu.Unlock()

	states := make([]CommandState, 0, len(m.commands))
	for _, cs := range m.commands {
		copy := *cs
		copy.process = nil
		states = append(states, copy)
	}
	return states
}

func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.commands, name)
}

func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands = make(map[string]*CommandState)
}
```

- [ ] **Step 4: Implement platform-specific kill**

Create `internal/process/kill_windows.go`:

```go
//go:build windows

package process

import (
	"fmt"
	"os/exec"
)

func killProcessTree(pid int) error {
	cmd := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", pid))
	return cmd.Run()
}
```

Create `internal/process/kill_unix.go`:

```go
//go:build !windows

package process

import (
	"syscall"
	"time"
)

func killProcessTree(pid int) error {
	// Send SIGTERM to process group
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		// Process may already be dead
		return nil
	}

	// Wait briefly, then force kill
	time.Sleep(2 * time.Second)
	syscall.Kill(-pid, syscall.SIGKILL)
	return nil
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/process/ -v -run TestStartCommand`
Expected: all PASS

- [ ] **Step 6: Write failing test for Kill**

Add to `internal/process/manager_test.go`:

```go
func TestKill_StopsRunningProcess(t *testing.T) {
	mgr := NewManager()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "ping -n 100 127.0.0.1"
	} else {
		cmd = "sleep 100"
	}

	mgr.Start("long-cmd", cmd, t.TempDir())

	time.Sleep(200 * time.Millisecond)
	if mgr.Status("long-cmd") != Running {
		t.Fatal("expected command to be running")
	}

	err := mgr.Kill("long-cmd")
	if err != nil {
		t.Fatalf("Kill failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	status := mgr.Status("long-cmd")
	if status == Running {
		t.Error("expected command to NOT be running after kill")
	}
}

func TestKillAll(t *testing.T) {
	mgr := NewManager()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "ping -n 100 127.0.0.1"
	} else {
		cmd = "sleep 100"
	}

	mgr.Start("cmd1", cmd, t.TempDir())
	mgr.Start("cmd2", cmd, t.TempDir())
	time.Sleep(200 * time.Millisecond)

	mgr.KillAll()
	time.Sleep(500 * time.Millisecond)

	if mgr.Status("cmd1") == Running {
		t.Error("cmd1 should not be running")
	}
	if mgr.Status("cmd2") == Running {
		t.Error("cmd2 should not be running")
	}
}
```

- [ ] **Step 7: Run Kill tests**

Run: `go test ./internal/process/ -v -run "TestKill"`
Expected: all PASS

- [ ] **Step 8: Commit**

```bash
git add internal/process/
git commit -m "feat: add ProcessManager with spawn, monitor, and tree kill"
```

---

### Task 5: SystrayManager — Icon States and Menu

**Files:**
- Create: `internal/tray/tray.go`
- Create: `internal/tray/icons.go`
- Create: `icons/grey.ico`, `icons/green.ico`, `icons/yellow.ico`, `icons/red.ico`

- [ ] **Step 1: Generate simple colored ICO files**

Create a small Go script `cmd/icongen/main.go` that generates 16x16 single-color ICO files:

```go
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	colors := map[string]color.RGBA{
		"icons/grey.png":   {R: 128, G: 128, B: 128, A: 255},
		"icons/green.png":  {R: 34, G: 197, B: 94, A: 255},
		"icons/yellow.png": {R: 234, G: 179, B: 8, A: 255},
		"icons/red.png":    {R: 220, G: 38, B: 38, A: 255},
	}

	os.MkdirAll("icons", 0755)

	for path, c := range colors {
		img := image.NewRGBA(image.Rect(0, 0, 64, 64))
		for y := 0; y < 64; y++ {
			for x := 0; x < 64; x++ {
				// Draw a circle
				dx := float64(x) - 31.5
				dy := float64(y) - 31.5
				if dx*dx+dy*dy <= 28*28 {
					img.Set(x, y, c)
				}
			}
		}
		f, _ := os.Create(path)
		png.Encode(f, img)
		f.Close()
	}
}
```

Run:
```bash
go run cmd/icongen/main.go
```

Expected: 4 PNG files in `icons/`

- [ ] **Step 2: Create icons.go with embedded icon bytes**

Create `internal/tray/icons.go`:

```go
package tray

import (
	_ "embed"
)

//go:embed ../../icons/grey.png
var IconGrey []byte

//go:embed ../../icons/green.png
var IconGreen []byte

//go:embed ../../icons/yellow.png
var IconYellow []byte

//go:embed ../../icons/red.png
var IconRed []byte
```

- [ ] **Step 3: Implement SystrayManager**

Create `internal/tray/tray.go`:

```go
package tray

import (
	"github.com/energye/systray"
)

type IconState int

const (
	StateIdle    IconState = iota // grey — no project
	StateAllUp                   // green — all running
	StatePartial                 // yellow — some down
	StateAllDown                 // red — all down
)

type Callbacks struct {
	OnLeftClick func()
	OnQuit      func()
}

type Manager struct {
	callbacks Callbacks
	state     IconState
	quitItem  *systray.MenuItem
}

func NewManager(cb Callbacks) *Manager {
	return &Manager{
		callbacks: cb,
		state:     StateIdle,
	}
}

func (m *Manager) Setup() {
	systray.SetIcon(IconGrey)
	systray.SetTitle("DevProc")
	systray.SetTooltip("DevProc — No project selected")

	systray.SetOnClick(func(menu systray.IMenu) {
		if m.callbacks.OnLeftClick != nil {
			m.callbacks.OnLeftClick()
		}
	})

	systray.SetOnRClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})

	m.quitItem = systray.AddMenuItem("Quit", "Quit DevProc")
	m.quitItem.Click(func() {
		if m.callbacks.OnQuit != nil {
			m.callbacks.OnQuit()
		}
	})
}

func (m *Manager) SetState(state IconState, tooltip string) {
	m.state = state
	systray.SetTooltip(tooltip)

	switch state {
	case StateIdle:
		systray.SetIcon(IconGrey)
	case StateAllUp:
		systray.SetIcon(IconGreen)
	case StatePartial:
		systray.SetIcon(IconYellow)
	case StateAllDown:
		systray.SetIcon(IconRed)
	}
}

func (m *Manager) GetState() IconState {
	return m.state
}
```

- [ ] **Step 4: Install systray dependency**

Run:
```bash
go get github.com/energye/systray
```

- [ ] **Step 5: Verify compilation**

Run: `go build ./internal/tray/`
Expected: compiles without errors

- [ ] **Step 6: Commit**

```bash
git add internal/tray/ icons/ cmd/icongen/
git commit -m "feat: add SystrayManager with colored icon states"
```

---

### Task 6: App Struct — Wails Bindings

**Files:**
- Modify: `app.go`
- Modify: `main.go`

- [ ] **Step 1: Implement App struct with all Wails-bound methods**

Replace `app.go` with:

```go
package main

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"devproc/internal/config"
	"devproc/internal/process"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type CommandStatus struct {
	Name      string `json:"name"`
	Cmd       string `json:"cmd"`
	Status    string `json:"status"`
	StatusMsg string `json:"statusMsg"`
}

type AppState struct {
	ActiveProject string          `json:"activeProject"`
	Projects      []config.Project `json:"projects"`
	Commands      []CommandStatus `json:"commands"`
	CommandHistory []string       `json:"commandHistory"`
}

type App struct {
	ctx     context.Context
	config  *config.Manager
	process *process.Manager
}

func NewApp(cfgPath string) *App {
	cfg, err := config.NewManager(cfgPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	procMgr := process.NewManager()

	app := &App{
		config:  cfg,
		process: procMgr,
	}

	procMgr.OnChange = func(name string, status process.Status, exitCode int) {
		app.emitStateUpdate()
		app.updateTrayIcon()
	}

	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetState() AppState {
	state := AppState{
		ActiveProject:  a.config.Config.LastActiveProject,
		Projects:       a.config.Config.Projects,
		CommandHistory: a.config.Config.CommandHistory,
		Commands:       []CommandStatus{},
	}

	for _, cs := range a.process.AllStates() {
		state.Commands = append(state.Commands, CommandStatus{
			Name:      cs.Name,
			Cmd:       cs.Cmd,
			Status:    cs.Status.String(),
			StatusMsg: cs.StatusMsg,
		})
	}

	return state
}

func (a *App) SelectProject(name string) error {
	// Kill all current processes
	a.process.KillAll()
	a.process.Clear()

	if name == "" {
		a.config.SetLastActiveProject("")
		a.emitStateUpdate()
		a.updateTrayIcon()
		return nil
	}

	proj, err := a.config.GetProject(name)
	if err != nil {
		return err
	}

	a.config.SetLastActiveProject(name)

	// Start all commands for this project
	for _, cmd := range proj.Commands {
		a.process.Start(cmd.Name, cmd.Cmd, proj.Path)
	}

	a.emitStateUpdate()
	a.updateTrayIcon()
	return nil
}

func (a *App) AddProject() (string, error) {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
	if err != nil || dir == "" {
		return "", err
	}

	name := filepath.Base(dir)
	if err := a.config.AddProject(name, dir); err != nil {
		return "", err
	}

	a.emitStateUpdate()
	return name, nil
}

func (a *App) RemoveProject(name string) error {
	if a.config.Config.LastActiveProject == name {
		a.process.KillAll()
		a.process.Clear()
		a.config.SetLastActiveProject("")
	}

	err := a.config.RemoveProject(name)
	a.emitStateUpdate()
	a.updateTrayIcon()
	return err
}

func (a *App) AddCommand(projectName, cmdName, cmd string) error {
	err := a.config.AddCommand(projectName, cmdName, cmd)
	if err != nil {
		return err
	}

	// If this is the active project, start the command immediately
	if a.config.Config.LastActiveProject == projectName {
		proj, _ := a.config.GetProject(projectName)
		a.process.Start(cmdName, cmd, proj.Path)
	}

	a.emitStateUpdate()
	return nil
}

func (a *App) RemoveCommand(projectName, cmdName string) error {
	// Kill the command if running
	a.process.Kill(cmdName)
	a.process.Remove(cmdName)

	err := a.config.RemoveCommand(projectName, cmdName)
	a.emitStateUpdate()
	a.updateTrayIcon()
	return err
}

func (a *App) StartCommand(name string) error {
	projName := a.config.Config.LastActiveProject
	proj, err := a.config.GetProject(projName)
	if err != nil {
		return err
	}

	for _, cmd := range proj.Commands {
		if cmd.Name == name {
			a.process.Kill(name)
			a.process.Remove(name)
			return a.process.Start(name, cmd.Cmd, proj.Path)
		}
	}
	return fmt.Errorf("command %q not found", name)
}

func (a *App) KillCommand(name string) error {
	return a.process.Kill(name)
}

func (a *App) StartAll() {
	projName := a.config.Config.LastActiveProject
	proj, _ := a.config.GetProject(projName)
	if proj == nil {
		return
	}
	for _, cmd := range proj.Commands {
		state := a.process.GetState(cmd.Name)
		if state == nil || state.Status != process.Running {
			a.process.Remove(cmd.Name)
			a.process.Start(cmd.Name, cmd.Cmd, proj.Path)
		}
	}
}

func (a *App) KillAll() {
	a.process.KillAll()
}

func (a *App) Shutdown() {
	a.process.KillAll()
}

func (a *App) HideWindow() {
	wailsRuntime.Hide(a.ctx)
}

func (a *App) emitStateUpdate() {
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "state-update", a.GetState())
	}
}

func (a *App) updateTrayIcon() {
	// This will be called from main.go where tray manager is accessible
	// We emit an event that main.go listens to
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "tray-update")
	}
}

func configPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(mustGetenv("APPDATA"), "devproc", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "devproc", "config.yaml")
}
```

- [ ] **Step 2: Update main.go to wire everything together**

Replace `main.go` with:

```go
package main

import (
	"embed"
	"os"
	"path/filepath"
	"runtime"

	"devproc/internal/process"
	"devproc/internal/tray"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func getConfigPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "devproc", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "devproc", "config.yaml")
}

func main() {
	app := NewApp(getConfigPath())

	var trayMgr *tray.Manager

	systray.Run(func() {
		trayMgr = tray.NewManager(tray.Callbacks{
			OnLeftClick: func() {
				if app.ctx != nil {
					wailsRuntime.Show(app.ctx)
				}
			},
			OnQuit: func() {
				app.Shutdown()
				wailsRuntime.Quit(app.ctx)
			},
		})
		trayMgr.Setup()

		// Listen for tray update events
		app.process.OnChange = func(name string, status process.Status, exitCode int) {
			app.emitStateUpdate()
			updateTrayState(trayMgr, app)
		}

		go func() {
			err := wails.Run(&options.App{
				Title:            "DevProc",
				Width:            320,
				Height:           480,
				Frameless:        true,
				StartHidden:      true,
				HideWindowOnClose: true,
				AssetServer: &assetserver.Options{
					Assets: assets,
				},
				OnStartup: app.startup,
				OnShutdown: func(ctx context.Context) {
					app.Shutdown()
				},
				Bind: []interface{}{
					app,
				},
			})
			if err != nil {
				println("Error:", err.Error())
				os.Exit(1)
			}
			systray.Quit()
		}()
	}, func() {
		// Cleanup on systray exit
	})
}

func updateTrayState(trayMgr *tray.Manager, app *App) {
	if app.config.Config.LastActiveProject == "" {
		trayMgr.SetState(tray.StateIdle, "DevProc — No project selected")
		return
	}

	states := app.process.AllStates()
	if len(states) == 0 {
		trayMgr.SetState(tray.StateIdle, "DevProc — No commands")
		return
	}

	running := 0
	for _, s := range states {
		if s.Status == process.Running {
			running++
		}
	}

	projName := app.config.Config.LastActiveProject
	switch {
	case running == len(states):
		trayMgr.SetState(tray.StateAllUp, "DevProc — "+projName+" (all running)")
	case running == 0:
		trayMgr.SetState(tray.StateAllDown, "DevProc — "+projName+" (all stopped)")
	default:
		trayMgr.SetState(tray.StatePartial, fmt.Sprintf("DevProc — %s (%d/%d running)", projName, running, len(states)))
	}
}
```

- [ ] **Step 3: Remove stale helpers from app.go**

Remove the `configPath()` and `mustGetenv()` functions from `app.go` — config path is now resolved in `main.go` and passed to `NewApp()`. Also add the missing `"os"` import to `main.go`.

- [ ] **Step 4: Verify compilation**

Run: `go build .`
Expected: compiles (frontend dist won't exist yet, that's fine for now — we'll do a full `wails build` later)

- [ ] **Step 5: Commit**

```bash
git add app.go main.go
git commit -m "feat: wire App struct with Wails bindings, systray, and process manager"
```

---

### Task 7: Frontend — Global Styles and TitleBar

**Files:**
- Modify: `frontend/src/app.css`
- Create: `frontend/src/lib/TitleBar.svelte`
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: Write global dark brutalist CSS**

Replace `frontend/src/app.css` with:

```css
:root {
  --bg: #0a0a0a;
  --surface: #111111;
  --border: #333333;
  --text: #e0e0e0;
  --text-dim: #888888;
  --text-muted: #555555;
  --green: #22c55e;
  --red: #dc2626;
  --yellow: #eab308;
  --font-mono: 'Courier New', 'Consolas', monospace;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  background: var(--bg);
  color: var(--text);
  font-family: var(--font-mono);
  font-size: 12px;
  width: 320px;
  overflow-x: hidden;
  user-select: none;
  -webkit-user-select: none;
}

button {
  font-family: var(--font-mono);
  cursor: pointer;
  border: none;
  text-transform: uppercase;
  letter-spacing: 1px;
  font-size: 9px;
  font-weight: bold;
  padding: 4px 10px;
}

button:active {
  opacity: 0.8;
}

.label {
  font-size: 9px;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--text-muted);
  margin-bottom: 4px;
}

/* Scrollbar */
::-webkit-scrollbar {
  width: 4px;
}
::-webkit-scrollbar-track {
  background: var(--bg);
}
::-webkit-scrollbar-thumb {
  background: var(--border);
}
```

- [ ] **Step 2: Create TitleBar component**

Create `frontend/src/lib/TitleBar.svelte`:

```svelte
<script lang="ts">
  import { HideWindow } from '../../wailsjs/go/main/App';

  // Enable window dragging on the title bar
  function onMouseDown(e: MouseEvent) {
    if ((e.target as HTMLElement).closest('.close-btn')) return;
    (window as any).runtime.WindowStartDrag();
  }
</script>

<div class="titlebar" onmousedown={onMouseDown} role="banner">
  <span class="title">DEVPROC</span>
  <button class="close-btn" onclick={() => HideWindow()}>✕</button>
</div>

<style>
  .titlebar {
    background: #1a1a1a;
    padding: 8px 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 2px solid var(--border);
    cursor: grab;
  }
  .titlebar:active {
    cursor: grabbing;
  }
  .title {
    font-size: 11px;
    font-weight: bold;
    text-transform: uppercase;
    letter-spacing: 2px;
    color: var(--text-dim);
  }
  .close-btn {
    background: transparent;
    color: var(--text-muted);
    padding: 2px 6px;
    font-size: 10px;
  }
  .close-btn:hover {
    color: var(--red);
  }
</style>
```

- [ ] **Step 3: Update App.svelte shell**

Replace `frontend/src/App.svelte` with:

```svelte
<script lang="ts">
  import './app.css';
  import TitleBar from './lib/TitleBar.svelte';
</script>

<TitleBar />
<main>
  <p style="color: var(--text-muted); padding: 12px;">Loading...</p>
</main>
```

- [ ] **Step 4: Build frontend to verify**

Run:
```bash
cd frontend && npm run build && cd ..
```
Expected: builds without errors

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app.css frontend/src/lib/TitleBar.svelte frontend/src/App.svelte
git commit -m "feat: add dark brutalist theme and TitleBar component"
```

---

### Task 8: Frontend — Reactive State Store

**Files:**
- Create: `frontend/src/lib/state.svelte.ts`

- [ ] **Step 1: Create the Svelte 5 state store**

Create `frontend/src/lib/state.svelte.ts`:

```typescript
import { GetState, SelectProject, AddProject, RemoveProject, AddCommand, RemoveCommand, StartCommand, KillCommand, StartAll, KillAll } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

export interface CommandStatus {
  name: string;
  cmd: string;
  status: 'stopped' | 'running' | 'errored';
  statusMsg: string;
}

export interface Project {
  name: string;
  path: string;
  commands: { name: string; cmd: string }[];
}

export interface AppState {
  activeProject: string;
  projects: Project[];
  commands: CommandStatus[];
  commandHistory: string[];
}

let appState = $state<AppState>({
  activeProject: '',
  projects: [],
  commands: [],
  commandHistory: [],
});

export function getAppState(): AppState {
  return appState;
}

export async function loadState() {
  const state = await GetState();
  appState.activeProject = state.activeProject;
  appState.projects = state.projects;
  appState.commands = state.commands;
  appState.commandHistory = state.commandHistory;
}

export async function selectProject(name: string) {
  await SelectProject(name);
}

export async function addProject() {
  await AddProject();
}

export async function removeProject(name: string) {
  await RemoveProject(name);
}

export async function addCommand(projectName: string, cmdName: string, cmd: string) {
  await AddCommand(projectName, cmdName, cmd);
}

export async function removeCommand(projectName: string, cmdName: string) {
  await RemoveCommand(projectName, cmdName);
}

export async function startCommand(name: string) {
  await StartCommand(name);
}

export async function killCommand(name: string) {
  await KillCommand(name);
}

export async function startAll() {
  await StartAll();
}

export async function killAll() {
  await KillAll();
}

export function initEventListeners() {
  EventsOn('state-update', (state: AppState) => {
    appState.activeProject = state.activeProject;
    appState.projects = state.projects;
    appState.commands = state.commands;
    appState.commandHistory = state.commandHistory;
  });
}
```

- [ ] **Step 2: Verify frontend compiles**

Run:
```bash
cd frontend && npm run build && cd ..
```
Expected: builds (Wails JS bindings will be auto-generated on `wails build`)

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/state.svelte.ts
git commit -m "feat: add Svelte 5 reactive state store with Wails bindings"
```

---

### Task 9: Frontend — ProjectSelector Component

**Files:**
- Create: `frontend/src/lib/ProjectSelector.svelte`

- [ ] **Step 1: Create ProjectSelector component**

Create `frontend/src/lib/ProjectSelector.svelte`:

```svelte
<script lang="ts">
  import { getAppState, selectProject, addProject, removeProject } from './state.svelte.ts';

  let showDropdown = $state(false);

  const state = $derived(getAppState());

  const activeProjectData = $derived(
    state.projects.find((p) => p.name === state.activeProject)
  );

  async function handleSelect(name: string) {
    showDropdown = false;
    await selectProject(name);
  }

  async function handleAddProject() {
    showDropdown = false;
    await addProject();
  }

  async function handleDeselect() {
    showDropdown = false;
    await selectProject('');
  }

  function handleRemove(e: MouseEvent, name: string) {
    e.stopPropagation();
    removeProject(name);
  }
</script>

<div class="project-selector">
  <div class="label">PROJECT</div>
  <button class="selector" onclick={() => (showDropdown = !showDropdown)}>
    <span>{state.activeProject || 'No project selected'}</span>
    <span class="arrow">{showDropdown ? '▴' : '▾'}</span>
  </button>
  {#if activeProjectData}
    <div class="path">{activeProjectData.path}</div>
  {/if}

  {#if showDropdown}
    <div class="dropdown">
      {#if state.activeProject}
        <button class="dropdown-item deselect" onclick={handleDeselect}>
          ✕ Deselect project
        </button>
      {/if}
      {#each state.projects as project}
        <button
          class="dropdown-item"
          class:active={project.name === state.activeProject}
          onclick={() => handleSelect(project.name)}
        >
          <span>{project.name}</span>
          <span
            class="remove"
            role="button"
            tabindex="-1"
            onclick={(e) => handleRemove(e, project.name)}
            onkeydown={() => {}}
          >✕</span>
        </button>
      {/each}
      <button class="dropdown-item add" onclick={handleAddProject}>
        + Add project...
      </button>
    </div>
  {/if}
</div>

<style>
  .project-selector {
    padding: 10px 12px;
    border-bottom: 2px solid var(--border);
    position: relative;
  }
  .selector {
    width: 100%;
    background: var(--surface);
    border: 1px solid #444;
    color: var(--text);
    padding: 6px 10px;
    font-size: 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    text-transform: none;
    letter-spacing: 0;
  }
  .arrow {
    color: var(--text-muted);
  }
  .path {
    font-size: 9px;
    color: var(--text-muted);
    margin-top: 3px;
  }
  .dropdown {
    position: absolute;
    left: 12px;
    right: 12px;
    top: 100%;
    background: var(--surface);
    border: 1px solid #444;
    z-index: 10;
    max-height: 200px;
    overflow-y: auto;
  }
  .dropdown-item {
    width: 100%;
    background: transparent;
    color: var(--text);
    padding: 6px 10px;
    font-size: 11px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    text-transform: none;
    letter-spacing: 0;
  }
  .dropdown-item:hover {
    background: #1a1a1a;
  }
  .dropdown-item.active {
    color: var(--green);
  }
  .dropdown-item.add {
    color: var(--text-dim);
    border-top: 1px solid var(--border);
  }
  .dropdown-item.deselect {
    color: var(--text-muted);
    border-bottom: 1px solid var(--border);
  }
  .remove {
    color: var(--text-muted);
    font-size: 9px;
  }
  .remove:hover {
    color: var(--red);
  }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/lib/ProjectSelector.svelte
git commit -m "feat: add ProjectSelector dropdown component"
```

---

### Task 10: Frontend — CommandItem and CommandList

**Files:**
- Create: `frontend/src/lib/CommandItem.svelte`
- Create: `frontend/src/lib/CommandList.svelte`

- [ ] **Step 1: Create CommandItem component**

Create `frontend/src/lib/CommandItem.svelte`:

```svelte
<script lang="ts">
  import { startCommand, killCommand, removeCommand } from './state.svelte.ts';
  import type { CommandStatus } from './state.svelte.ts';

  interface Props {
    command: CommandStatus;
    projectName: string;
  }

  let { command, projectName }: Props = $props();

  let showDelete = $state(false);

  function handleAction() {
    if (command.status === 'running') {
      killCommand(command.name);
    } else {
      startCommand(command.name);
    }
  }

  function handleDelete() {
    removeCommand(projectName, command.name);
  }

  const dotClass = $derived(
    command.status === 'running' ? 'dot green' :
    command.status === 'errored' ? 'dot red' : 'dot grey'
  );

  const buttonLabel = $derived(
    command.status === 'running' ? 'KILL' :
    command.status === 'errored' ? 'RESTART' : 'START'
  );

  const buttonClass = $derived(
    command.status === 'running' ? 'btn-kill' :
    command.status === 'errored' ? 'btn-restart' : 'btn-start'
  );

  const statusText = $derived(
    command.status === 'running' ? command.cmd :
    command.statusMsg || command.cmd
  );

  const statusClass = $derived(
    command.status === 'errored' ? 'status errored' : 'status'
  );
</script>

<div
  class="command-item"
  onmouseenter={() => (showDelete = true)}
  onmouseleave={() => (showDelete = false)}
  role="listitem"
>
  <div class="left">
    <span class={dotClass}></span>
    <div class="info">
      <div class="name">{command.name}</div>
      <div class={statusClass}>{statusText}</div>
    </div>
  </div>
  <div class="right">
    {#if showDelete}
      <button class="btn-delete" onclick={handleDelete}>✕</button>
    {/if}
    <button class={buttonClass} onclick={handleAction}>{buttonLabel}</button>
  </div>
</div>

<style>
  .command-item {
    background: var(--surface);
    border: 1px solid #2a2a2a;
    padding: 8px 10px;
    margin-bottom: 4px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .left {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .dot.green {
    background: var(--green);
    box-shadow: 0 0 6px var(--green);
  }
  .dot.red {
    background: var(--red);
    box-shadow: 0 0 6px var(--red);
  }
  .dot.grey {
    background: #666;
  }
  .info {
    min-width: 0;
  }
  .name {
    font-size: 12px;
    font-weight: bold;
  }
  .status {
    font-size: 9px;
    color: var(--text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .status.errored {
    color: var(--red);
  }
  .right {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
  }
  .btn-kill {
    background: var(--red);
    color: white;
  }
  .btn-start {
    background: var(--green);
    color: black;
  }
  .btn-restart {
    background: var(--yellow);
    color: black;
  }
  .btn-delete {
    background: transparent;
    color: var(--text-muted);
    padding: 2px 6px;
    font-size: 9px;
  }
  .btn-delete:hover {
    color: var(--red);
  }
</style>
```

- [ ] **Step 2: Create CommandList component**

Create `frontend/src/lib/CommandList.svelte`:

```svelte
<script lang="ts">
  import { getAppState } from './state.svelte.ts';
  import CommandItem from './CommandItem.svelte';

  const state = $derived(getAppState());
</script>

<div class="command-list">
  <div class="label">COMMANDS</div>

  {#if !state.activeProject}
    <div class="empty">Select a project to see commands</div>
  {:else if state.commands.length === 0}
    <div class="empty">No commands configured</div>
  {:else}
    {#each state.commands as command (command.name)}
      <CommandItem {command} projectName={state.activeProject} />
    {/each}
  {/if}
</div>

<style>
  .command-list {
    padding: 8px 12px;
    flex: 1;
    overflow-y: auto;
  }
  .empty {
    color: var(--text-muted);
    font-size: 10px;
    padding: 12px 0;
    text-align: center;
  }
</style>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/CommandItem.svelte frontend/src/lib/CommandList.svelte
git commit -m "feat: add CommandItem and CommandList components"
```

---

### Task 11: Frontend — AddCommandDialog and BottomBar

**Files:**
- Create: `frontend/src/lib/AddCommandDialog.svelte`
- Create: `frontend/src/lib/BottomBar.svelte`

- [ ] **Step 1: Create AddCommandDialog component**

Create `frontend/src/lib/AddCommandDialog.svelte`:

```svelte
<script lang="ts">
  import { getAppState, addCommand } from './state.svelte.ts';

  interface Props {
    onclose: () => void;
  }

  let { onclose }: Props = $props();

  const state = $derived(getAppState());

  let cmdName = $state('');
  let cmdStr = $state('');

  function selectHistory(cmd: string) {
    cmdStr = cmd;
  }

  async function handleSave() {
    if (!cmdName.trim() || !cmdStr.trim()) return;
    await addCommand(state.activeProject, cmdName.trim(), cmdStr.trim());
    onclose();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose();
    if (e.key === 'Enter') handleSave();
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="overlay" onclick={onclose} onkeydown={handleKeydown}>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="dialog" onclick|stopPropagation onkeydown={handleKeydown}>
    <div class="dialog-header">
      <span>ADD COMMAND</span>
      <button class="close" onclick={onclose}>✕</button>
    </div>

    <div class="dialog-body">
      <div class="field">
        <div class="label">NAME</div>
        <input
          type="text"
          bind:value={cmdName}
          placeholder="e.g. Vite"
          class="input"
        />
      </div>

      <div class="field">
        <div class="label">COMMAND</div>
        <input
          type="text"
          bind:value={cmdStr}
          placeholder="e.g. npm run dev"
          class="input"
        />
      </div>

      {#if state.commandHistory.length > 0}
        <div class="field">
          <div class="label">HISTORY</div>
          <div class="history">
            {#each state.commandHistory as cmd}
              <button class="history-item" onclick={() => selectHistory(cmd)}>
                {cmd}
              </button>
            {/each}
          </div>
        </div>
      {/if}
    </div>

    <div class="dialog-footer">
      <button class="btn-cancel" onclick={onclose}>CANCEL</button>
      <button class="btn-save" onclick={handleSave}>SAVE</button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }
  .dialog {
    background: var(--bg);
    border: 2px solid var(--border);
    width: 290px;
  }
  .dialog-header {
    background: #1a1a1a;
    padding: 8px 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 2px solid var(--border);
    font-size: 10px;
    font-weight: bold;
    letter-spacing: 2px;
    color: var(--text-dim);
  }
  .close {
    background: transparent;
    color: var(--text-muted);
    font-size: 10px;
    padding: 2px 6px;
  }
  .dialog-body {
    padding: 12px;
  }
  .field {
    margin-bottom: 10px;
  }
  .input {
    width: 100%;
    background: var(--surface);
    border: 1px solid #444;
    color: var(--text);
    padding: 6px 10px;
    font-family: var(--font-mono);
    font-size: 12px;
    outline: none;
  }
  .input:focus {
    border-color: var(--text-dim);
  }
  .history {
    max-height: 100px;
    overflow-y: auto;
    border: 1px solid #444;
  }
  .history-item {
    width: 100%;
    background: transparent;
    color: var(--text);
    padding: 4px 10px;
    font-size: 11px;
    text-align: left;
    text-transform: none;
    letter-spacing: 0;
  }
  .history-item:hover {
    background: #1a1a1a;
  }
  .dialog-footer {
    padding: 8px 12px;
    border-top: 2px solid var(--border);
    display: flex;
    justify-content: flex-end;
    gap: 6px;
  }
  .btn-cancel {
    background: transparent;
    color: var(--text-dim);
    border: 1px solid #444;
  }
  .btn-save {
    background: var(--green);
    color: black;
  }
</style>
```

- [ ] **Step 2: Create BottomBar component**

Create `frontend/src/lib/BottomBar.svelte`:

```svelte
<script lang="ts">
  import { getAppState, startAll, killAll } from './state.svelte.ts';

  interface Props {
    onAddCommand: () => void;
  }

  let { onAddCommand }: Props = $props();

  const state = $derived(getAppState());
</script>

<div class="bottom-bar">
  <button class="btn-add" onclick={onAddCommand} disabled={!state.activeProject}>
    + ADD CMD
  </button>
  <div class="actions">
    <button class="btn-start-all" onclick={() => startAll()} disabled={!state.activeProject}>
      START ALL
    </button>
    <button class="btn-kill-all" onclick={() => killAll()} disabled={!state.activeProject}>
      KILL ALL
    </button>
  </div>
</div>

<style>
  .bottom-bar {
    border-top: 2px solid var(--border);
    padding: 8px 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .btn-add {
    background: transparent;
    color: var(--text-dim);
    border: 1px solid #444;
  }
  .btn-add:hover:not(:disabled) {
    color: var(--text);
    border-color: var(--text-dim);
  }
  .actions {
    display: flex;
    gap: 6px;
  }
  .btn-start-all {
    background: var(--green);
    color: black;
  }
  .btn-kill-all {
    background: var(--red);
    color: white;
  }
  button:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }
</style>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/AddCommandDialog.svelte frontend/src/lib/BottomBar.svelte
git commit -m "feat: add AddCommandDialog and BottomBar components"
```

---

### Task 12: Frontend — Wire Everything in App.svelte

**Files:**
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: Update App.svelte to assemble all components**

Replace `frontend/src/App.svelte` with:

```svelte
<script lang="ts">
  import './app.css';
  import { onMount } from 'svelte';
  import { loadState, initEventListeners } from './lib/state.svelte.ts';
  import TitleBar from './lib/TitleBar.svelte';
  import ProjectSelector from './lib/ProjectSelector.svelte';
  import CommandList from './lib/CommandList.svelte';
  import BottomBar from './lib/BottomBar.svelte';
  import AddCommandDialog from './lib/AddCommandDialog.svelte';

  let showAddDialog = $state(false);

  onMount(() => {
    loadState();
    initEventListeners();
  });
</script>

<div class="app">
  <TitleBar />
  <ProjectSelector />
  <CommandList />
  <BottomBar onAddCommand={() => (showAddDialog = true)} />

  {#if showAddDialog}
    <AddCommandDialog onclose={() => (showAddDialog = false)} />
  {/if}
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    height: 100vh;
    width: 320px;
  }
</style>
```

- [ ] **Step 2: Build frontend**

Run:
```bash
cd frontend && npm run build && cd ..
```
Expected: builds without errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/App.svelte
git commit -m "feat: wire all components in App.svelte"
```

---

### Task 13: Full Build and Smoke Test

**Files:**
- No new files — integration verification

- [ ] **Step 1: Generate Wails bindings and build**

Run:
```bash
wails build
```
Expected: produces `build/bin/devproc.exe`

- [ ] **Step 2: Fix any compilation errors**

If `wails build` fails, read the error output and fix. Common issues:
- Missing imports in `main.go` (add `"context"`, `"fmt"` as needed)
- Wails JS bindings not generated yet — `wails build` generates them on first run

- [ ] **Step 3: Run the app manually**

Run: `./build/bin/devproc.exe`

Verify:
- Systray icon appears (grey dot)
- Clicking systray icon shows the popup window
- Window is frameless, dark theme
- "No project selected" is shown
- Close button hides window (doesn't quit)
- Right-click systray → Quit exits the app

- [ ] **Step 4: Test basic workflow**

1. Click "Add project" in the dropdown → directory picker opens
2. Select a project folder → project appears in dropdown
3. Click "+ ADD CMD" → dialog opens
4. Enter name "Test" and command "ping -n 100 127.0.0.1" → save
5. Command should start running (green dot)
6. Click KILL → command stops (grey dot)
7. Click START → command restarts
8. Systray icon reflects status (green when running, red when stopped)

- [ ] **Step 5: Test project switching**

1. Add a second project with a command
2. Switch between projects via dropdown
3. Verify: old project's commands are killed, new project's commands start
4. Verify: switching to "Deselect project" kills everything, systray goes grey

- [ ] **Step 6: Commit any fixes**

```bash
git add -A
git commit -m "fix: resolve build and integration issues from smoke test"
```

---

### Task 14: Final Cleanup

**Files:**
- Remove: `cmd/icongen/main.go` (one-time script, icons are already generated)
- Remove: Wails template boilerplate files that aren't needed

- [ ] **Step 1: Remove icon generator script**

```bash
rm -rf cmd/
```

- [ ] **Step 2: Remove any unused Wails template files**

Check for and remove any default Wails template files that were replaced (e.g., default `greet.go` if it still exists).

- [ ] **Step 3: Run all Go tests**

Run: `go test ./... -v`
Expected: all tests pass

- [ ] **Step 4: Final build**

Run: `wails build`
Expected: clean build, no warnings

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "chore: remove boilerplate and verify clean build"
```
