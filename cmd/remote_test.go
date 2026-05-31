package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goodway/godos/config"
)

type remoteState struct {
	lists       []remoteList
	tasks       []remoteTask
	noteFolders []remoteNoteFolder
	notes       []remoteNote
	goals       []remoteGoal
	requests    []string
	listIDHit   int
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

type remoteNoteFolder struct {
	ID        string
	Name      string
	IsDefault bool
}

type remoteNote struct {
	ID       string
	FolderID string
	Title    string
	Body     string
	Pinned   bool
	Deleted  bool
}

type remoteGoal struct {
	ID          string
	Title       string
	Description string
	Reason      string
	Progress    int
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
	case r.Method == http.MethodGet && r.URL.Path == "/api/note-folders":
		folders := make([]map[string]any, 0, len(s.noteFolders))
		for i, folder := range s.noteFolders {
			folders = append(folders, remoteNoteFolderJSON(folder.ID, folder.Name, folder.IsDefault, i))
		}
		writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"note_folders": folders}})
	case r.Method == http.MethodGet && r.URL.Path == "/api/notes":
		query := r.URL.Query()
		notes := make([]map[string]any, 0, len(s.notes))
		for i, note := range s.notes {
			if folderID := query.Get("folder_id"); folderID != "" && note.FolderID != folderID {
				continue
			}
			if q := query.Get("q"); q != "" && !strings.Contains(strings.ToLower(note.Title), strings.ToLower(q)) {
				continue
			}
			if value := query.Get("pinned"); value != "" && ((value == "true") != note.Pinned) {
				continue
			}
			if value := query.Get("deleted"); value != "" && ((value == "true") != note.Deleted) {
				continue
			}
			notes = append(notes, remoteNoteJSON(note.ID, note.FolderID, note.Title, note.Body, note.Pinned, note.Deleted, i))
		}
		writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"notes": notes}})
	case r.Method == http.MethodPost && r.URL.Path == "/api/notes":
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		note := remoteNote{ID: nextRemoteNoteID(len(s.notes) + 1), FolderID: body["folder_id"].(string), Title: body["title"].(string)}
		if value, ok := body["body"].(string); ok {
			note.Body = value
		}
		s.notes = append(s.notes, note)
		writeRemoteJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{"note": remoteNoteJSON(note.ID, note.FolderID, note.Title, note.Body, note.Pinned, note.Deleted, len(s.notes)-1)}})
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/notes/"):
		s.writeNoteByID(w, strings.TrimPrefix(r.URL.Path, "/api/notes/"))
	case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/notes/"):
		id := strings.TrimPrefix(r.URL.Path, "/api/notes/")
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		for i := range s.notes {
			if s.notes[i].ID == id {
				if value, ok := body["body"].(string); ok {
					s.notes[i].Body = value
				}
				writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"note": remoteNoteJSON(s.notes[i].ID, s.notes[i].FolderID, s.notes[i].Title, s.notes[i].Body, s.notes[i].Pinned, s.notes[i].Deleted, i)}})
				return
			}
		}
		writeRemoteError(w, http.StatusNotFound, "note not found")
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/notes/"):
		id := strings.TrimPrefix(r.URL.Path, "/api/notes/")
		for i := range s.notes {
			if s.notes[i].ID == id {
				s.notes[i].Deleted = true
				writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"note": remoteNoteJSON(s.notes[i].ID, s.notes[i].FolderID, s.notes[i].Title, s.notes[i].Body, s.notes[i].Pinned, s.notes[i].Deleted, i)}})
				return
			}
		}
		writeRemoteError(w, http.StatusNotFound, "note not found")
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/restore"):
		s.updateNote(w, strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/notes/"), "/restore"), func(note *remoteNote) { note.Deleted = false })
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/pin"):
		s.updateNote(w, strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/notes/"), "/pin"), func(note *remoteNote) { note.Pinned = true })
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/unpin"):
		s.updateNote(w, strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/notes/"), "/unpin"), func(note *remoteNote) { note.Pinned = false })
	case r.Method == http.MethodGet && r.URL.Path == "/api/goals":
		goals := make([]map[string]any, 0, len(s.goals))
		for _, goal := range s.goals {
			goals = append(goals, remoteGoalJSON(goal.ID, goal.Title, goal.Description, goal.Reason, goal.Progress))
		}
		writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"goals": goals}})
	case r.Method == http.MethodPost && r.URL.Path == "/api/goals":
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		goal := remoteGoal{ID: nextRemoteGoalID(len(s.goals) + 1), Title: body["title"].(string)}
		goal.Description, _ = body["description"].(string)
		goal.Reason, _ = body["reason"].(string)
		s.goals = append(s.goals, goal)
		writeRemoteJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{"goal": remoteGoalJSON(goal.ID, goal.Title, goal.Description, goal.Reason, goal.Progress)}})
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/goals/"):
		s.writeGoalByID(w, strings.TrimPrefix(r.URL.Path, "/api/goals/"))
	case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/goals/"):
		s.updateGoalFields(w, strings.TrimPrefix(r.URL.Path, "/api/goals/"), r)
	case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/tasks/"):
		s.writeGoalByID(w, strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/goals/"), "/tasks/"+filepath.Base(r.URL.Path)))
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/tasks"):
		s.writeGoalByID(w, strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/goals/"), "/tasks"))
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/goals/"):
		id := strings.TrimPrefix(r.URL.Path, "/api/goals/")
		for i, goal := range s.goals {
			if goal.ID == id {
				s.goals = append(s.goals[:i], s.goals[i+1:]...)
				writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"goal": remoteGoalJSON(goal.ID, goal.Title, goal.Description, goal.Reason, goal.Progress)}})
				return
			}
		}
		writeRemoteError(w, http.StatusNotFound, "goal not found")
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

