# DevProc — Development Process Manager

A system tray application for managing development commands (npm run dev, wails dev, air, etc.) across projects. Built with Wails v2 + Svelte 5 + Go.

## Problem

When developing web apps (Laravel, Go, etc.), you often need multiple long-running commands in separate terminals (Vite, queue workers, file watchers). Managing these manually means:
- Forgetting to start them
- Port conflicts when they don't get killed properly
- No quick way to see if everything is running
- Tedious context switching between projects

## Solution

A lightweight systray app that manages development commands per-project. Select a project, configure commands once, and monitor/control them from a single popup window.

## Tech Stack

- **Backend:** Go (process management, config, systray)
- **Frontend:** Svelte 5 (borderless popup window)
- **Framework:** Wails v2 (native webview, systray, cross-platform)
- **Config:** YAML in `%APPDATA%/devproc/config.yaml`

## Architecture

### Components

1. **ProcessManager** — spawns, monitors, and kills child processes. Each command runs via `cmd /C <command>` on Windows (platform-appropriate shell on others). Processes are killed with full tree kill (`taskkill /T /F /PID` on Windows) to avoid orphaned sub-processes and port conflicts.

2. **ConfigManager** — reads/writes YAML config. Manages projects, commands per project, and global command history.

3. **SystrayManager** — controls the system tray icon and its color state. Handles left-click (toggle popup) and right-click (context menu with Quit).

4. **Svelte UI** — borderless popup window (~320px wide) with project selector, command list, and add-command dialog.

### Data Flow

```
User clicks systray icon
  → Wails toggles popup window
    → Svelte reads state from Go backend via Wails bindings
      → User interacts (start/stop/switch project)
        → Svelte calls Go backend methods
          → ProcessManager spawns/kills processes
          → ConfigManager persists changes
          → SystrayManager updates icon color
```

## Config File

Location: `%APPDATA%/devproc/config.yaml` (Windows), `~/.config/devproc/config.yaml` (Linux/macOS)

```yaml
projects:
  - name: "my-laravel-app"
    path: "~/Projects/my-laravel-app"
    commands:
      - name: "Vite"
        cmd: "npm run dev"
      - name: "Queue Worker"
        cmd: "php artisan queue:work"

  - name: "go-webapp"
    path: "~/Projects/go-webapp"
    commands:
      - name: "Air"
        cmd: "air"

command_history:
  - "npm run dev"
  - "wails dev"
  - "air"
  - "php artisan queue:work"

last_active_project: ""
```

- `command_history` is global, populated whenever a new command is added to any project. Powers the "pick from existing" feature in the add-command dialog.
- `last_active_project` is saved for convenience but NOT auto-started on launch. App always starts with no project selected.

## Systray Icon States

| State | Icon Color | Condition |
|-------|-----------|-----------|
| Idle | Grey | No project selected |
| All running | Green | All commands for active project are running |
| Partial | Yellow | Some commands running, some stopped/errored |
| All down | Red | All commands stopped or errored |

- **Left-click:** Toggle popup window visibility
- **Right-click:** Context menu with "Quit" option

## Main Window UI

Borderless popup, dark brutalist theme (~320px wide). Compact monospace style.

### Layout (top to bottom)

1. **Title bar** — "DEVPROC" label + close button (hides window, doesn't quit app)
2. **Project selector** — dropdown showing current project name + path underneath. Includes an "Add Project" option that opens a native directory picker (via Wails dialog). The selected folder's name becomes the default project name (editable before confirming).
3. **Command list** — each command shows:
   - Status indicator (colored dot: green=running, grey=stopped, red=errored)
   - Command name (bold) + actual command string below
   - Action button: KILL (running), START (stopped), RESTART (errored)
   - Right-click or hover-X to delete a command from the project
4. **Bottom bar** — "+ ADD CMD" button, "START ALL" button, "KILL ALL" button

### Visual Style

- Dark background (#0a0a0a)
- Monospace font (Courier New / system mono)
- Uppercase labels with letter-spacing
- Status dots with glow effect for running/errored
- Minimal borders (#333)
- Color-coded action buttons (green=start, red=kill, yellow=restart)

## Add Command Dialog

Small secondary borderless window, same dark style. Contains:

1. **Name field** — display label for the command (e.g., "Vite")
2. **Command field** — text input for the actual command (e.g., "npm run dev")
3. **Command history list** — clickable list of previously used commands from `command_history`. Clicking one populates the command field.
4. **Save button** — adds command to current project and to global `command_history`
5. **Cancel button** — closes dialog

## Process Management

### Spawning
- Commands run via `cmd /C <command>` on Windows, `sh -c <command>` on Unix
- Working directory set to the project's `path`
- Each command gets its own goroutine monitoring the process

### Killing
- Windows: `taskkill /T /F /PID <pid>` for full process tree kill
- Unix: kill the process group (`-pid`) with SIGTERM, then SIGKILL after timeout
- Tree kill is essential — commands like `npm run dev` spawn child processes that hold ports

### Status Monitoring
- Goroutine per command watches process exit
- On exit: check exit code — 0 = stopped, non-zero = errored
- Status changes are pushed to the frontend via Wails events
- Systray icon color updates whenever any command status changes

### Project Switching
- All running processes for the current project are killed (tree kill)
- New project's commands are automatically started
- If switching to "no project", just kill everything

### App Exit
- On quit (from systray context menu): kill all running processes before exiting
- Ensure no orphaned processes remain

## Error Handling

- If a command fails to start (bad command, missing binary): mark as errored immediately, show error in status text
- If a command exits unexpectedly: mark as errored, show exit code
- Config file read errors on startup: create default empty config
- Config file write errors: log to stderr, don't crash

## Scope Exclusions

- No live log/output viewer — user runs commands manually for debugging
- No auto-restart on crash — manual restart only
- No templates/presets — command history provides the convenience
- No multi-project simultaneous operation — one active project at a time
- No remote/SSH process management
- No scheduling or timers
