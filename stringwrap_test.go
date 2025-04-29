package stringwrap

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
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

	lines := strings.Split(wrapped, "\n")
	for _, line := range lines {
		if runewidth.StringWidth(line) > limit {
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

	lines := strings.Split(wrapped, "\n")
	for _, line := range lines {
		if runewidth.StringWidth(line) > limit {
			t.Errorf("line exceeds limit: %q", line)
		}
	}

	if len(seq.WrappedLines) != 4 {
		t.Errorf("expected the long word to be split across four lines")
	}

	if !seq.WrappedLines[0].EndsWithSplitWord ||
		!seq.WrappedLines[1].EndsWithSplitWord ||
		!seq.WrappedLines[2].EndsWithSplitWord {
		t.Errorf("expected the first three lines to end with a split word")
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
	if wrapped != "hello   world" {
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

	if wrapped != "\x1b[31mred\x1b[0m text \nnormal" {
		t.Errorf("expected ANSI codes to be preserved")
	}
}