func TestNotesCommandListsRemoteNotesWithFilters(t *testing.T) {
	state := setupRemoteCommandTest(t)
	state.noteFolders = []remoteNoteFolder{{ID: nextRemoteNoteFolderID(1), Name: "Work", IsDefault: true}}
	state.notes = []remoteNote{{ID: nextRemoteNoteID(1), FolderID: nextRemoteNoteFolderID(1), Title: "Plan", Body: "body", Pinned: true, Deleted: true}}

	out, err := executeCommand(t, "notes", "--folder", "Work", "--query", "plan", "--pinned", "--deleted")
	if err != nil {
		t.Fatalf("notes failed: %v", err)
	}
	if !strings.Contains(out, "aaaaaaaa") || !strings.Contains(out, "Plan") {
		t.Fatalf("expected remote note in output, got %q", out)
	}
}

func TestNoteCommandsUseRemoteEditorAndLifecycle(t *testing.T) {
	state := setupRemoteCommandTest(t)
	state.noteFolders = []remoteNoteFolder{{ID: nextRemoteNoteFolderID(1), Name: "Notes", IsDefault: true}}
	state.notes = []remoteNote{{ID: nextRemoteNoteID(1), FolderID: nextRemoteNoteFolderID(1), Title: "Existing", Body: "old"}}
	setTestEditor(t, "edited body\n")

	if _, err := executeCommand(t, "note", "add", "Created"); err != nil {
		t.Fatalf("note add failed: %v", err)
	}
	if len(state.notes) != 2 || state.notes[1].Body != "edited body\n" {
		t.Fatalf("expected created note body from editor, notes: %#v", state.notes)
	}
	if out, err := executeCommand(t, "note", "show", "aaaaaaaa"); err != nil || !strings.Contains(out, "old") {
		t.Fatalf("note show = %q, %v", out, err)
	}
	if _, err := executeCommand(t, "note", "edit", "aaaaaaaa"); err != nil {
		t.Fatalf("note edit failed: %v", err)
	}
	if state.notes[0].Body != "edited body\n" {
		t.Fatalf("expected edited body, got %#v", state.notes[0])
	}
	if _, err := executeCommand(t, "note", "rm", "aaaaaaaa", "--force"); err != nil || !state.notes[0].Deleted {
		t.Fatalf("note rm failed: %v, note %#v", err, state.notes[0])
	}
	if _, err := executeCommand(t, "note", "restore", "aaaaaaaa"); err != nil || state.notes[0].Deleted {
		t.Fatalf("note restore failed: %v, note %#v", err, state.notes[0])
	}
	if _, err := executeCommand(t, "note", "pin", "aaaaaaaa"); err != nil || !state.notes[0].Pinned {
		t.Fatalf("note pin failed: %v, note %#v", err, state.notes[0])
	}
	if _, err := executeCommand(t, "note", "unpin", "aaaaaaaa"); err != nil || state.notes[0].Pinned {
		t.Fatalf("note unpin failed: %v, note %#v", err, state.notes[0])
	}
}

