package openclaw

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fjrevoredo/openclaudio/internal/config"
)

type Service struct {
	cfg config.Config
}

type Summary struct {
	Version            string
	PrimaryModel       string
	FallbackModels     []string
	GatewayPort        int
	GatewayBind        string
	ActiveSessionCount int
	Service            ServiceState
	Process            ProcessMetrics
	LogTail            []string
}

type SessionSummary struct {
	ActiveCount int
	Recent      []SessionInfo
}

type SessionInfo struct {
	Key       string
	Label     string
	Channel   string
	UpdatedAt time.Time
}

type CronSummary struct {
	Jobs []CronJobSummary
}

type CronJobSummary struct {
	ID               string
	Name             string
	Enabled          bool
	Schedule         string
	TimeZone         string
	LastStatus       string
	LastDurationMS   int64
	LastRunAt        time.Time
	NextRunAt        time.Time
	SuccessCount24h  int
	ErrorCount24h    int
	LastDelivery     string
	ConsecutiveError int
}

type ServiceState struct {
	Available   bool
	ActiveState string
	SubState    string
	MainPID     int
	Error       string
	CheckedAt   time.Time
}

type ProcessMetrics struct {
	Available  bool
	CPUPercent string
	RSSKB      int64
	ElapsedSec int64
	Command    string
	Error      string
}

type GatewayActionResult struct {
	Action    string
	Success   bool
	Output    string
	Timestamp time.Time
}

func New(cfg config.Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) Summary(ctx context.Context) (Summary, error) {
	oc, _ := s.readOpenClawConfig()
	sessions, _ := s.Sessions()
	service := s.ServiceState(ctx)
	process := s.ProcessMetrics(ctx, service.MainPID)
	logTail, _ := s.LogTail(time.Now(), 12)

	return Summary{
		Version:            s.Version(),
		PrimaryModel:       oc.PrimaryModel,
		FallbackModels:     oc.FallbackModels,
		GatewayPort:        oc.GatewayPort,
		GatewayBind:        oc.GatewayBind,
		ActiveSessionCount: sessions.ActiveCount,
		Service:            service,
		Process:            process,
		LogTail:            logTail,
	}, nil
}

func (s *Service) Version() string {
	paths := s.packageCandidates()
	for _, candidate := range paths {
		data, err := os.ReadFile(candidate)
		if err != nil {
			continue
		}
		var pkg struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(data, &pkg) == nil && pkg.Version != "" {
			return pkg.Version
		}
	}
	return "unavailable"
}

func (s *Service) Sessions() (SessionSummary, error) {
	path := filepath.Join(s.cfg.OpenClawRoot, "agents", "main", "sessions", "sessions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return SessionSummary{}, err
	}

	var raw map[string]struct {
		UpdatedAt       int64 `json:"updatedAt"`
		DeliveryContext struct {
			Channel string `json:"channel"`
		} `json:"deliveryContext"`
		Origin struct {
			Label string `json:"label"`
		} `json:"origin"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return SessionSummary{}, err
	}

	cutoff := time.Now().Add(-24 * time.Hour)
	var recent []SessionInfo
	active := 0

	for key, item := range raw {
		updated := time.UnixMilli(item.UpdatedAt)
		if updated.After(cutoff) {
			active++
		}
		recent = append(recent, SessionInfo{
			Key:       key,
			Label:     item.Origin.Label,
			Channel:   item.DeliveryContext.Channel,
			UpdatedAt: updated,
		})
	}

	sort.Slice(recent, func(i, j int) bool {
		return recent[i].UpdatedAt.After(recent[j].UpdatedAt)
	})
	if len(recent) > 10 {
		recent = recent[:10]
	}

	return SessionSummary{ActiveCount: active, Recent: recent}, nil
}

func (s *Service) Cron() (CronSummary, error) {
	path := filepath.Join(s.cfg.OpenClawRoot, "cron", "jobs.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return CronSummary{}, err
	}

	var jobs struct {
		Jobs []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Enabled  bool   `json:"enabled"`
			Schedule struct {
				Expr string `json:"expr"`
				TZ   string `json:"tz"`
			} `json:"schedule"`
			State struct {
				LastStatus        string `json:"lastStatus"`
				LastDurationMS    int64  `json:"lastDurationMs"`
				LastRunAtMS       int64  `json:"lastRunAtMs"`
				NextRunAtMS       int64  `json:"nextRunAtMs"`
				LastDelivery      string `json:"lastDeliveryStatus"`
				ConsecutiveErrors int    `json:"consecutiveErrors"`
			} `json:"state"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(data, &jobs); err != nil {
		return CronSummary{}, err
	}

	type agg struct{ ok, err int }
	counts := map[string]agg{}
	cutoff := time.Now().Add(-24 * time.Hour)
	pattern := filepath.Join(s.cfg.OpenClawRoot, "cron", "runs", "*.jsonl")
	runFiles, _ := filepath.Glob(pattern)
	for _, runFile := range runFiles {
		file, err := os.Open(runFile)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var item struct {
				TS     int64  `json:"ts"`
				JobID  string `json:"jobId"`
				Status string `json:"status"`
			}
			if json.Unmarshal(scanner.Bytes(), &item) != nil {
				continue
			}
			if time.UnixMilli(item.TS).Before(cutoff) {
				continue
			}
			cur := counts[item.JobID]
			if item.Status == "ok" {
				cur.ok++
			} else {
				cur.err++
			}
			counts[item.JobID] = cur
		}
		_ = file.Close()
	}

	summary := CronSummary{Jobs: make([]CronJobSummary, 0, len(jobs.Jobs))}
	for _, job := range jobs.Jobs {
		count := counts[job.ID]
		summary.Jobs = append(summary.Jobs, CronJobSummary{
			ID:               job.ID,
			Name:             job.Name,
			Enabled:          job.Enabled,
			Schedule:         job.Schedule.Expr,
			TimeZone:         job.Schedule.TZ,
			LastStatus:       job.State.LastStatus,
			LastDurationMS:   job.State.LastDurationMS,
			LastRunAt:        time.UnixMilli(job.State.LastRunAtMS),
			NextRunAt:        time.UnixMilli(job.State.NextRunAtMS),
			SuccessCount24h:  count.ok,
			ErrorCount24h:    count.err,
			LastDelivery:     job.State.LastDelivery,
			ConsecutiveError: job.State.ConsecutiveErrors,
		})
	}
	sort.Slice(summary.Jobs, func(i, j int) bool {
		return summary.Jobs[i].Name < summary.Jobs[j].Name
	})

	return summary, nil
}

