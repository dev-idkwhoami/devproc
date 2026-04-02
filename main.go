package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
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

func getAppDataDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "devproc")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "devproc")
}

func main() {
	verbose := flag.Bool("verbose", false, "Enable command output logging to %APPDATA%/devproc/logs/")
	flag.Parse()

	appDataDir := getAppDataDir()
	logDir := ""
	if *verbose {
		logDir = filepath.Join(appDataDir, "logs")
	}
	app := NewApp(
		filepath.Join(appDataDir, "config.yaml"),
		logDir,
	)

	var trayMgr *tray.Manager

	// Register systray (non-blocking) — Wails needs the main thread on Windows
	systray.Register(func() {
		trayMgr = tray.NewManager(tray.Callbacks{
			OnLeftClick: func() {
				if app.ctx != nil {
					wailsRuntime.Show(app.ctx)
				}
			},
			OnQuit: func() {
				app.Shutdown()
				if app.ctx != nil {
					wailsRuntime.Quit(app.ctx)
				}
			},
		})
		trayMgr.Setup()

		// Wire tray updates to every state change
		app.OnTrayUpdate = func() {
			updateTrayState(trayMgr, app)
		}

		// Process status changes update frontend (which triggers tray update too)
		app.process.OnChange = func(name string, status process.Status, exitCode int) {
			app.emitStateUpdate()
		}

		// Set initial tray state
		updateTrayState(trayMgr, app)
	}, func() {
		// Cleanup on systray exit
	})

	// Wails runs on the main thread (required on Windows)
	err := wails.Run(&options.App{
		Title:             "DevProc",
		Width:             320,
		Height:            480,
		Frameless:         true,
		StartHidden:       true,
		HideWindowOnClose: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		OnShutdown: func(ctx context.Context) {
			app.Shutdown()
			systray.Quit()
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
}

func updateTrayState(trayMgr *tray.Manager, app *App) {
	projName := app.config.Config.LastActiveProject
	if projName == "" {
		trayMgr.SetState(tray.StateIdle, "DevProc — No project selected")
		return
	}

	proj, err := app.config.GetProject(projName)
	if err != nil || len(proj.Commands) == 0 {
		trayMgr.SetState(tray.StateIdle, "DevProc — "+projName+" (no commands)")
		return
	}

	// Count running commands against total configured commands
	total := len(proj.Commands)
	running := 0
	for _, cmd := range proj.Commands {
		cs := app.process.GetState(cmd.Name)
		if cs != nil && cs.Status == process.Running {
			running++
		}
	}

	switch {
	case running == total:
		trayMgr.SetState(tray.StateAllUp, fmt.Sprintf("DevProc — %s (all running)", projName))
	case running == 0:
		trayMgr.SetState(tray.StateAllDown, fmt.Sprintf("DevProc — %s (all stopped)", projName))
	default:
		trayMgr.SetState(tray.StatePartial, fmt.Sprintf("DevProc — %s (%d/%d running)", projName, running, total))
	}
}
