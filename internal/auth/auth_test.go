package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginSetsSecureCookieForHTTPS(t *testing.T) {
	manager := New("secret")
	req := httptest.NewRequest(http.MethodGet, "https://example.test/login", nil)
	rec := httptest.NewRecorder()

	if err := manager.Login(rec, req, "admin"); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || !cookies[0].Secure {
		t.Fatalf("session cookie Secure = false, want true")
	}
}

func TestEnsureCSRFCookieHonorsForwardedProto(t *testing.T) {
	manager := New("secret")
	req := httptest.NewRequest(http.MethodGet, "http://example.test/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	manager.EnsureCSRFCookie(rec, req)

	found := false
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == csrfCookie {
			found = true
			if !cookie.Secure {
				t.Fatalf("csrf cookie Secure = false, want true")
			}
		}
	}
	if !found {
		t.Fatal("csrf cookie not set")
	}
}
