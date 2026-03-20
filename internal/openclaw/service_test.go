package openclaw

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fjrevoredo/openclaudio/internal/config"
)

func TestSessionsAndCron(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll := func(rel string) {
		if err := os.MkdirAll(filepath.Join(root, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite := func(rel, body string) {
		if err := os.WriteFile(filepath.Join(root, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	mustMkdirAll("agents/main/sessions")
	mustMkdirAll("cron/runs")
	mustWrite("agents/main/sessions/sessions.json", `{"a":{"updatedAt":9999999999999,"deliveryContext":{"channel":"telegram"},"origin":{"label":"Test"}}}`)
	mustWrite("cron/jobs.json", `{"jobs":[{"id":"job1","name":"Check","enabled":true,"schedule":{"expr":"0 * * * *","tz":"UTC"},"state":{"lastStatus":"ok","lastDurationMs":1000,"lastRunAtMs":9999999999999,"nextRunAtMs":9999999999999,"lastDeliveryStatus":"delivered","consecutiveErrors":0}}]}`)
	mustWrite("cron/runs/job1.jsonl", `{"ts":9999999999999,"jobId":"job1","status":"ok"}`)
	mustWrite("openclaw.json", `{"agents":{"defaults":{"model":{"primary":"zai/glm-5","fallbacks":["zai/glm-4.7-flash"]}}},"gateway":{"port":18789,"bind":"auto"}}`)

	svc := New(config.Config{
		OpenClawRoot: root,
		LogDir:       t.TempDir(),
		GatewayUnit:  "openclaw-gateway.service",
	})

	sessions, err := svc.Sessions()
	if err != nil {
		t.Fatal(err)
	}
	if sessions.ActiveCount != 1 {
		t.Fatalf("ActiveCount = %d, want 1", sessions.ActiveCount)
	}

	cron, err := svc.Cron()
	if err != nil {
		t.Fatal(err)
	}
	if len(cron.Jobs) != 1 || cron.Jobs[0].SuccessCount24h != 1 {
		t.Fatalf("Cron summary = %+v, want one successful job", cron.Jobs)
	}
}
