package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsDotEnv(t *testing.T) {
	dir := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldwd)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	dotenv := "OPENCLAUDIO_SESSION_SECRET=secret\nOPENCLAUDIO_ADMIN_USER=admin\nOPENCLAUDIO_ADMIN_PASSWORD_HASH=$2a$10$example\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(dotenv), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SessionSecret != "secret" {
		t.Fatalf("SessionSecret = %q, want secret", cfg.SessionSecret)
	}
	if cfg.AdminPasswordHash != "$2a$10$example" {
		t.Fatalf("AdminPasswordHash = %q, want bcrypt-like value", cfg.AdminPasswordHash)
	}
}

func TestLoadUsesEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldwd)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENCLAUDIO_SESSION_SECRET", "secret")
	t.Setenv("OPENCLAUDIO_ADMIN_USER", "admin")
	t.Setenv("OPENCLAUDIO_ADMIN_PASSWORD_HASH", "hash")
	t.Setenv("OPENCLAUDIO_PORT", "19001")
	t.Setenv("OPENCLAUDIO_BIND_ADDRESS", "0.0.0.0")
	t.Setenv("OPENCLAUDIO_WORKSPACE_ROOT", "/tmp/workspace")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != 19001 {
		t.Fatalf("Port = %d, want 19001", cfg.Port)
	}
	if cfg.BindAddress != "0.0.0.0" {
		t.Fatalf("BindAddress = %q, want 0.0.0.0", cfg.BindAddress)
	}
	if cfg.WorkspaceRoot != "/tmp/workspace" {
		t.Fatalf("WorkspaceRoot = %q", cfg.WorkspaceRoot)
	}
	if cfg.ListenAddr() != "0.0.0.0:19001" {
		t.Fatalf("ListenAddr() = %q, want 0.0.0.0:19001", cfg.ListenAddr())
	}
}
