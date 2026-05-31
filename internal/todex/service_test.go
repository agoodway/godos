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

func TestNoteServiceResolvesFoldersListsAndCreatesNotes(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		switch r.Method + " " + r.URL.Path {
		case "GET /api/note-folders":
			writeJSON(t, w, map[string]any{"data": map[string]any{"note_folders": []map[string]any{
				remoteNoteFolderJSON("11111111-1111-4111-8111-111111111111", "Notes", true, 0),
				remoteNoteFolderJSON("22222222-2222-4222-8222-222222222222", "Work", false, 1),
			}}})
		case "GET /api/notes":
			query := r.URL.Query()
			if query.Get("folder_id") != "22222222-2222-4222-8222-222222222222" || query.Get("q") != "plan" || query.Get("pinned") != "true" || query.Get("deleted") != "false" {
				t.Fatalf("unexpected note query: %q", r.URL.RawQuery)
			}
			writeJSON(t, w, map[string]any{"data": map[string]any{"notes": []map[string]any{
				remoteNoteJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "22222222-2222-4222-8222-222222222222", "Plan", "body", true, false, 0),
			}}})
		case "POST /api/notes":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["title"] != "New note" || body["folder_id"] != "11111111-1111-4111-8111-111111111111" {
				t.Fatalf("unexpected create note body: %#v", body)
			}
			writeJSONStatus(t, w, http.StatusCreated, map[string]any{"data": map[string]any{"note": remoteNoteJSON("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", "11111111-1111-4111-8111-111111111111", "New note", "", false, false, 0)}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}

	pinned := true
	deleted := false
	notes, err := svc.ListNotes(context.Background(), NoteFilters{FolderName: "Work", Query: "plan", Pinned: &pinned, Deleted: &deleted})
	if err != nil {
		t.Fatalf("ListNotes returned error: %v", err)
	}
	if len(notes) != 1 || notes[0].ShortID != "aaaaaaaa" || notes[0].Body != "body" || !notes[0].Pinned {
		t.Fatalf("unexpected notes: %#v", notes)
	}

	note, err := svc.CreateNote(context.Background(), "New note", "", "")
	if err != nil {
		t.Fatalf("CreateNote returned error: %v", err)
	}
	if note.ShortID != "bbbbbbbb" || note.Title != "New note" {
		t.Fatalf("unexpected created note: %#v", note)
	}

	want := []string{"GET /api/note-folders?", "GET /api/notes?deleted=false&folder_id=22222222-2222-4222-8222-222222222222&pinned=true&q=plan", "GET /api/note-folders?", "POST /api/notes?"}
	if strings.Join(requests, "|") != strings.Join(want, "|") {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
}

func TestNoteServiceRejectsMissingAndAmbiguousPrefixes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/notes" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeJSON(t, w, map[string]any{"data": map[string]any{"notes": []map[string]any{
			remoteNoteJSON("9f3a0000-0000-4000-8000-000000000000", "11111111-1111-4111-8111-111111111111", "one", "", false, false, 0),
			remoteNoteJSON("9f3abbbb-0000-4000-8000-000000000000", "11111111-1111-4111-8111-111111111111", "two", "", false, false, 1),
		}}})
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetNote(context.Background(), "missing"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
	if _, err := svc.GetNote(context.Background(), "9f3a"); err == nil || !strings.Contains(err.Error(), "more characters") {
		t.Fatalf("expected ambiguous error, got %v", err)
	}
}

