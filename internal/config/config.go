package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultPort        = 18890
	defaultGatewayUnit = "openclaw-gateway.service"
	defaultLogDir      = "/tmp/openclaw"
)

type Config struct {
	Port                int    `json:"port"`
	BindAddress         string `json:"bindAddress"`
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
	defaultOpenClawRoot, defaultWorkspace, err := defaultPaths()
	if err != nil {
		return Config{}, err
	}
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Port:          defaultPort,
		BindAddress:   "127.0.0.1",
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

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open .env: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set %s from .env: %w", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read .env: %w", err)
	}
	return nil
}

func defaultPaths() (openClawRoot string, workspaceRoot string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("resolve user home: %w", err)
	}
	openClawRoot = filepath.Join(home, ".openclaw")
	workspaceRoot = filepath.Join(openClawRoot, "workspace")
	return openClawRoot, workspaceRoot, nil
}

func (c Config) ListenAddr() string {
	return net.JoinHostPort(c.BindAddress, fmt.Sprintf("%d", c.Port))
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
	if c.BindAddress == "" {
		c.BindAddress = "127.0.0.1"
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
	if v := os.Getenv("OPENCLAUDIO_BIND_ADDRESS"); v != "" {
		cfg.BindAddress = v
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
