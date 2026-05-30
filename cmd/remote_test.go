package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goodway/godos/config"
)

type remoteState struct {
	lists     []remoteList
	tasks     []remoteTask
	requests  []string
	listIDHit int
}

type remoteList struct {
	ID   string
	Name string
}

type remoteTask struct {
	ID     string
	ListID string
	Title  string
	Status string
}

func setupRemoteCommandTest(t *testing.T) *remoteState {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	state := &remoteState{}
	server := httptest.NewServer(http.HandlerFunc(state.handle))
	t.Cleanup(server.Close)
	t.Setenv(config.APIBaseURLEnv, server.URL)
	t.Setenv(config.APITokenEnv, "test-token")
	return state
}

func (s *remoteState) handle(w http.ResponseWriter, r *http.Request) {
	s.requests = append(s.requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
	if r.Header.Get("Authorization") != "Bearer test-token" {
		writeRemoteError(w, http.StatusUnauthorized, "missing auth")
		return
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/lists":
		lists := make([]map[string]any, 0, len(s.lists))
		for i, list := range s.lists {
			lists = append(lists, remoteListJSON(list.ID, list.Name, i))
		}
		writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"lists": lists}})
	case r.Method == http.MethodPost && r.URL.Path == "/api/lists":
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		name, _ := body["name"].(string)
		if strings.ContainsAny(name, " !") {
			writeRemoteError(w, http.StatusUnprocessableEntity, "invalid list name")
			return
		}
		for _, list := range s.lists {
			if list.Name == name {
				writeRemoteError(w, http.StatusConflict, "list already exists")
				return
			}
		}
		list := remoteList{ID: nextRemoteID(len(s.lists) + 1), Name: name}
		s.lists = append(s.lists, list)
		writeRemoteJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{"list": remoteListJSON(list.ID, list.Name, len(s.lists)-1)}})
	case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/lists/"):
		id := strings.TrimPrefix(r.URL.Path, "/api/lists/")
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		for i := range s.lists {
			if s.lists[i].ID == id {
				s.lists[i].Name, _ = body["name"].(string)
				writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"list": remoteListJSON(s.lists[i].ID, s.lists[i].Name, i)}})
				return
			}
		}
		writeRemoteError(w, http.StatusNotFound, "list not found")
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/lists/"):
		id := strings.TrimPrefix(r.URL.Path, "/api/lists/")
		for i, list := range s.lists {
			if list.ID == id {
				s.lists = append(s.lists[:i], s.lists[i+1:]...)
				writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"list": remoteListJSON(list.ID, list.Name, i)}})
				return
			}
		}
		writeRemoteError(w, http.StatusNotFound, "list not found")
	case r.Method == http.MethodGet && r.URL.Path == "/api/tasks":
		listID := r.URL.Query().Get("list_id")
		tasks := make([]map[string]any, 0, len(s.tasks))
		for i, task := range s.tasks {
			if listID == "" || task.ListID == listID {
				tasks = append(tasks, remoteTaskJSON(task.ID, task.ListID, task.Title, task.Status, i))
			}
		}
		writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"tasks": tasks}})
	case r.Method == http.MethodPost && r.URL.Path == "/api/tasks":
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		task := remoteTask{ID: nextRemoteTaskID(len(s.tasks) + 1), ListID: body["list_id"].(string), Title: body["title"].(string), Status: "active"}
		s.tasks = append(s.tasks, task)
		writeRemoteJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{"task": remoteTaskJSON(task.ID, task.ListID, task.Title, task.Status, len(s.tasks)-1)}})
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/complete"):
		s.updateTaskStatus(w, strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/tasks/"), "/complete"), "completed")
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/reopen"):
		s.updateTaskStatus(w, strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/tasks/"), "/reopen"), "active")
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/tasks/"):
		id := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		for i, task := range s.tasks {
			if task.ID == id {
				s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
				writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"task": remoteTaskJSON(task.ID, task.ListID, task.Title, task.Status, i)}})
				return
			}
		}
		writeRemoteError(w, http.StatusNotFound, "task not found")
	default:
		writeRemoteError(w, http.StatusNotFound, "not found")
	}
}

func TestDoneAndRemoveRejectListFlag(t *testing.T) {
	setupRemoteCommandTest(t)

	if _, err := executeCommand(t, "done", "aaaaaaaa", "--list", "work"); err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("expected done --list to be rejected, got %v", err)
	}
	if _, err := executeCommand(t, "rm", "aaaaaaaa", "--list", "work"); err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("expected rm --list to be rejected, got %v", err)
	}
}

func TestUndoneCommandReopensRemoteTask(t *testing.T) {
	state := setupRemoteCommandTest(t)
	state.lists = []remoteList{{ID: nextRemoteID(1), Name: "todo"}}
	state.tasks = []remoteTask{{ID: nextRemoteTaskID(1), ListID: nextRemoteID(1), Title: "write tests", Status: "completed"}}

	out, err := executeCommand(t, "undone", "aaaaaaaa")
	if err != nil {
		t.Fatalf("undone failed: %v", err)
	}
	if state.tasks[0].Status != "active" {
		t.Fatalf("expected task to be active, got %q", state.tasks[0].Status)
	}
	if !strings.Contains(out, "Reopened aaaaaaaa") {
		t.Fatalf("expected reopened output, got %q", out)
	}
}

func TestListAllDoesNotRefetchEachList(t *testing.T) {
	state := setupRemoteCommandTest(t)
	state.lists = []remoteList{{ID: nextRemoteID(1), Name: "todo"}, {ID: nextRemoteID(2), Name: "work"}}
	state.tasks = []remoteTask{
		{ID: nextRemoteTaskID(1), ListID: nextRemoteID(1), Title: "one", Status: "active"},
		{ID: nextRemoteTaskID(2), ListID: nextRemoteID(2), Title: "two", Status: "completed"},
	}

	out, err := executeCommand(t, "list", "--all")
	if err != nil {
		t.Fatalf("list --all failed: %v", err)
	}
	if !strings.Contains(out, "=== todo ===") || !strings.Contains(out, "=== work ===") {
		t.Fatalf("expected both lists in output, got %q", out)
	}
	for _, req := range state.requests {
		if strings.Contains(req, "list_id=") {
			t.Fatalf("list --all should fetch tasks once without per-list refetches, requests: %#v", state.requests)
		}
	}
}

func (s *remoteState) updateTaskStatus(w http.ResponseWriter, id, status string) {
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Status = status
			writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"task": remoteTaskJSON(s.tasks[i].ID, s.tasks[i].ListID, s.tasks[i].Title, s.tasks[i].Status, i)}})
			return
		}
	}
	writeRemoteError(w, http.StatusNotFound, "task not found")
}

func nextRemoteID(n int) string {
	return strings.Repeat(string(rune('0'+n)), 8) + "-1111-4111-8111-111111111111"
}

func nextRemoteTaskID(n int) string {
	return strings.Repeat(string(rune('a'+n-1)), 8) + "-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
}

func remoteListJSON(id, name string, position int) map[string]any {
	return map[string]any{"id": id, "name": name, "position": position, "is_default": position == 0, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
}

func remoteTaskJSON(id, listID, title, status string, position int) map[string]any {
	return map[string]any{"id": id, "list_id": listID, "title": title, "status": status, "position": position, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
}

func writeRemoteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeRemoteError(w http.ResponseWriter, status int, message string) {
	writeRemoteJSON(w, status, map[string]any{"error": map[string]any{"code": "error", "message": message, "details": map[string]any{}}})
}