func (s *Service) LogTail(date time.Time, lines int) ([]string, error) {
	if lines <= 0 {
		lines = 200
	}
	path := filepath.Join(s.cfg.LogDir, "openclaw-"+date.Format("2006-01-02")+".log")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer := make([]string, 0, lines)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(buffer) == lines {
			copy(buffer, buffer[1:])
			buffer[len(buffer)-1] = scanner.Text()
			continue
		}
		buffer = append(buffer, scanner.Text())
	}
	return buffer, scanner.Err()
}

func (s *Service) ServiceState(ctx context.Context) ServiceState {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", "show", s.cfg.GatewayUnit,
		"--property=ActiveState,SubState,MainPID", "--no-pager")
	output, err := cmd.CombinedOutput()
	state := ServiceState{CheckedAt: time.Now()}
	if err != nil {
		state.Error = strings.TrimSpace(string(output))
		return state
	}

	state.Available = true
	for _, line := range strings.Split(string(output), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "ActiveState":
			state.ActiveState = value
		case "SubState":
			state.SubState = value
		case "MainPID":
			if pid, convErr := strconv.Atoi(value); convErr == nil {
				state.MainPID = pid
			}
		}
	}
	return state
}

func (s *Service) ProcessMetrics(ctx context.Context, pid int) ProcessMetrics {
	if pid <= 0 {
		return ProcessMetrics{}
	}
	cmd := exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(pid), "-o", "%cpu=,rss=,etimes=,args=")
	output, err := cmd.Output()
	if err != nil {
		return ProcessMetrics{Error: strings.TrimSpace(string(output))}
	}
	fields := strings.Fields(string(output))
	if len(fields) < 4 {
		return ProcessMetrics{Error: "unexpected ps output"}
	}
	rss, _ := strconv.ParseInt(fields[1], 10, 64)
	elapsed, _ := strconv.ParseInt(fields[2], 10, 64)
	return ProcessMetrics{
		Available:  true,
		CPUPercent: fields[0],
		RSSKB:      rss,
		ElapsedSec: elapsed,
		Command:    strings.Join(fields[3:], " "),
	}
}

func (s *Service) GatewayAction(ctx context.Context, action string) (GatewayActionResult, error) {
	switch action {
	case "start", "stop", "restart":
	default:
		return GatewayActionResult{}, errors.New("unsupported gateway action")
	}
	cmd := exec.CommandContext(ctx, "systemctl", "--user", action, s.cfg.GatewayUnit)
	output, err := cmd.CombinedOutput()
	return GatewayActionResult{
		Action:    action,
		Success:   err == nil,
		Output:    strings.TrimSpace(string(output)),
		Timestamp: time.Now(),
	}, err
}

type openClawConfig struct {
	PrimaryModel   string
	FallbackModels []string
	GatewayPort    int
	GatewayBind    string
}

func (s *Service) readOpenClawConfig() (openClawConfig, error) {
	path := filepath.Join(s.cfg.OpenClawRoot, "openclaw.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return openClawConfig{}, err
	}
	var raw struct {
		Agents struct {
			Defaults struct {
				Model struct {
					Primary   string   `json:"primary"`
					Fallbacks []string `json:"fallbacks"`
				} `json:"model"`
			} `json:"defaults"`
		} `json:"agents"`
		Gateway struct {
			Port int    `json:"port"`
			Bind string `json:"bind"`
		} `json:"gateway"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return openClawConfig{}, err
	}
	return openClawConfig{
		PrimaryModel:   raw.Agents.Defaults.Model.Primary,
		FallbackModels: raw.Agents.Defaults.Model.Fallbacks,
		GatewayPort:    raw.Gateway.Port,
		GatewayBind:    raw.Gateway.Bind,
	}, nil
}

func (s *Service) packageCandidates() []string {
	if s.cfg.OpenClawPackageJSON != "" {
		return []string{s.cfg.OpenClawPackageJSON}
	}

	home := s.cfg.HomeDir()
	patterns := []string{
		filepath.Join(home, ".local", "share", "pnpm", "global", "*", ".pnpm", "openclaw@*", "node_modules", "openclaw", "package.json"),
		filepath.Join(home, ".local", "lib", "node_modules", "openclaw", "package.json"),
	}

	var out []string
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		sort.Strings(matches)
		out = append(out, matches...)
	}
	return out
}