func TestNoteServiceCrudAndLifecycleUseExpectedHTTPContracts(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		if r.Method == http.MethodGet && r.URL.Path == "/api/notes" {
			deleted := r.URL.Query().Get("deleted") == "true"
			writeJSON(t, w, map[string]any{"data": map[string]any{"notes": []map[string]any{
				remoteNoteJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "11111111-1111-4111-8111-111111111111", "note", "old", false, deleted, 0),
			}}})
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/api/notes/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" {
			writeJSON(t, w, map[string]any{"data": map[string]any{"note": remoteNoteJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "11111111-1111-4111-8111-111111111111", "note", "old", false, false, 0)}})
			return
		}
		if r.Method == http.MethodPatch && r.URL.Path == "/api/notes/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["body"] != "new body" {
				t.Fatalf("unexpected update body: %#v", body)
			}
			writeJSON(t, w, map[string]any{"data": map[string]any{"note": remoteNoteJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "11111111-1111-4111-8111-111111111111", "note", "new body", false, false, 0)}})
			return
		}
		if r.Method == http.MethodDelete && r.URL.Path == "/api/notes/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" {
			writeJSON(t, w, map[string]any{"data": map[string]any{"note": remoteNoteJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "11111111-1111-4111-8111-111111111111", "note", "old", false, true, 0)}})
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/restore") {
			writeJSON(t, w, map[string]any{"data": map[string]any{"note": remoteNoteJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "11111111-1111-4111-8111-111111111111", "note", "old", false, false, 0)}})
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/pin") {
			writeJSON(t, w, map[string]any{"data": map[string]any{"note": remoteNoteJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "11111111-1111-4111-8111-111111111111", "note", "old", true, false, 0)}})
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/unpin") {
			writeJSON(t, w, map[string]any{"data": map[string]any{"note": remoteNoteJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "11111111-1111-4111-8111-111111111111", "note", "old", false, false, 0)}})
			return
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if note, err := svc.GetNote(context.Background(), "aaaaaaaa"); err != nil || note.Body != "old" {
		t.Fatalf("GetNote = %#v, %v", note, err)
	}
	if note, err := svc.UpdateNoteBody(context.Background(), "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "new body"); err != nil || note.Body != "new body" {
		t.Fatalf("UpdateNoteBody = %#v, %v", note, err)
	}
	if _, err := svc.DeleteNote(context.Background(), "aaaaaaaa"); err != nil {
		t.Fatalf("DeleteNote returned error: %v", err)
	}
	if _, err := svc.RestoreNote(context.Background(), "aaaaaaaa"); err != nil {
		t.Fatalf("RestoreNote returned error: %v", err)
	}
	if note, err := svc.PinNote(context.Background(), "aaaaaaaa"); err != nil || !note.Pinned {
		t.Fatalf("PinNote = %#v, %v", note, err)
	}
	if note, err := svc.UnpinNote(context.Background(), "aaaaaaaa"); err != nil || note.Pinned {
		t.Fatalf("UnpinNote = %#v, %v", note, err)
	}
}

