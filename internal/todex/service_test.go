package todex

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServiceSendsBearerTokenAndListsSummaries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			t.Fatalf("expected bearer token header, got %q", r.Header.Get("Authorization"))
		}

		switch r.URL.Path {
		case "/api/lists":
			writeJSON(t, w, map[string]any{"data": map[string]any{"lists": []map[string]any{
				{"id": "11111111-1111-4111-8111-111111111111", "name": "todo", "position": 0, "is_default": true, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"},
			}}})
		case "/api/tasks":
			if r.URL.RawQuery != "" {
				t.Fatalf("expected unfiltered task request, got %q", r.URL.RawQuery)
			}
			writeJSON(t, w, map[string]any{"data": map[string]any{"tasks": []map[string]any{
				{"id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "list_id": "11111111-1111-4111-8111-111111111111", "title": "open", "status": "active", "position": 0, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"},
				{"id": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", "list_id": "11111111-1111-4111-8111-111111111111", "title": "done", "status": "completed", "position": 1, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"},
			}}})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "secret-token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	summaries, err := svc.ListSummaries(context.Background())
	if err != nil {
		t.Fatalf("ListSummaries returned error: %v", err)
	}
	if len(summaries) != 1 || summaries[0].Name != "todo" || summaries[0].Completed != 1 || summaries[0].Total != 2 {
		t.Fatalf("unexpected summaries: %#v", summaries)
	}
}

func TestAddTaskAutoCreatesMissingList(t *testing.T) {
	var createdList bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/lists":
			lists := []map[string]any{}
			if createdList {
				lists = append(lists, map[string]any{"id": "22222222-2222-4222-8222-222222222222", "name": "ideas", "position": 0, "is_default": false, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"})
			}
			writeJSON(t, w, map[string]any{"data": map[string]any{"lists": lists}})
		case "POST /api/lists":
			createdList = true
			writeJSONStatus(t, w, http.StatusCreated, map[string]any{"data": map[string]any{"list": map[string]any{"id": "22222222-2222-4222-8222-222222222222", "name": "ideas", "position": 0, "is_default": false, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}}})
		case "POST /api/tasks":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["title"] != "Try Todex" || body["list_id"] != "22222222-2222-4222-8222-222222222222" {
				t.Fatalf("unexpected task body: %#v", body)
			}
			writeJSONStatus(t, w, http.StatusCreated, map[string]any{"data": map[string]any{"task": map[string]any{"id": "cccccccc-cccc-4ccc-8ccc-cccccccccccc", "list_id": "22222222-2222-4222-8222-222222222222", "title": "Try Todex", "status": "active", "position": 0, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}

	task, err := svc.AddTask(context.Background(), "ideas", "Try Todex")
	if err != nil {
		t.Fatalf("AddTask returned error: %v", err)
	}
	if task.ShortID != "cccccccc" || task.Title != "Try Todex" {
		t.Fatalf("unexpected task: %#v", task)
	}
}

func TestResolveTaskPrefixRejectsMissingAmbiguousAndNumeric(t *testing.T) {
	svc := &Service{tasks: []Task{
		{ID: "9f3a0000-0000-4000-8000-000000000000", Title: "one"},
		{ID: "9f3abbbb-0000-4000-8000-000000000000", Title: "two"},
	}}

	if _, err := svc.ResolveTaskPrefix("3"); err == nil || !strings.Contains(err.Error(), "task ID prefix") {
		t.Fatalf("expected numeric prefix error, got %v", err)
	}
	if _, err := svc.ResolveTaskPrefix("deadbeef"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
	if _, err := svc.ResolveTaskPrefix("9f3a"); err == nil || !strings.Contains(err.Error(), "more characters") {
		t.Fatalf("expected ambiguous error, got %v", err)
	}
	id, err := svc.ResolveTaskPrefix("9f3ab")
	if err != nil {
		t.Fatalf("expected unique match, got %v", err)
	}
	if id != "9f3abbbb-0000-4000-8000-000000000000" {
		t.Fatalf("unexpected id %q", id)
	}
}

func TestServiceNormalizesConnectionFailures(t *testing.T) {
	svc, err := New(ServiceConfig{BaseURL: "http://127.0.0.1:1", Token: "token"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.ListSummaries(context.Background())
	if err == nil || !strings.Contains(err.Error(), "connecting to Todex API") {
		t.Fatalf("expected connection error, got %v", err)
	}
}

func TestServiceNormalizesMalformedResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"lists":`))
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.ListSummaries(context.Background())
	if err == nil || !strings.Contains(err.Error(), "malformed response") {
		t.Fatalf("expected malformed response error, got %v", err)
	}
}

func TestNewRejectsUnsafeBaseURL(t *testing.T) {
	if _, err := New(ServiceConfig{BaseURL: "http://api.example.com", Token: "token"}); err == nil {
		t.Fatal("expected unsafe HTTP base URL to be rejected")
	}
}

func TestNewDefaultClientHasTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token"})
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	_, err = svc.ListSummaries(context.Background())
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if time.Since(start) > 1500*time.Millisecond {
		t.Fatalf("expected request to time out quickly, took %s", time.Since(start))
	}
}

func TestServiceLimitsResponseBodySize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(strings.Repeat("x", 2*1024*1024)))
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.ListSummaries(context.Background())
	if err == nil || !strings.Contains(err.Error(), "response too large") {
		t.Fatalf("expected response size error, got %v", err)
	}
}

func TestServiceMapsListConflictToSentinel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/lists" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeJSONStatus(t, w, http.StatusConflict, map[string]any{"error": map[string]any{"code": "conflict", "message": "list already exists", "details": map[string]any{}}})
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateList(context.Background(), "todo")
	if !errors.Is(err, ErrListExists) {
		t.Fatalf("expected ErrListExists, got %v", err)
	}
}

func TestServiceCompleteDeleteAndReopenUseExpectedHTTPContracts(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("expected bearer token, got %q", r.Header.Get("Authorization"))
		}
		if r.Method == http.MethodGet && r.URL.Path == "/api/tasks" {
			writeJSON(t, w, map[string]any{"data": map[string]any{"tasks": []map[string]any{
				{"id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "list_id": "11111111-1111-4111-8111-111111111111", "title": "task", "status": "active", "position": 0, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"},
			}}})
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/complete") {
			writeJSON(t, w, map[string]any{"data": map[string]any{"task": map[string]any{"id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "list_id": "11111111-1111-4111-8111-111111111111", "title": "task", "status": "completed", "position": 0, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}}})
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/reopen") {
			writeJSON(t, w, map[string]any{"data": map[string]any{"task": map[string]any{"id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "list_id": "11111111-1111-4111-8111-111111111111", "title": "task", "status": "active", "position": 0, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}}})
			return
		}
		if r.Method == http.MethodDelete && r.URL.Path == "/api/tasks/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" {
			writeJSON(t, w, map[string]any{"data": map[string]any{"task": map[string]any{"id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "list_id": "11111111-1111-4111-8111-111111111111", "title": "task", "status": "active", "position": 0, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}}})
			return
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CompleteTask(context.Background(), "aaaaaaaa"); err != nil {
		t.Fatalf("CompleteTask returned error: %v", err)
	}
	if _, err := svc.ReopenTask(context.Background(), "aaaaaaaa"); err != nil {
		t.Fatalf("ReopenTask returned error: %v", err)
	}
	if _, err := svc.DeleteTask(context.Background(), "aaaaaaaa"); err != nil {
		t.Fatalf("DeleteTask returned error: %v", err)
	}

	want := []string{
		"GET /api/tasks", "POST /api/tasks/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/complete",
		"GET /api/tasks", "POST /api/tasks/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/reopen",
		"GET /api/tasks", "DELETE /api/tasks/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	}
	if strings.Join(requests, "|") != strings.Join(want, "|") {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}

func writeJSONStatus(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}
