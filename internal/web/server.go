package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fjrevoredo/openclaudio/internal/auth"
	"github.com/fjrevoredo/openclaudio/internal/config"
	"github.com/fjrevoredo/openclaudio/internal/files"
	"github.com/fjrevoredo/openclaudio/internal/markdown"
	"github.com/fjrevoredo/openclaudio/internal/openclaw"
	webassets "github.com/fjrevoredo/openclaudio/web"
)

type Server struct {
	cfg       config.Config
	auth      *auth.Manager
	files     *files.Service
	openclaw  *openclaw.Service
	templates *template.Template
}

type pageData struct {
	Title        string
	User         string
	CSRFToken    string
	SelectedPath string
}

type loginData struct {
	Title     string
	CSRFToken string
	Error     string
}

type treeData struct {
	Nodes       []files.TreeNode
	CurrentPath string
	Query       string
	Error       string
}

type fileData struct {
	Document     files.Document
	RenderedHTML template.HTML
	Error        string
}

type summaryData struct {
	Summary openclaw.Summary
	Error   string
}

type sessionsData struct {
	Sessions openclaw.SessionSummary
	Error    string
}

type cronData struct {
	Cron  openclaw.CronSummary
	Error string
}

type logsData struct {
	Date  string
	Lines []string
	Error string
}

type gatewayActionData struct {
	Result openclaw.GatewayActionResult
	Error  string
}

func New(cfg config.Config) (*Server, error) {
	renderer := markdown.New()
	fileSvc, err := files.New(cfg.WorkspaceRoot, renderer)
	if err != nil {
		return nil, err
	}

	tplFS, err := fs.Sub(webassets.FS, "templates")
	if err != nil {
		return nil, err
	}

	funcs := template.FuncMap{
		"fmtTime": func(t time.Time) string {
			if t.IsZero() {
				return "n/a"
			}
			return t.Format("2006-01-02 15:04")
		},
		"timeFromNs": func(v int64) time.Time {
			if v <= 0 {
				return time.Time{}
			}
			return time.Unix(0, v)
		},
		"humanKB": func(v int64) string {
			if v <= 0 {
				return "0 KB"
			}
			return fmt.Sprintf("%.1f MB", float64(v)/1024)
		},
		"humanBytes": func(v int64) string {
			switch {
			case v > 1<<30:
				return fmt.Sprintf("%.1f GB", float64(v)/(1<<30))
			case v > 1<<20:
				return fmt.Sprintf("%.1f MB", float64(v)/(1<<20))
			case v > 1<<10:
				return fmt.Sprintf("%.1f KB", float64(v)/(1<<10))
			default:
				return fmt.Sprintf("%d B", v)
			}
		},
		"trimLabel": func(v string) string {
			if v == "" {
				return "unknown"
			}
			if len(v) <= 48 {
				return v
			}
			return v[:48] + "..."
		},
		"fileURL": func(rel string) string {
			return "/files/" + rel
		},
	}

	templates, err := template.New("").Funcs(funcs).ParseFS(tplFS, "*.html")
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg:       cfg,
		auth:      auth.New(cfg.SessionSecret),
		files:     fileSvc,
		openclaw:  openclaw.New(cfg),
		templates: templates,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	staticFS, _ := fs.Sub(webassets.FS, "static/dist")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("GET /login", s.handleLoginPage)
	mux.HandleFunc("POST /login", s.handleLogin)
	mux.Handle("POST /logout", s.auth.Require(http.HandlerFunc(s.handleLogout)))
	mux.Handle("GET /", s.auth.Require(http.HandlerFunc(s.handleHome)))
	mux.Handle("GET /files/", s.auth.Require(http.HandlerFunc(s.handleFilePage)))
	mux.Handle("GET /api/tree", s.auth.Require(http.HandlerFunc(s.handleTree)))
	mux.Handle("GET /api/file", s.auth.Require(http.HandlerFunc(s.handleFile)))
	mux.Handle("PUT /api/file", s.auth.Require(http.HandlerFunc(s.handleSaveFile)))
	mux.Handle("POST /api/file/copy-path", s.auth.Require(http.HandlerFunc(s.handleCopyPath)))
	mux.Handle("GET /api/openclaw/summary", s.auth.Require(http.HandlerFunc(s.handleSummary)))
	mux.Handle("GET /api/openclaw/sessions", s.auth.Require(http.HandlerFunc(s.handleSessions)))
	mux.Handle("GET /api/openclaw/cron", s.auth.Require(http.HandlerFunc(s.handleCron)))
	mux.Handle("GET /api/openclaw/logs", s.auth.Require(http.HandlerFunc(s.handleLogs)))
	mux.Handle("POST /api/openclaw/gateway/start", s.auth.Require(http.HandlerFunc(s.handleGatewayAction("start"))))
	mux.Handle("POST /api/openclaw/gateway/stop", s.auth.Require(http.HandlerFunc(s.handleGatewayAction("stop"))))
	mux.Handle("POST /api/openclaw/gateway/restart", s.auth.Require(http.HandlerFunc(s.handleGatewayAction("restart"))))
	mux.ServeHTTP(w, r)
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if _, err := s.auth.CurrentUser(r); err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	token := s.auth.EnsureCSRFCookie(w, r)
	s.render(w, "login", loginData{
		Title:     "Login",
		CSRFToken: token,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	token := s.auth.EnsureCSRFCookie(w, r)
	if !s.auth.ValidateCSRF(r) {
		s.renderStatus(w, http.StatusForbidden, "login", loginData{
			Title:     "Login",
			CSRFToken: token,
			Error:     "invalid CSRF token",
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		s.renderStatus(w, http.StatusBadRequest, "login", loginData{
			Title:     "Login",
			CSRFToken: token,
			Error:     "invalid form",
		})
		return
	}

	if r.Form.Get("username") != s.cfg.AdminUser || auth.VerifyPassword(s.cfg.AdminPasswordHash, r.Form.Get("password")) != nil {
		s.renderStatus(w, http.StatusUnauthorized, "login", loginData{
			Title:     "Login",
			CSRFToken: token,
			Error:     "invalid username or password",
		})
		return
	}

	if err := s.auth.Login(w, r, s.cfg.AdminUser); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.auth.EnsureCSRFCookie(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if !s.auth.ValidateCSRF(r) {
		http.Error(w, "invalid CSRF token", http.StatusForbidden)
		return
	}
	s.auth.Logout(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	s.renderShell(w, r, "")
}

func (s *Server) handleFilePage(w http.ResponseWriter, r *http.Request) {
	selected := strings.TrimPrefix(r.URL.Path, "/files/")
	s.renderShell(w, r, selected)
}

func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.files.List(r.URL.Query().Get("path"), r.URL.Query().Get("q"))
	data := treeData{
		Nodes:       nodes,
		CurrentPath: r.URL.Query().Get("path"),
		Query:       r.URL.Query().Get("q"),
	}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, "tree", data)
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	doc, err := s.files.Read(r.URL.Query().Get("path"), r.URL.Query().Get("view"))
	data := fileData{Document: doc, RenderedHTML: template.HTML(doc.RenderedHTML)}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, "file", data)
}

func (s *Server) handleSaveFile(w http.ResponseWriter, r *http.Request) {
	if !s.auth.ValidateCSRF(r) {
		http.Error(w, "invalid CSRF token", http.StatusForbidden)
		return
	}
	var payload struct {
		Text           string `json:"text"`
		LastModifiedNS int64  `json:"lastModifiedNs"`
		ContentHash    string `json:"contentHash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := s.files.Save(files.SaveRequest{
		RelativePath:   r.URL.Query().Get("path"),
		Text:           payload.Text,
		LastModifiedNS: payload.LastModifiedNS,
		ContentHash:    payload.ContentHash,
	})
	if err != nil {
		var conflict *files.ConflictError
		if errors.As(err, &conflict) {
			writeJSON(w, http.StatusConflict, conflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCopyPath(w http.ResponseWriter, r *http.Request) {
	if !s.auth.ValidateCSRF(r) {
		http.Error(w, "invalid CSRF token", http.StatusForbidden)
		return
	}
	var payload struct {
		Path string `json:"path"`
		Kind string `json:"kind"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	value, err := s.files.CopyPath(payload.Path, payload.Kind)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"value": value})
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.openclaw.Summary(r.Context())
	data := summaryData{Summary: summary}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, "summary", data)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.openclaw.Sessions()
	data := sessionsData{Sessions: sessions}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, "sessions", data)
}

