package store

import (
	"fmt"
	"strings"
)

// Todo represents a single todo item.
type Todo struct {
	Text string
	Done bool
}

// Line represents a line in a markdown file — either a todo or a non-todo line.
type Line struct {
	Todo *Todo  // non-nil if this line is a todo
	Raw  string // original text for non-todo lines
}

// ParseMarkdown parses markdown content into a slice of Lines,
// preserving non-todo lines for round-trip fidelity.
func ParseMarkdown(content string) []Line {
	if content == "" {
		return nil
	}
	rawLines := strings.Split(content, "\n")
	// Remove trailing empty line from final newline
	if len(rawLines) > 0 && rawLines[len(rawLines)-1] == "" {
		rawLines = rawLines[:len(rawLines)-1]
	}

	lines := make([]Line, 0, len(rawLines))
	for _, raw := range rawLines {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "- [ ] ") {
			text := strings.TrimPrefix(trimmed, "- [ ] ")
			lines = append(lines, Line{Todo: &Todo{Text: text, Done: false}})
		} else if strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "- [X] ") {
			text := trimmed[6:] // skip "- [x] " or "- [X] "
			lines = append(lines, Line{Todo: &Todo{Text: text, Done: true}})
		} else {
			lines = append(lines, Line{Raw: raw})
		}
	}
	return lines
}

// RenderMarkdown converts Lines back to markdown content.
func RenderMarkdown(lines []Line) string {
	var b strings.Builder
	for _, l := range lines {
		if l.Todo != nil {
			if l.Todo.Done {
				fmt.Fprintf(&b, "- [x] %s\n", l.Todo.Text)
			} else {
				fmt.Fprintf(&b, "- [ ] %s\n", l.Todo.Text)
			}
		} else {
			b.WriteString(l.Raw)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// ExtractTodos returns only the todo items from parsed lines, with their 1-based indices
// relative to other todos (not line numbers).
func ExtractTodos(lines []Line) []Todo {
	var todos []Todo
	for _, l := range lines {
		if l.Todo != nil {
			todos = append(todos, *l.Todo)
		}
	}
	return todos
}

// todoIndex returns the line-slice index of the nth todo (1-based).
// Returns -1 if n is out of range.
func todoIndex(lines []Line, n int) int {
	count := 0
	for i, l := range lines {
		if l.Todo != nil {
			count++
			if count == n {
				return i
			}
		}
	}
	return -1
}
