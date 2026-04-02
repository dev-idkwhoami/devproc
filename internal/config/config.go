package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Command struct {
	Name string `yaml:"name" json:"name"`
	Cmd  string `yaml:"cmd" json:"cmd"`
}

type Project struct {
	Name     string    `yaml:"name" json:"name"`
	Path     string    `yaml:"path" json:"path"`
	Commands []Command `yaml:"commands" json:"commands"`
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
