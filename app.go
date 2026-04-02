package main

import (
	"context"
	"fmt"
	"path/filepath"

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
	ActiveProject  string           `json:"activeProject"`
	Projects       []config.Project `json:"projects"`
	Commands       []CommandStatus  `json:"commands"`
	CommandHistory []string         `json:"commandHistory"`
}

type App struct {
	ctx          context.Context
	config       *config.Manager
	process      *process.Manager
	OnTrayUpdate func()
}

func NewApp(cfgPath, logDir string) *App {
	cfg, err := config.NewManager(cfgPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	procMgr := process.NewManager(logDir)

	app := &App{
		config:  cfg,
		process: procMgr,
	}

	procMgr.OnChange = func(name string, status process.Status, exitCode int) {
		app.emitStateUpdate()
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

	// Build commands from config for the active project,
	// overlaying with live process state where available
	if proj, err := a.config.GetProject(state.ActiveProject); err == nil {
		for _, cmd := range proj.Commands {
			cs := a.process.GetState(cmd.Name)
			if cs != nil {
				state.Commands = append(state.Commands, CommandStatus{
					Name:      cs.Name,
					Cmd:       cs.Cmd,
					Status:    cs.Status.String(),
					StatusMsg: cs.StatusMsg,
				})
			} else {
				state.Commands = append(state.Commands, CommandStatus{
					Name:   cmd.Name,
					Cmd:    cmd.Cmd,
					Status: "stopped",
				})
			}
		}
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
		return nil
	}

	_, err := a.config.GetProject(name)
	if err != nil {
		return err
	}

	a.config.SetLastActiveProject(name)

	a.emitStateUpdate()
	return nil
}

func (a *App) AddProject() (string, error) {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
	if err != nil || dir == "" {
		return "", err
	}

	// Use last two path segments as project name (e.g. "Go/dev-processes")
	dir = filepath.Clean(dir)
	name := filepath.Base(dir)
	parent := filepath.Base(filepath.Dir(dir))
	if parent != "" && parent != "." && parent != "/" && parent != "\\" {
		name = parent + "/" + name
	}

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
	return err
}

func (a *App) StartCommand(name string) error {
	projName := a.config.Config.LastActiveProject
	if projName == "" {
		return nil
	}
	proj, err := a.config.GetProject(projName)
	if err != nil {
		return err
	}

	for _, cmd := range proj.Commands {
		if cmd.Name == name {
			a.process.Kill(name)
			a.process.Remove(name)
			err = a.process.Start(name, cmd.Cmd, proj.Path)
			a.emitStateUpdate()
			return err
		}
	}
	return fmt.Errorf("command %q not found", name)
}

func (a *App) KillCommand(name string) error {
	err := a.process.Kill(name)
	a.emitStateUpdate()
	return err
}

func (a *App) StartAll() {
	projName := a.config.Config.LastActiveProject
	if projName == "" {
		return
	}
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
	a.emitStateUpdate()
}

func (a *App) KillAll() {
	a.process.KillAll()
	a.emitStateUpdate()
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
	if a.OnTrayUpdate != nil {
		a.OnTrayUpdate()
	}
}