func (s *Server) handleCron(w http.ResponseWriter, r *http.Request) {
	cron, err := s.openclaw.Cron()
	data := cronData{Cron: cron}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, "cron", data)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	lines := 200
	if raw := r.URL.Query().Get("lines"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			lines = parsed
		}
	}
	date := time.Now()
	dateRaw := r.URL.Query().Get("date")
	if dateRaw != "" {
		if parsed, err := time.Parse("2006-01-02", dateRaw); err == nil {
			date = parsed
		}
	}
	tail, err := s.openclaw.LogTail(date, lines)
	data := logsData{
		Date:  date.Format("2006-01-02"),
		Lines: tail,
	}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, "logs", data)
}

func (s *Server) handleGatewayAction(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.auth.ValidateCSRF(r) {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}
		result, err := s.openclaw.GatewayAction(context.Background(), action)
		data := gatewayActionData{Result: result}
		if err != nil {
			data.Error = err.Error()
		}
		s.render(w, "gateway-action", data)
	}
}

func (s *Server) renderShell(w http.ResponseWriter, r *http.Request, selectedPath string) {
	user, _ := s.auth.CurrentUser(r)
	token := s.auth.EnsureCSRFCookie(w, r)
	s.render(w, "index", pageData{
		Title:        "openclaudio",
		User:         user,
		CSRFToken:    token,
		SelectedPath: strings.TrimPrefix(filepath.ToSlash(selectedPath), "/"),
	})
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderStatus(w http.ResponseWriter, status int, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
