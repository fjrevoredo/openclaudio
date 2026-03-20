package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fjrevoredo/openclaudio/internal/config"
	"github.com/fjrevoredo/openclaudio/internal/openclaw"
)

func TestGatewayActionTemplateShowsFailureDetails(t *testing.T) {
	server := testServer(t)

	var b strings.Builder
	err := server.templates.ExecuteTemplate(&b, "gateway-action", gatewayActionData{
		Result: openclaw.GatewayActionResult{
			Action:    "restart",
			Success:   false,
			Output:    "permission denied",
			Timestamp: time.Unix(0, 0),
		},
		Error: "systemctl failed",
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	got := b.String()
	if !strings.Contains(got, "restart failed") {
		t.Fatalf("output missing failure heading: %s", got)
	}
	if !strings.Contains(got, "systemctl failed") {
		t.Fatalf("output missing error detail: %s", got)
	}
	if !strings.Contains(got, "permission denied") {
		t.Fatalf("output missing command output: %s", got)
	}
}

func testServer(t *testing.T) *Server {
	t.Helper()

	root := t.TempDir()
	workspace := filepath.Join(root, "workspace")
	openclawRoot := filepath.Join(root, ".openclaw")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(openclawRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	server, err := New(config.Config{
		Port:              18890,
		WorkspaceRoot:     workspace,
		OpenClawRoot:      openclawRoot,
		GatewayUnit:       "openclaw-gateway.service",
		LogDir:            t.TempDir(),
		SessionSecret:     "secret",
		AdminUser:         "admin",
		AdminPasswordHash: "$2a$10$i5yBxxe9Bz3SANaHmUtlXe.hVuoEeY7fU7TlqblzgZiMEiefINlIa",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return server
}
