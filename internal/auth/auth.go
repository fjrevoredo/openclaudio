package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookie = "openclaudio_session"
	csrfCookie    = "openclaudio_csrf"
)

type Manager struct {
	secret []byte
}

type Session struct {
	Username string `json:"username"`
	IssuedAt int64  `json:"issuedAt"`
}

func New(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (m *Manager) Login(w http.ResponseWriter, r *http.Request, username string) error {
	payload, err := json.Marshal(Session{
		Username: username,
		IssuedAt: time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	value := m.sign(payload)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
		MaxAge:   60 * 60 * 24 * 7,
	})
	return nil
}

func (m *Manager) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
		MaxAge:   -1,
	})
}

func (m *Manager) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := m.CurrentUser(r); err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Manager) CurrentUser(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return "", errors.New("missing session cookie")
	}
	payload, err := m.verify(cookie.Value)
	if err != nil {
		return "", err
	}

	var session Session
	if err := json.Unmarshal(payload, &session); err != nil {
		return "", err
	}
	if session.Username == "" {
		return "", errors.New("invalid session")
	}
	return session.Username, nil
}

func (m *Manager) EnsureCSRFCookie(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(csrfCookie); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	token := randomToken()
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
		MaxAge:   60 * 60 * 24 * 7,
	})
	return token
}

func (m *Manager) CSRFFromRequest(r *http.Request) string {
	token := r.Header.Get("X-CSRF-Token")
	if token != "" {
		return token
	}
	if err := r.ParseForm(); err == nil {
		return r.Form.Get("csrf_token")
	}
	return ""
}

func (m *Manager) ValidateCSRF(r *http.Request) bool {
	token := m.CSRFFromRequest(r)
	if token == "" {
		return false
	}
	cookie, err := r.Cookie(csrfCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	return hmac.Equal([]byte(token), []byte(cookie.Value))
}

func (m *Manager) sign(payload []byte) string {
	payloadEnc := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payloadEnc))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payloadEnc + "." + sig
}

func (m *Manager) verify(value string) ([]byte, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid session format")
	}

	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(parts[0]))
	actualSig := mac.Sum(nil)
	givenSig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}
	if !hmac.Equal(actualSig, givenSig) {
		return nil, errors.New("invalid session signature")
	}
	return base64.RawURLEncoding.DecodeString(parts[0])
}

func randomToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

func isSecureRequest(r *http.Request) bool {
	if r != nil && r.TLS != nil {
		return true
	}
	if r == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))) {
	case "https":
		return true
	}
	switch strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Ssl"))) {
	case "on", "1", "true":
		return true
	}
	return false
}