func TestGoalServiceCrudAndProgressData(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		switch r.Method + " " + r.URL.Path {
		case "GET /api/goals":
			writeJSON(t, w, map[string]any{"data": map[string]any{"goals": []map[string]any{
				remoteGoalJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "Launch", "Ship MVP", "Learn", 42),
			}}})
		case "POST /api/goals":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["title"] != "New goal" || body["description"] != "Ship" || body["reason"] != "Learn" {
				t.Fatalf("unexpected goal create body: %#v", body)
			}
			writeJSONStatus(t, w, http.StatusCreated, map[string]any{"data": map[string]any{"goal": remoteGoalJSON("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", "New goal", "Ship", "Learn", 0)}})
		case "GET /api/goals/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa":
			writeJSON(t, w, map[string]any{"data": map[string]any{"goal": remoteGoalJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "Launch", "Ship MVP", "Learn", 42)}})
		case "PATCH /api/goals/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["title"] != "Updated" || body["description"] != "Updated desc" || body["reason"] != "Updated reason" {
				t.Fatalf("unexpected goal update body: %#v", body)
			}
			writeJSON(t, w, map[string]any{"data": map[string]any{"goal": remoteGoalJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "Updated", "Updated desc", "Updated reason", 42)}})
		case "DELETE /api/goals/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa":
			writeJSON(t, w, map[string]any{"data": map[string]any{"goal": remoteGoalJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "Launch", "Ship MVP", "Learn", 42)}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	goals, err := svc.ListGoals(context.Background())
	if err != nil {
		t.Fatalf("ListGoals returned error: %v", err)
	}
	if len(goals) != 1 || goals[0].ShortID != "aaaaaaaa" || goals[0].Progress != 42 {
		t.Fatalf("unexpected goals: %#v", goals)
	}
	created, err := svc.CreateGoal(context.Background(), GoalChanges{Title: stringPtr("New goal"), Description: stringPtr("Ship"), Reason: stringPtr("Learn")})
	if err != nil || created.ShortID != "bbbbbbbb" {
		t.Fatalf("CreateGoal = %#v, %v", created, err)
	}
	shown, err := svc.GetGoal(context.Background(), "aaaaaaaa")
	if err != nil || shown.Title != "Launch" || shown.Description != "Ship MVP" || shown.Reason != "Learn" {
		t.Fatalf("GetGoal = %#v, %v", shown, err)
	}
	updated, err := svc.UpdateGoal(context.Background(), "aaaaaaaa", GoalChanges{Title: stringPtr("Updated"), Description: stringPtr("Updated desc"), Reason: stringPtr("Updated reason")})
	if err != nil || updated.Title != "Updated" {
		t.Fatalf("UpdateGoal = %#v, %v", updated, err)
	}
	if _, err := svc.DeleteGoal(context.Background(), "aaaaaaaa"); err != nil {
		t.Fatalf("DeleteGoal returned error: %v", err)
	}

	want := []string{"GET /api/goals?", "POST /api/goals?", "GET /api/goals?", "GET /api/goals/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa?", "GET /api/goals?", "PATCH /api/goals/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa?", "GET /api/goals?", "DELETE /api/goals/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa?"}
	if strings.Join(requests, "|") != strings.Join(want, "|") {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
}

func TestGoalServiceRejectsMissingAndAmbiguousPrefixes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/goals" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeJSON(t, w, map[string]any{"data": map[string]any{"goals": []map[string]any{
			remoteGoalJSON("9f3a0000-0000-4000-8000-000000000000", "one", "", "", 0),
			remoteGoalJSON("9f3abbbb-0000-4000-8000-000000000000", "two", "", "", 0),
		}}})
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetGoal(context.Background(), "missing"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
	if _, err := svc.GetGoal(context.Background(), "9f3a"); err == nil || !strings.Contains(err.Error(), "more characters") {
		t.Fatalf("expected ambiguous error, got %v", err)
	}
}

func TestGoalServiceLinkAndUnlinkResolveGoalAndTaskPrefixes(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/goals":
			writeJSON(t, w, map[string]any{"data": map[string]any{"goals": []map[string]any{
				remoteGoalJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "Launch", "", "", 0),
			}}})
		case r.Method == http.MethodGet && r.URL.Path == "/api/tasks":
			writeJSON(t, w, map[string]any{"data": map[string]any{"tasks": []map[string]any{
				{"id": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", "list_id": "11111111-1111-4111-8111-111111111111", "title": "task", "status": "active", "position": 0, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"},
			}}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/goals/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/tasks":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["task_id"] != "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb" {
				t.Fatalf("unexpected link body: %#v", body)
			}
			writeJSON(t, w, map[string]any{"data": map[string]any{"goal": remoteGoalJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "Launch", "", "", 20)}})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/goals/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/tasks/bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb":
			writeJSON(t, w, map[string]any{"data": map[string]any{"goal": remoteGoalJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "Launch", "", "", 0)}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if goal, err := svc.LinkGoalTask(context.Background(), "aaaaaaaa", "bbbbbbbb"); err != nil || goal.Progress != 20 {
		t.Fatalf("LinkGoalTask = %#v, %v", goal, err)
	}
	if goal, err := svc.UnlinkGoalTask(context.Background(), "aaaaaaaa", "bbbbbbbb"); err != nil || goal.Progress != 0 {
		t.Fatalf("UnlinkGoalTask = %#v, %v", goal, err)
	}
}

