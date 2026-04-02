package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
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
	logFile   *os.File
}

type StatusChangeFunc func(name string, status Status, exitCode int)

type Manager struct {
	commands map[string]*CommandState
	mu       sync.Mutex
	OnChange StatusChangeFunc
	LogDir   string
}

func NewManager(logDir string) *Manager {
	if logDir != "" {
		os.MkdirAll(logDir, 0755)
	}
	return &Manager{
		commands: make(map[string]*CommandState),
		LogDir:   logDir,
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

	// Pipe stdout/stderr to a log file
	var logFile *os.File
	if m.LogDir != "" {
		safeName := strings.ReplaceAll(name, "/", "_")
		safeName = strings.ReplaceAll(safeName, "\\", "_")
		safeName = strings.ReplaceAll(safeName, " ", "_")
		logPath := filepath.Join(m.LogDir, fmt.Sprintf("%s_%s.log", safeName, time.Now().Format("2006-01-02")))
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			fmt.Fprintf(f, "\n=== %s started at %s ===\n", name, time.Now().Format("15:04:05"))
			fmt.Fprintf(f, "cmd: %s\n", cmd)
			fmt.Fprintf(f, "dir: %s\n\n", workDir)
			proc.Stdout = f
			proc.Stderr = f
			logFile = f
		}
	}

	if err := proc.Start(); err != nil {
		if logFile != nil {
			fmt.Fprintf(logFile, "=== failed to start: %v ===\n", err)
			logFile.Close()
		}
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
		logFile: logFile,
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

	if cs.logFile != nil {
		fmt.Fprintf(cs.logFile, "\n=== %s at %s ===\n", cs.StatusMsg, time.Now().Format("15:04:05"))
		cs.logFile.Close()
		cs.logFile = nil
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
