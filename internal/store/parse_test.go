package store

import (
	"testing"
)

func TestParseMarkdown_Empty(t *testing.T) {
	lines := ParseMarkdown("")
	if lines != nil {
		t.Fatalf("expected nil, got %v", lines)
	}
}

func TestParseMarkdown_IncompleteTodo(t *testing.T) {
	lines := ParseMarkdown("- [ ] buy milk\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Todo == nil {
		t.Fatal("expected todo, got raw line")
	}
	if lines[0].Todo.Text != "buy milk" {
		t.Errorf("expected text %q, got %q", "buy milk", lines[0].Todo.Text)
	}
	if lines[0].Todo.Done {
		t.Error("expected todo to be incomplete")
	}
}

func TestParseMarkdown_CompleteTodo(t *testing.T) {
	lines := ParseMarkdown("- [x] buy milk\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Todo == nil {
		t.Fatal("expected todo, got raw line")
	}
	if !lines[0].Todo.Done {
		t.Error("expected todo to be complete")
	}
}

func TestParseMarkdown_UppercaseX(t *testing.T) {
	lines := ParseMarkdown("- [X] buy milk\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Todo == nil {
		t.Fatal("expected todo, got raw line")
	}
	if !lines[0].Todo.Done {
		t.Error("expected uppercase [X] to be treated as complete")
	}
	if lines[0].Todo.Text != "buy milk" {
		t.Errorf("expected text %q, got %q", "buy milk", lines[0].Todo.Text)
	}
}

func TestParseMarkdown_MixedContent(t *testing.T) {
	input := "# My Todos\n- [ ] first\n- [x] second\nsome text\n- [ ] third\n"
	lines := ParseMarkdown(input)
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	// Line 0: heading (raw)
	if lines[0].Todo != nil {
		t.Error("expected line 0 to be raw")
	}
	if lines[0].Raw != "# My Todos" {
		t.Errorf("expected raw %q, got %q", "# My Todos", lines[0].Raw)
	}
	// Line 1: incomplete todo
	if lines[1].Todo == nil || lines[1].Todo.Done {
		t.Error("expected line 1 to be incomplete todo")
	}
	// Line 2: complete todo
	if lines[2].Todo == nil || !lines[2].Todo.Done {
		t.Error("expected line 2 to be complete todo")
	}
	// Line 3: raw
	if lines[3].Todo != nil {
		t.Error("expected line 3 to be raw")
	}
	// Line 4: incomplete todo
	if lines[4].Todo == nil || lines[4].Todo.Done {
		t.Error("expected line 4 to be incomplete todo")
	}
}

func TestRenderMarkdown_RoundTrip(t *testing.T) {
	input := "# My Todos\n- [ ] first\n- [x] second\nsome text\n"
	lines := ParseMarkdown(input)
	output := RenderMarkdown(lines)
	if output != input {
		t.Errorf("round-trip failed:\ninput:  %q\noutput: %q", input, output)
	}
}

func TestRenderMarkdown_Empty(t *testing.T) {
	output := RenderMarkdown(nil)
	if output != "" {
		t.Errorf("expected empty string, got %q", output)
	}
}

func TestExtractTodos(t *testing.T) {
	lines := ParseMarkdown("# Title\n- [ ] a\ntext\n- [x] b\n- [ ] c\n")
	todos := ExtractTodos(lines)
	if len(todos) != 3 {
		t.Fatalf("expected 3 todos, got %d", len(todos))
	}
	if todos[0].Text != "a" || todos[0].Done {
		t.Errorf("todo 0: want {a, false}, got {%s, %v}", todos[0].Text, todos[0].Done)
	}
	if todos[1].Text != "b" || !todos[1].Done {
		t.Errorf("todo 1: want {b, true}, got {%s, %v}", todos[1].Text, todos[1].Done)
	}
	if todos[2].Text != "c" || todos[2].Done {
		t.Errorf("todo 2: want {c, false}, got {%s, %v}", todos[2].Text, todos[2].Done)
	}
}

func TestTodoIndex(t *testing.T) {
	lines := ParseMarkdown("# Title\n- [ ] a\ntext\n- [x] b\n- [ ] c\n")

	tests := []struct {
		n    int
		want int
	}{
		{1, 1},  // first todo is at line index 1
		{2, 3},  // second todo is at line index 3
		{3, 4},  // third todo is at line index 4
		{0, -1}, // out of range
		{4, -1}, // out of range
	}
	for _, tt := range tests {
		got := todoIndex(lines, tt.n)
		if got != tt.want {
			t.Errorf("todoIndex(lines, %d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestParseMarkdown_BareCheckbox(t *testing.T) {
	tests := []struct {
		input string
		done  bool
	}{
		{"- [ ]\n", false},
		{"- [x]\n", true},
		{"- [X]\n", true},
	}
	for _, tt := range tests {
		lines := ParseMarkdown(tt.input)
		if len(lines) != 1 {
			t.Fatalf("input %q: expected 1 line, got %d", tt.input, len(lines))
		}
		if lines[0].Todo == nil {
			t.Errorf("input %q: expected todo, got raw line", tt.input)
			continue
		}
		if lines[0].Todo.Text != "" {
			t.Errorf("input %q: expected empty text, got %q", tt.input, lines[0].Todo.Text)
		}
		if lines[0].Todo.Done != tt.done {
			t.Errorf("input %q: expected done=%v, got %v", tt.input, tt.done, lines[0].Todo.Done)
		}
	}
}
