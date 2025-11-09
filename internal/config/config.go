package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/chrmzio/dotty/internal/utils"
)

type Config struct {
	RepoURL  string   `toml:"repo_url"`
	Dotfiles []string `toml:"dotfiles"`
}

type Manager struct {
	configPath string
	config     *Config
}

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "dotty", "config.toml")

	manager := &Manager{
		configPath: configPath,
	}

	if err := manager.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	return manager, nil
}

func NewManagerWithPath(configPath string) (*Manager, error) {
	manager := &Manager{
		configPath: configPath,
	}

	if err := manager.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	return manager, nil
}

func (m *Manager) Initialize() error {
	if err := m.ensureConfigExists(); err != nil {
		return fmt.Errorf("failed to ensure config exists: %w", err)
	}

	if err := m.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return nil
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := m.validateConfig(&config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.config = &config
	return nil
}

func (m *Manager) validateConfig(cfg *Config) error {
	for i, dotfile := range cfg.Dotfiles {
		if dotfile == "" {
			return fmt.Errorf("empty dotfile path at index %d", i)
		}

		expanded, err := utils.ExpandPath(dotfile)
		if err != nil {
			return fmt.Errorf("invalid dotfile path %q: %w", dotfile, err)
		}

		// We don't check if file exists here, as it might not exist yet
		// That check should be done in status command
		_ = expanded
	}

	return nil
}

func (m *Manager) AddDotfile(path string) error {
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Normalize the path for storage (keep ~ for readability in config)
	// But first expand it to check if it's valid
	expanded, err := utils.ExpandPath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Convert back to normalized form for storage
	normalized := utils.NormalizePath(expanded)

	for _, existing := range m.config.Dotfiles {
		existingExpanded, _ := utils.ExpandPath(existing)
		if existingExpanded == expanded {
			return fmt.Errorf("dotfile already exists: %s", path)
		}
	}

	m.config.Dotfiles = append(m.config.Dotfiles, normalized)
	return m.Save()
}

// RemoveDotfile removes a dotfile from the configuration
func (m *Manager) RemoveDotfile(path string) error {
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	expanded, err := utils.ExpandPath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	newDotfiles := make([]string, 0, len(m.config.Dotfiles))
	found := false

	for _, dotfile := range m.config.Dotfiles {
		dotfileExpanded, _ := utils.ExpandPath(dotfile)
		if dotfileExpanded != expanded {
			newDotfiles = append(newDotfiles, dotfile)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("dotfile not found: %s", path)
	}

	m.config.Dotfiles = newDotfiles
	return m.Save()
}

func (m *Manager) GetExpandedDotfiles() (map[string]string, error) {
	if m.config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	result := make(map[string]string)
	for _, dotfile := range m.config.Dotfiles {
		expanded, err := utils.ExpandPath(dotfile)
		if err != nil {
			return nil, fmt.Errorf("failed to expand %s: %w", dotfile, err)
		}
		result[dotfile] = expanded
	}

	return result, nil
}

func (m *Manager) ensureConfigExists() error {
	info, err := os.Stat(m.configPath)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("config path is a directory: %s", m.configPath)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	defaultConfig := &Config{
		RepoURL:  "",
		Dotfiles: []string{},
	}

	if err := m.saveConfig(defaultConfig); err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}

	return nil
}

func (m *Manager) Save() error {
	if m.config == nil {
		return fmt.Errorf("no configuration to save")
	}
	return m.saveConfig(m.config)
}

func (m *Manager) saveConfig(cfg *Config) error {
	file, err := os.Create(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

func (m *Manager) Get() *Config {
	if m.config == nil {
		return &Config{}
	}
	return &Config{
		RepoURL:  m.config.RepoURL,
		Dotfiles: append([]string{}, m.config.Dotfiles...),
	}
}

func (m *Manager) GetPath() string {
	return m.configPath
}

func (m *Manager) SetRepoURL(url string) error {
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	m.config.RepoURL = url
	return m.Save()
}
