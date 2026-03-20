package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const (
	defaultPort         = 18890
	defaultWorkspace    = "/home/francisco/.openclaw/workspace"
	defaultOpenClawRoot = "/home/francisco/.openclaw"
	defaultGatewayUnit  = "openclaw-gateway.service"
	defaultLogDir       = "/tmp/openclaw"
)

type Config struct {
	Port                int    `json:"port"`
	WorkspaceRoot       string `json:"workspaceRoot"`
	OpenClawRoot        string `json:"openClawRoot"`
	GatewayUnit         string `json:"gatewayUnit"`
	LogDir              string `json:"logDir"`
	SessionSecret       string `json:"sessionSecret"`
	AdminUser           string `json:"adminUser"`
	AdminPasswordHash   string `json:"adminPasswordHash"`
	OpenClawPackageJSON string `json:"openClawPackageJSON"`
	ConfigPath          string `json:"-"`
}

func Load() (Config, error) {
	cfg := Config{
		Port:          defaultPort,
		WorkspaceRoot: defaultWorkspace,
		OpenClawRoot:  defaultOpenClawRoot,
		GatewayUnit:   defaultGatewayUnit,
		LogDir:        defaultLogDir,
	}

	cfg.ConfigPath = configPath()
	if err := loadFileConfig(&cfg, cfg.ConfigPath); err != nil {
		return Config{}, err
	}

	applyEnv(&cfg)
	if err := cfg.normalize(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) ListenAddr() string {
	return fmt.Sprintf("127.0.0.1:%d", c.Port)
}

func (c Config) HomeDir() string {
	return filepath.Dir(c.OpenClawRoot)
}

func (c *Config) normalize() error {
	var err error

	c.WorkspaceRoot, err = filepath.Abs(c.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("workspaceRoot: %w", err)
	}

	c.OpenClawRoot, err = filepath.Abs(c.OpenClawRoot)
	if err != nil {
		return fmt.Errorf("openClawRoot: %w", err)
	}

	if c.LogDir == "" {
		c.LogDir = defaultLogDir
	}

	if c.Port < 1 || c.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if c.SessionSecret == "" {
		return errors.New("OPENCLAUDIO_SESSION_SECRET is required")
	}
	if c.AdminUser == "" {
		return errors.New("OPENCLAUDIO_ADMIN_USER is required")
	}
	if c.AdminPasswordHash == "" {
		return errors.New("OPENCLAUDIO_ADMIN_PASSWORD_HASH is required")
	}
	return nil
}

func configPath() string {
	if path := os.Getenv("OPENCLAUDIO_CONFIG"); path != "" {
		return path
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "openclaudio", "config.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "openclaudio", "config.json")
}

func loadFileConfig(cfg *Config, path string) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}
	return nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("OPENCLAUDIO_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Port = port
		}
	}
	if v := os.Getenv("OPENCLAUDIO_WORKSPACE_ROOT"); v != "" {
		cfg.WorkspaceRoot = v
	}
	if v := os.Getenv("OPENCLAUDIO_OPENCLAW_ROOT"); v != "" {
		cfg.OpenClawRoot = v
	}
	if v := os.Getenv("OPENCLAUDIO_GATEWAY_UNIT"); v != "" {
		cfg.GatewayUnit = v
	}
	if v := os.Getenv("OPENCLAUDIO_LOG_DIR"); v != "" {
		cfg.LogDir = v
	}
	if v := os.Getenv("OPENCLAUDIO_SESSION_SECRET"); v != "" {
		cfg.SessionSecret = v
	}
	if v := os.Getenv("OPENCLAUDIO_ADMIN_USER"); v != "" {
		cfg.AdminUser = v
	}
	if v := os.Getenv("OPENCLAUDIO_ADMIN_PASSWORD_HASH"); v != "" {
		cfg.AdminPasswordHash = v
	}
	if v := os.Getenv("OPENCLAUDIO_OPENCLAW_PACKAGE_JSON"); v != "" {
		cfg.OpenClawPackageJSON = v
	}
}
