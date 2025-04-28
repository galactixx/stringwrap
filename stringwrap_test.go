package stringwrap

import (
	"strings"
	"testing"
)

func TestStringWrapBasic(t *testing.T) {
	input := "The quick brown fox jumps over the lazy dog"
	limit := 10
	tabSize := 4

	wrapped, seq, err := StringWrap(input, limit, tabSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wrapped == "" {
		t.Fatalf("expected wrapped string to be non-empty")
	}

	lines := strings.Split(strings.TrimSpace(wrapped), "\n")
	for _, line := range lines {
		if len(line) > limit {
			t.Errorf("line exceeds limit: %q", line)
		}
	}

	if len(seq.WrappedLines) != len(lines) {
		t.Errorf("expected %d wrapped lines, got %d", len(lines), len(seq.WrappedLines))
	}
}

func TestStringWrapSplitLongWord(t *testing.T) {
	input := "Supercalifragilisticexpialidocious"
	limit := 10
	tabSize := 4

	wrapped, seq, err := StringWrapSplit(input, limit, tabSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wrapped == "" {
		t.Fatalf("expected wrapped string to be non-empty")
	}

	lines := strings.Split(strings.TrimSpace(wrapped), "\n")
	// +1 for optional hyphen
	for _, line := range lines {
		if len(line) > limit+1 {
			t.Errorf("line exceeds limit: %q", line)
		}
	}

	if len(seq.WrappedLines) < 2 {
		t.Errorf("expected the long word to be split across multiple lines")
	}

	foundSplit := false
	for _, ws := range seq.WrappedLines {
		if ws.EndsWithSplitWord {
			foundSplit = true
			break
		}
	}
	if !foundSplit {
		t.Errorf("expected at least one line to end with a split word")
	}
}

func TestStringWrapErrorOnSmallLimit(t *testing.T) {
	input := "hello"
	limit := 1
	tabSize := 4

	_, _, err := StringWrap(input, limit, tabSize)
	if err == nil {
		t.Fatalf("expected an error when limit < 2, but got nil")
	}
}

func TestStringWrapTabHandling(t *testing.T) {
	input := "hello\tworld"
	limit := 15
	tabSize := 4

	wrapped, _, err := StringWrap(input, limit, tabSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// tab should be expanded into spaces
	if !strings.Contains(wrapped, "   ") {
		t.Errorf("expected tab character to be expanded into spaces")
	}
}

func TestStringWrapANSIHandling(t *testing.T) {
	input := "\x1b[31mred\x1b[0m text normal"
	limit := 10
	tabSize := 4

	wrapped, _, err := StringWrap(input, limit, tabSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still contain the ANSI escape sequences
	if !strings.Contains(wrapped, "\x1b[31m") || !strings.Contains(wrapped, "\x1b[0m") {
		t.Errorf("expected ANSI codes to be preserved")
	}
}