func TestNoteRemoveCancelledDoesNotDelete(t *testing.T) {
	state := setupRemoteCommandTest(t)
	state.notes = []remoteNote{{ID: nextRemoteNoteID(1), FolderID: nextRemoteNoteFolderID(1), Title: "Existing", Body: "old"}}

	out, err := executeCommandWithInput(t, strings.NewReader("n\n"), "note", "rm", "aaaaaaaa")
	if err != nil {
		t.Fatalf("note rm failed: %v", err)
	}
	if state.notes[0].Deleted || !strings.Contains(out, "Cancelled") {
		t.Fatalf("expected cancelled delete, output %q note %#v", out, state.notes[0])
	}
}

func TestGoalCommandsUseRemoteGoals(t *testing.T) {
	state := setupRemoteCommandTest(t)
	state.goals = []remoteGoal{{ID: nextRemoteGoalID(1), Title: "Launch", Description: "Ship", Reason: "Learn", Progress: 25}}
	state.tasks = []remoteTask{{ID: nextRemoteTaskID(1), ListID: nextRemoteID(1), Title: "task", Status: "active"}}

	if out, err := executeCommand(t, "goals"); err != nil || !strings.Contains(out, "aaaaaaaa") || !strings.Contains(out, "25%") {
		t.Fatalf("goals = %q, %v", out, err)
	}
	if _, err := executeCommand(t, "goal", "add", "New", "--description", "Desc", "--reason", "Why"); err != nil {
		t.Fatalf("goal add failed: %v", err)
	}
	if len(state.goals) != 2 || state.goals[1].Description != "Desc" || state.goals[1].Reason != "Why" {
		t.Fatalf("unexpected goals after add: %#v", state.goals)
	}
	if out, err := executeCommand(t, "goal", "show", "aaaaaaaa"); err != nil || !strings.Contains(out, "Launch") || !strings.Contains(out, "Ship") {
		t.Fatalf("goal show = %q, %v", out, err)
	}
	if _, err := executeCommand(t, "goal", "edit", "aaaaaaaa", "--title", "Updated", "--description", "Updated desc", "--reason", "Updated reason"); err != nil {
		t.Fatalf("goal edit failed: %v", err)
	}
	if state.goals[0].Title != "Updated" || state.goals[0].Description != "Updated desc" || state.goals[0].Reason != "Updated reason" {
		t.Fatalf("unexpected goal after edit: %#v", state.goals[0])
	}
	if _, err := executeCommand(t, "goal", "link", "aaaaaaaa", "aaaaaaaa"); err != nil {
		t.Fatalf("goal link failed: %v", err)
	}
	if _, err := executeCommand(t, "goal", "unlink", "aaaaaaaa", "aaaaaaaa"); err != nil {
		t.Fatalf("goal unlink failed: %v", err)
	}
	if _, err := executeCommand(t, "goal", "rm", "aaaaaaaa", "--force"); err != nil {
		t.Fatalf("goal rm failed: %v", err)
	}
	if len(state.goals) != 1 {
		t.Fatalf("expected goal to be removed, got %#v", state.goals)
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

func (s *remoteState) writeNoteByID(w http.ResponseWriter, id string) {
	for i, note := range s.notes {
		if note.ID == id {
			writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"note": remoteNoteJSON(note.ID, note.FolderID, note.Title, note.Body, note.Pinned, note.Deleted, i)}})
			return
		}
	}
	writeRemoteError(w, http.StatusNotFound, "note not found")
}

func (s *remoteState) updateNote(w http.ResponseWriter, id string, update func(*remoteNote)) {
	for i := range s.notes {
		if s.notes[i].ID == id {
			update(&s.notes[i])
			writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"note": remoteNoteJSON(s.notes[i].ID, s.notes[i].FolderID, s.notes[i].Title, s.notes[i].Body, s.notes[i].Pinned, s.notes[i].Deleted, i)}})
			return
		}
	}
	writeRemoteError(w, http.StatusNotFound, "note not found")
}

func (s *remoteState) writeGoalByID(w http.ResponseWriter, id string) {
	for _, goal := range s.goals {
		if goal.ID == id {
			writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"goal": remoteGoalJSON(goal.ID, goal.Title, goal.Description, goal.Reason, goal.Progress)}})
			return
		}
	}
	writeRemoteError(w, http.StatusNotFound, "goal not found")
}