func TestCreateNoteSendsBodyInSingleCreateCall(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch r.Method + " " + r.URL.Path {
		case "GET /api/note-folders":
			writeJSON(t, w, map[string]any{"data": map[string]any{"note_folders": []map[string]any{
				remoteNoteFolderJSON("11111111-1111-4111-8111-111111111111", "Notes", true, 0),
			}}})
		case "POST /api/notes":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["body"] != "the body" {
				t.Fatalf("expected body in create request, got %#v", body)
			}
			writeJSONStatus(t, w, http.StatusCreated, map[string]any{"data": map[string]any{"note": remoteNoteJSON("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", "11111111-1111-4111-8111-111111111111", "New note", "the body", false, false, 0)}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	note, err := svc.CreateNote(context.Background(), "New note", "", "the body")
	if err != nil {
		t.Fatalf("CreateNote returned error: %v", err)
	}
	if note.Body != "the body" {
		t.Fatalf("unexpected note body: %#v", note)
	}
	patchCalls := 0
	for _, req := range requests {
		if strings.HasPrefix(req, "PATCH ") {
			patchCalls++
		}
	}
	if patchCalls != 0 {
		t.Fatalf("expected no follow-up PATCH, requests: %#v", requests)
	}
}

func TestCreateNoteOmitsEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/note-folders":
			writeJSON(t, w, map[string]any{"data": map[string]any{"note_folders": []map[string]any{
				remoteNoteFolderJSON("11111111-1111-4111-8111-111111111111", "Notes", true, 0),
			}}})
		case "POST /api/notes":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if _, ok := body["body"]; ok {
				t.Fatalf("expected body to be omitted, got %#v", body)
			}
			writeJSONStatus(t, w, http.StatusCreated, map[string]any{"data": map[string]any{"note": remoteNoteJSON("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", "11111111-1111-4111-8111-111111111111", "New note", "", false, false, 0)}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateNote(context.Background(), "New note", "", ""); err != nil {
		t.Fatalf("CreateNote returned error: %v", err)
	}
}

func TestActiveNoteOperationsScopeToNonDeleted(t *testing.T) {
	noteID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	cases := []struct {
		name string
		call func(svc *Service) error
	}{
		{"show", func(svc *Service) error { _, err := svc.GetNote(context.Background(), "aaaaaaaa"); return err }},
		{"rm", func(svc *Service) error { _, err := svc.DeleteNote(context.Background(), "aaaaaaaa"); return err }},
		{"pin", func(svc *Service) error { _, err := svc.PinNote(context.Background(), "aaaaaaaa"); return err }},
		{"unpin", func(svc *Service) error { _, err := svc.UnpinNote(context.Background(), "aaaaaaaa"); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var listQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet && r.URL.Path == "/api/notes" {
					listQuery = r.URL.RawQuery
					writeJSON(t, w, map[string]any{"data": map[string]any{"notes": []map[string]any{
						remoteNoteJSON(noteID, "11111111-1111-4111-8111-111111111111", "note", "old", false, false, 0),
					}}})
					return
				}
				writeJSON(t, w, map[string]any{"data": map[string]any{"note": remoteNoteJSON(noteID, "11111111-1111-4111-8111-111111111111", "note", "old", false, false, 0)}})
			}))
			defer server.Close()

			svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
			if err != nil {
				t.Fatal(err)
			}
			if err := tc.call(svc); err != nil {
				t.Fatalf("%s returned error: %v", tc.name, err)
			}
			if listQuery != "deleted=false" {
				t.Fatalf("expected deleted=false resolution query, got %q", listQuery)
			}
		})
	}
}

func TestRestoreNoteScopesToDeleted(t *testing.T) {
	noteID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	var listQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/notes" {
			listQuery = r.URL.RawQuery
			writeJSON(t, w, map[string]any{"data": map[string]any{"notes": []map[string]any{
				remoteNoteJSON(noteID, "11111111-1111-4111-8111-111111111111", "note", "old", false, true, 0),
			}}})
			return
		}
		writeJSON(t, w, map[string]any{"data": map[string]any{"note": remoteNoteJSON(noteID, "11111111-1111-4111-8111-111111111111", "note", "old", false, false, 0)}})
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RestoreNote(context.Background(), "aaaaaaaa"); err != nil {
		t.Fatalf("RestoreNote returned error: %v", err)
	}
	if listQuery != "deleted=true" {
		t.Fatalf("expected deleted=true resolution query, got %q", listQuery)
	}
}

func TestUpdateGoalRejectsWhitespaceTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/goals" {
			writeJSON(t, w, map[string]any{"data": map[string]any{"goals": []map[string]any{
				remoteGoalJSON("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "Launch", "", "", 0),
			}}})
			return
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateGoal(context.Background(), "aaaaaaaa", GoalChanges{Title: stringPtr("   ")}); err == nil || !strings.Contains(err.Error(), "title cannot be empty") {
		t.Fatalf("expected empty title error, got %v", err)
	}
}

func TestResolveNotePrefix(t *testing.T) {
	notes := []Note{
		{ID: "9f3a0000-0000-4000-8000-000000000000", Title: "one"},
		{ID: "9f3abbbb-0000-4000-8000-000000000000", Title: "two"},
	}
	if _, err := resolveNotePrefix(notes, ""); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error, got %v", err)
	}
	if _, err := resolveNotePrefix(notes, "deadbeef"); !errors.Is(err, ErrNoteNotFound) {
		t.Fatalf("expected ErrNoteNotFound, got %v", err)
	}
	if _, err := resolveNotePrefix(notes, "9f3a"); !errors.Is(err, ErrAmbiguousNoteID) {
		t.Fatalf("expected ErrAmbiguousNoteID, got %v", err)
	}
	id, err := resolveNotePrefix(notes, "9f3ab")
	if err != nil || id != "9f3abbbb-0000-4000-8000-000000000000" {
		t.Fatalf("expected unique match, got %q %v", id, err)
	}
}

func TestResolveGoalPrefix(t *testing.T) {
	goals := []Goal{
		{ID: "9f3a0000-0000-4000-8000-000000000000", Title: "one"},
		{ID: "9f3abbbb-0000-4000-8000-000000000000", Title: "two"},
	}
	if _, err := resolveGoalPrefix(goals, ""); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error, got %v", err)
	}
	if _, err := resolveGoalPrefix(goals, "deadbeef"); !errors.Is(err, ErrGoalNotFound) {
		t.Fatalf("expected ErrGoalNotFound, got %v", err)
	}
	if _, err := resolveGoalPrefix(goals, "9f3a"); !errors.Is(err, ErrAmbiguousGoalID) {
		t.Fatalf("expected ErrAmbiguousGoalID, got %v", err)
	}
	id, err := resolveGoalPrefix(goals, "9f3ab")
	if err != nil || id != "9f3abbbb-0000-4000-8000-000000000000" {
		t.Fatalf("expected unique match, got %q %v", id, err)
	}
}

