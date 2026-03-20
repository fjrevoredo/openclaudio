package config

import "testing"

func TestLoadUsesEnvOverrides(t *testing.T) {
	t.Setenv("OPENCLAUDIO_SESSION_SECRET", "secret")
	t.Setenv("OPENCLAUDIO_ADMIN_USER", "admin")
	t.Setenv("OPENCLAUDIO_ADMIN_PASSWORD_HASH", "hash")
	t.Setenv("OPENCLAUDIO_PORT", "19001")
	t.Setenv("OPENCLAUDIO_WORKSPACE_ROOT", "/tmp/workspace")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != 19001 {
		t.Fatalf("Port = %d, want 19001", cfg.Port)
	}
	if cfg.WorkspaceRoot != "/tmp/workspace" {
		t.Fatalf("WorkspaceRoot = %q", cfg.WorkspaceRoot)
	}
}
