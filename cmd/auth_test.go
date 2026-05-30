package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goodway/godos/config"
)

func TestLoginPersistsReturnedToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	server := authServer(t, http.StatusOK, `{"data":{"token":"login-token","user":{"id":"11111111-1111-4111-8111-111111111111","email":"user@example.com","inserted_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}}}`)
	defer server.Close()
	t.Setenv(config.APIBaseURLEnv, server.URL)

	out, err := executeCommandWithInput(t, strings.NewReader("secret\n"), "login", "user@example.com")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if !strings.Contains(out, "Logged in") {
		t.Fatalf("expected login success output, got %q", out)
	}
	token, err := config.Get(config.APITokenKey)
	if err != nil || token != "login-token" {
		t.Fatalf("expected persisted token, got %q, %v", token, err)
	}
}

func TestLoginFailureDoesNotOverwriteToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.Set(config.APITokenKey, "old-token"); err != nil {
		t.Fatal(err)
	}
	server := authServer(t, http.StatusUnauthorized, `{"error":{"code":"unauthorized","message":"bad credentials","details":{}}}`)
	defer server.Close()
	t.Setenv(config.APIBaseURLEnv, server.URL)

	_, err := executeCommandWithInput(t, strings.NewReader("wrong\n"), "login", "user@example.com")
	if err == nil || !strings.Contains(err.Error(), "bad credentials") {
		t.Fatalf("expected bad credentials error, got %v", err)
	}
	token, err := config.Get(config.APITokenKey)
	if err != nil || token != "old-token" {
		t.Fatalf("expected old token preserved, got %q, %v", token, err)
	}
}

func TestRegisterPersistsReturnedToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	server := authServer(t, http.StatusCreated, `{"data":{"token":"register-token","user":{"id":"11111111-1111-4111-8111-111111111111","email":"user@example.com","inserted_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}}}`)
	defer server.Close()
	t.Setenv(config.APIBaseURLEnv, server.URL)

	_, err := executeCommandWithInput(t, strings.NewReader("secret\n"), "register", "user@example.com")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	token, err := config.Get(config.APITokenKey)
	if err != nil || token != "register-token" {
		t.Fatalf("expected persisted token, got %q, %v", token, err)
	}
}

func TestRegisterFailureDoesNotOverwriteToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.Set(config.APITokenKey, "old-token"); err != nil {
		t.Fatal(err)
	}
	server := authServer(t, http.StatusUnprocessableEntity, `{"error":{"code":"validation","message":"email has already been taken","details":{}}}`)
	defer server.Close()
	t.Setenv(config.APIBaseURLEnv, server.URL)

	_, err := executeCommandWithInput(t, strings.NewReader("secret\n"), "register", "user@example.com")
	if err == nil || !strings.Contains(err.Error(), "email has already been taken") {
		t.Fatalf("expected registration validation error, got %v", err)
	}
	token, err := config.Get(config.APITokenKey)
	if err != nil || token != "old-token" {
		t.Fatalf("expected old token preserved, got %q, %v", token, err)
	}
}

func TestAuthStatusReportsCurrentUser(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.Set(config.APITokenKey, "stored-token"); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/me" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer stored-token" {
			t.Fatalf("expected auth header, got %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"user":{"id":"11111111-1111-4111-8111-111111111111","email":"user@example.com","inserted_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()
	t.Setenv(config.APIBaseURLEnv, server.URL)

	out, err := executeCommand(t, "auth", "status")
	if err != nil {
		t.Fatalf("auth status failed: %v", err)
	}
	if !strings.Contains(out, "user@example.com") {
		t.Fatalf("expected current user email, got %q", out)
	}
}

func TestAuthStatusExpiredTokenReportsLoginGuidance(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.Set(config.APITokenKey, "expired-token"); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":"unauthorized","message":"token expired","details":{}}}`))
	}))
	defer server.Close()
	t.Setenv(config.APIBaseURLEnv, server.URL)

	_, err := executeCommand(t, "auth", "status")
	if err == nil || !strings.Contains(err.Error(), "login") {
		t.Fatalf("expected login guidance for expired token, got %v", err)
	}
}

func TestLogoutClearsLocalTokenEvenWhenAPIRequestFails(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.Set(config.APITokenKey, "old-token"); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/logout" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer old-token" {
			t.Fatalf("expected logout auth header, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv(config.APIBaseURLEnv, server.URL)

	_, err := executeCommand(t, "logout")
	if err != nil {
		t.Fatalf("logout should clear token despite API error: %v", err)
	}
	if _, err := config.Get(config.APITokenKey); err == nil {
		t.Fatal("expected token to be cleared")
	}
}

func TestLogoutWithEnvTokenAndNoStoredTokenSucceeds(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(config.APITokenEnv, "env-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/logout" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer env-token" {
			t.Fatalf("expected env token auth header, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	t.Setenv(config.APIBaseURLEnv, server.URL)

	out, err := executeCommand(t, "logout")
	if err != nil {
		t.Fatalf("logout with env token should succeed: %v", err)
	}
	if !strings.Contains(out, "Logged out") {
		t.Fatalf("expected logged out output, got %q", out)
	}
}

func TestLogoutWithoutTokenSucceeds(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	out, err := executeCommand(t, "logout")
	if err != nil {
		t.Fatalf("logout without token failed: %v", err)
	}
	if !strings.Contains(out, "No active session") {
		t.Fatalf("expected no session output, got %q", out)
	}
}

func TestAuthenticatedCommandRequiresBaseURLAndToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := executeCommand(t, "list")
	if err == nil || !strings.Contains(err.Error(), config.APIBaseURLKey) {
		t.Fatalf("expected missing base URL error, got %v", err)
	}

	t.Setenv(config.APIBaseURLEnv, "http://127.0.0.1:1")
	_, err = executeCommand(t, "list")
	if err == nil || !strings.Contains(err.Error(), "login") {
		t.Fatalf("expected missing token login guidance, got %v", err)
	}
}

func authServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/login" && r.URL.Path != "/api/auth/register" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("auth request should not send bearer token, got %q", r.Header.Get("Authorization"))
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Fatalf("expected JSON content type, got %q", ct)
		}
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding auth request: %v", err)
		}
		if req.Email != "user@example.com" || req.Password == "" {
			t.Fatalf("unexpected auth request body: %#v", req)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}
