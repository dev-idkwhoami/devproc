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
