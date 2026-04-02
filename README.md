# DevProc

<img src="icons/grey.png" width="64" alt="DevProc logo">

A lightweight system tray app for managing development commands across projects. Stop juggling terminal windows for `npm run dev`, `wails dev`, `air`, and other long-running dev commands.

## Features

- **System tray status** at a glance -- know if your dev commands are running
- **Per-project command configs** -- switch between projects instantly
- **One-click kill** -- no more orphaned processes or port conflicts (full process tree kill)
- **Command history** -- reuse commands across projects without retyping
- **Borderless dark UI** -- compact, stays out of your way

## Tray Icon States

| Icon | Meaning |
|------|---------|
| Grey | No project selected |
| Green | All commands running |
| Yellow | Some commands not running |
| Red | All commands stopped |

## Installation

### From Releases

Download the latest `devproc-installer.exe` from [Releases](../../releases) and run it.

### Build from Source

Requires [Go 1.21+](https://go.dev/dl/) and [Wails v2](https://wails.io/docs/gettingstarted/installation).

```bash
git clone git@github:dev-idkwhoami/devproc.git
cd dev-processes
wails build
```

The binary is produced at `build/bin/devproc.exe`.

## Usage

1. **Launch** -- DevProc starts minimized in the system tray
2. **Click the tray icon** to open the popup window
3. **Add a project** -- click the project dropdown, then "+ Add project..." and select a folder
4. **Add commands** -- click "+ ADD CMD" and enter a name + command (e.g. "Vite" / `npm run dev`)
5. **Start/Stop** -- use the per-command buttons or "START ALL" / "KILL ALL"
6. **Switch projects** -- select a different project from the dropdown; running commands are killed automatically

### Verbose Logging

To log command stdout/stderr to `%APPDATA%/devproc/logs/`:

```bash
devproc.exe -verbose
```

## Configuration

Config is stored at:
- **Windows:** `%APPDATA%\devproc\config.yaml`
- **Linux/macOS:** `~/.config/devproc/config.yaml`

```yaml
projects:
  - name: "my-project"
    path: "~/Projects/my-project"
    commands:
      - name: "Vite"
        cmd: "npm run dev"
      - name: "Queue"
        cmd: "php artisan queue:work"

command_history:
  - "npm run dev"
  - "php artisan queue:work"

last_active_project: "my-project"
```

## Tech Stack

- **Backend:** Go + [Wails v2](https://wails.io)
- **Frontend:** Svelte 5 + TypeScript
- **System Tray:** [energye/systray](https://github.com/energye/systray)

## License

MIT