func TestResolveNoteFolder(t *testing.T) {
	cases := []struct {
		name     string
		folders  []map[string]any
		input    string
		wantErr  string
		sentinel error
	}{
		{
			name: "duplicate names",
			folders: []map[string]any{
				remoteNoteFolderJSON("11111111-1111-4111-8111-111111111111", "Work", false, 0),
				remoteNoteFolderJSON("22222222-2222-4222-8222-222222222222", "Work", false, 1),
			},
			input:   "Work",
			wantErr: "duplicate remote note folder name",
		},
		{
			name: "named not found",
			folders: []map[string]any{
				remoteNoteFolderJSON("11111111-1111-4111-8111-111111111111", "Notes", true, 0),
			},
			input:    "Missing",
			sentinel: ErrNoteFolderNotFound,
		},
		{
			name: "no default fallback",
			folders: []map[string]any{
				remoteNoteFolderJSON("11111111-1111-4111-8111-111111111111", "Work", false, 0),
			},
			input:    "",
			wantErr:  "no default note folder configured",
			sentinel: ErrNoteFolderNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeJSON(t, w, map[string]any{"data": map[string]any{"note_folders": tc.folders}})
			}))
			defer server.Close()

			svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
			if err != nil {
				t.Fatal(err)
			}
			_, err = svc.resolveNoteFolder(context.Background(), tc.input)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
			if tc.sentinel != nil && !errors.Is(err, tc.sentinel) {
				t.Fatalf("expected sentinel %v, got %v", tc.sentinel, err)
			}
		})
	}
}

func TestListNotesPropagatesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONStatus(t, w, http.StatusBadRequest, map[string]any{"error": map[string]any{"code": "bad_request", "message": "bad filter", "details": map[string]any{}}})
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ListNotes(context.Background(), NoteFilters{}); err == nil || !strings.Contains(err.Error(), "bad filter") {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestCreateNotePropagatesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/note-folders":
			writeJSON(t, w, map[string]any{"data": map[string]any{"note_folders": []map[string]any{
				remoteNoteFolderJSON("11111111-1111-4111-8111-111111111111", "Notes", true, 0),
			}}})
		case "POST /api/notes":
			writeJSONStatus(t, w, http.StatusUnprocessableEntity, map[string]any{"error": map[string]any{"code": "unprocessable", "message": "title taken", "details": map[string]any{}}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateNote(context.Background(), "New", "", ""); err == nil || !strings.Contains(err.Error(), "title taken") {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestListGoalsPropagatesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONStatus(t, w, http.StatusBadRequest, map[string]any{"error": map[string]any{"code": "bad_request", "message": "bad request", "details": map[string]any{}}})
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ListGoals(context.Background()); err == nil || !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestCreateGoalPropagatesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONStatus(t, w, http.StatusUnprocessableEntity, map[string]any{"error": map[string]any{"code": "unprocessable", "message": "goal rejected", "details": map[string]any{}}})
	}))
	defer server.Close()

	svc, err := New(ServiceConfig{BaseURL: server.URL, Token: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateGoal(context.Background(), GoalChanges{Title: stringPtr("Goal")}); err == nil || !strings.Contains(err.Error(), "goal rejected") {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestListNotesAndGoalsTransportError(t *testing.T) {
	svc, err := New(ServiceConfig{BaseURL: "http://127.0.0.1:1", Token: "token"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ListNotes(context.Background(), NoteFilters{}); err == nil || !strings.Contains(err.Error(), "connecting to Todex API") {
		t.Fatalf("expected transport error for ListNotes, got %v", err)
	}
	if _, err := svc.ListGoals(context.Background()); err == nil || !strings.Contains(err.Error(), "connecting to Todex API") {
		t.Fatalf("expected transport error for ListGoals, got %v", err)
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

func remoteNoteFolderJSON(id, name string, isDefault bool, position int) map[string]any {
	return map[string]any{"id": id, "name": name, "is_default": isDefault, "position": position, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
}

func remoteNoteJSON(id, folderID, title, body string, pinned, deleted bool, position int) map[string]any {
	note := map[string]any{"id": id, "folder_id": folderID, "title": title, "body": body, "pinned": pinned, "position": position, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
	if deleted {
		note["deleted_at"] = "2026-01-02T00:00:00Z"
	}
	return note
}

func remoteGoalJSON(id, title, description, reason string, progress int) map[string]any {
	return map[string]any{"id": id, "title": title, "description": description, "reason": reason, "progress": progress, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
}

func stringPtr(value string) *string {
	return &value
}