func (s *remoteState) updateGoalFields(w http.ResponseWriter, id string, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	for i := range s.goals {
		if s.goals[i].ID == id {
			if value, ok := body["title"].(string); ok {
				s.goals[i].Title = value
			}
			if value, ok := body["description"].(string); ok {
				s.goals[i].Description = value
			}
			if value, ok := body["reason"].(string); ok {
				s.goals[i].Reason = value
			}
			writeRemoteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"goal": remoteGoalJSON(s.goals[i].ID, s.goals[i].Title, s.goals[i].Description, s.goals[i].Reason, s.goals[i].Progress)}})
			return
		}
	}
	writeRemoteError(w, http.StatusNotFound, "goal not found")
}

func TestNoteAddCreatesAtomicallyWithBody(t *testing.T) {
	state := setupRemoteCommandTest(t)
	state.noteFolders = []remoteNoteFolder{{ID: nextRemoteNoteFolderID(1), Name: "Notes", IsDefault: true}}
	setTestEditor(t, "edited body\n")

	if _, err := executeCommand(t, "note", "add", "Created"); err != nil {
		t.Fatalf("note add failed: %v", err)
	}
	if len(state.notes) != 1 || state.notes[0].Body != "edited body\n" {
		t.Fatalf("expected created note with body, notes: %#v", state.notes)
	}
	creates := 0
	for _, req := range state.requests {
		if req == "POST /api/notes?" {
			creates++
		}
		if strings.HasPrefix(req, "PATCH /api/notes/") {
			t.Fatalf("expected no follow-up PATCH, requests: %#v", state.requests)
		}
	}
	if creates != 1 {
		t.Fatalf("expected exactly one create call, requests: %#v", state.requests)
	}
}

func TestNoteAddEditorFailureCreatesNoNote(t *testing.T) {
	state := setupRemoteCommandTest(t)
	state.noteFolders = []remoteNoteFolder{{ID: nextRemoteNoteFolderID(1), Name: "Notes", IsDefault: true}}
	setFailingEditor(t)

	if _, err := executeCommand(t, "note", "add", "Created"); err == nil {
		t.Fatal("expected note add to fail when editor fails")
	}
	if len(state.notes) != 0 {
		t.Fatalf("expected no note created on editor failure, notes: %#v", state.notes)
	}
	for _, req := range state.requests {
		if strings.HasPrefix(req, "POST /api/notes") {
			t.Fatalf("expected no create call on editor failure, requests: %#v", state.requests)
		}
	}
}

func TestGoalEditWithoutFlagsReturnsErrorAndSendsNoPatch(t *testing.T) {
	state := setupRemoteCommandTest(t)
	state.goals = []remoteGoal{{ID: nextRemoteGoalID(1), Title: "Launch", Progress: 0}}

	if _, err := executeCommand(t, "goal", "edit", "aaaaaaaa"); err == nil || !strings.Contains(err.Error(), "no changes specified") {
		t.Fatalf("expected no-changes error, got %v", err)
	}
	for _, req := range state.requests {
		if strings.HasPrefix(req, "PATCH /api/goals/") {
			t.Fatalf("expected no PATCH when no flags set, requests: %#v", state.requests)
		}
	}
}

func setTestEditor(t *testing.T, body string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "editor.sh")
	content := "#!/bin/sh\nprintf '%s' '" + strings.ReplaceAll(body, "'", "'\\''") + "' > \"$1\"\n"
	if err := os.WriteFile(path, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", path)
}

func setFailingEditor(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", path)
}

func nextRemoteID(n int) string {
	return strings.Repeat(string(rune('0'+n)), 8) + "-1111-4111-8111-111111111111"
}

func nextRemoteTaskID(n int) string {
	return strings.Repeat(string(rune('a'+n-1)), 8) + "-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
}

func nextRemoteNoteFolderID(n int) string {
	return strings.Repeat(string(rune('0'+n)), 8) + "-2222-4222-8222-222222222222"
}

func nextRemoteNoteID(n int) string {
	return strings.Repeat(string(rune('a'+n-1)), 8) + "-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
}

func nextRemoteGoalID(n int) string {
	return strings.Repeat(string(rune('a'+n-1)), 8) + "-cccc-4ccc-8ccc-cccccccccccc"
}

func remoteListJSON(id, name string, position int) map[string]any {
	return map[string]any{"id": id, "name": name, "position": position, "is_default": position == 0, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
}

func remoteTaskJSON(id, listID, title, status string, position int) map[string]any {
	return map[string]any{"id": id, "list_id": listID, "title": title, "status": status, "position": position, "inserted_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
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

func writeRemoteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeRemoteError(w http.ResponseWriter, status int, message string) {
	writeRemoteJSON(w, status, map[string]any{"error": map[string]any{"code": "error", "message": message, "details": map[string]any{}}})
}
