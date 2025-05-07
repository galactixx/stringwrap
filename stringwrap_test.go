package stringwrap

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/stretchr/testify/assert"
)

func TestStringWrapBasic(t *testing.T) {
	input := "The quick brown fox jumps over the lazy dog"
	limit := 10
	tabSize := 4

	wrapped, seq, err := StringWrap(input, limit, tabSize)
	assert.Nil(t, err)
	assert.NotEqual(t, "", wrapped)

	lines := strings.Split(wrapped, "\n")
	for _, line := range lines {
		assert.LessOrEqual(t, runewidth.StringWidth(line), limit)
	}

	assert.Equal(t, len(lines), len(seq.WrappedLines))
}

func TestStringWrapSplitLongWord(t *testing.T) {
	input := "Supercalifragilisticexpialidocious"
	limit := 10
	tabSize := 4

	wrapped, seq, err := StringWrapSplit(input, limit, tabSize)
	assert.Nil(t, err)
	assert.NotEqual(t, "", wrapped)

	lines := strings.Split(wrapped, "\n")
	for _, line := range lines {
		assert.LessOrEqual(t, runewidth.StringWidth(line), limit)
	}

	assert.Equal(t, len(lines), len(seq.WrappedLines))
	assert.True(t, (seq.WrappedLines[0].EndsWithSplitWord &&
		seq.WrappedLines[1].EndsWithSplitWord &&
		seq.WrappedLines[2].EndsWithSplitWord),
	)
}

func TestStringWrapErrorOnSmallLimit(t *testing.T) {
	input := "hello"
	limit := 1
	tabSize := 4

	_, _, err := StringWrap(input, limit, tabSize)
	assert.NotNil(t, err)
}

func TestStringWrapTabHandling(t *testing.T) {
	input := "hello\tworld"
	limit := 15
	tabSize := 4

	wrapped, _, err := StringWrap(input, limit, tabSize)
	assert.Nil(t, err)

	// tab should be expanded into spaces
	assert.Equal(t, "hello   world", wrapped)
}

func TestStringWrapTabSplit(t *testing.T) {
	input := "hello\tworld"
	limit := 7
	tabSize := 4

	wrapped, _, err := StringWrap(input, limit, tabSize)
	assert.Nil(t, err)

	fmt.Printf("%q", wrapped)
	// tab should not be split across lines
	assert.Equal(t, "hello\n    \nworld", wrapped)
}

func TestStringWrapANSIHandling(t *testing.T) {
	input := "\x1b[31mred\x1b[0m text normal"
	limit := 10
	tabSize := 4

	wrapped, _, err := StringWrap(input, limit, tabSize)
	assert.Nil(t, err)
	assert.Equal(t, "\x1b[31mred\x1b[0m text \nnormal", wrapped)
}

func TestStringWrapSplitNearPunc(t *testing.T) {
	input := "Hello."
	limit := 5
	tabSize := 4

	wrapped, _, err := StringWrapSplit(input, limit, tabSize)
	assert.Nil(t, err)
	assert.Equal(t, "Hell-\no.", wrapped)
}

func TestStringWrapHardBreak(t *testing.T) {
	input := "Hello.\nThis is \ntesting explicit\n new lines."
	limit := 30
	tabSize := 4

	wrapped, _, err := StringWrap(input, limit, tabSize)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(strings.Split(wrapped, "\n")))
}

func TestWrappedStringSeq(t *testing.T) {
	input := "Hello world!\nLine two with 🌟stars\nFinal"
	limit := 8
	tabSize := 4

	wrapped, seq, _ := StringWrap(input, limit, tabSize)
	assert.Equal(t, "Hello \nworld!\nLine two\nwith \n🌟stars\nFinal", wrapped)

	lines := strings.Split(wrapped, "\n")
	assert.Equal(t, len(lines), len(seq.WrappedLines))
	tests := []WrappedString{
		{
			CurLineNum:        1,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 0, End: 6},
			OrigRuneOffset:    LineOffset{Start: 0, End: 6},
			SegmentInOrig:     1,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             6,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        2,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 6, End: 12},
			OrigRuneOffset:    LineOffset{Start: 6, End: 12},
			SegmentInOrig:     2,
			LastSegmentInOrig: true,
			NotWithinLimit:    false,
			IsHardBreak:       true,
			Width:             6,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        3,
			OrigLineNum:       2,
			OrigByteOffset:    LineOffset{Start: 12, End: 20},
			OrigRuneOffset:    LineOffset{Start: 12, End: 20},
			SegmentInOrig:     1,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             8,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        4,
			OrigLineNum:       2,
			OrigByteOffset:    LineOffset{Start: 20, End: 26},
			OrigRuneOffset:    LineOffset{Start: 20, End: 26},
			SegmentInOrig:     2,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             5,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        5,
			OrigLineNum:       2,
			OrigByteOffset:    LineOffset{Start: 26, End: 35},
			OrigRuneOffset:    LineOffset{Start: 26, End: 32},
			SegmentInOrig:     3,
			LastSegmentInOrig: true,
			NotWithinLimit:    false,
			IsHardBreak:       true,
			Width:             7,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        6,
			OrigLineNum:       3,
			OrigByteOffset:    LineOffset{Start: 35, End: 40},
			OrigRuneOffset:    LineOffset{Start: 32, End: 37},
			SegmentInOrig:     1,
			LastSegmentInOrig: true,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             5,
			EndsWithSplitWord: false,
		},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("Wrapped String Test %d", idx+1), func(t *testing.T) {
			wrappedLine := seq.WrappedLines[idx]
			assert.Equal(t, tt, wrappedLine)
		})
	}
}

func TestWrappedStringSplitSeq(t *testing.T) {
	input := "Supercalifragilisticexpialidocious is a long word often used to test wrapping behavior."
	limit := 10
	tabSize := 4

	wrapped, seq, _ := StringWrapSplit(input, limit, tabSize)
	assert.Equal(
		t,
		"Supercali-\nfragilist-\nicexpiali-\ndocious is\na long wo-\nrd often \nused to t-\nest wrapp-\ning behav-\nior.",
		wrapped,
	)

	lines := strings.Split(wrapped, "\n")
	assert.Equal(t, len(lines), len(seq.WrappedLines))
	tests := []WrappedString{
		{
			CurLineNum:        1,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 0, End: 10},
			OrigRuneOffset:    LineOffset{Start: 0, End: 10},
			SegmentInOrig:     1,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             10,
			EndsWithSplitWord: true,
		},
		{
			CurLineNum:        2,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 10, End: 20},
			OrigRuneOffset:    LineOffset{Start: 10, End: 20},
			SegmentInOrig:     2,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             10,
			EndsWithSplitWord: true,
		},
		{
			CurLineNum:        3,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 20, End: 30},
			OrigRuneOffset:    LineOffset{Start: 20, End: 30},
			SegmentInOrig:     3,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             10,
			EndsWithSplitWord: true,
		},
		{
			CurLineNum:        4,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 30, End: 40},
			OrigRuneOffset:    LineOffset{Start: 30, End: 40},
			SegmentInOrig:     4,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             10,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        5,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 40, End: 51},
			OrigRuneOffset:    LineOffset{Start: 40, End: 51},
			SegmentInOrig:     5,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             10,
			EndsWithSplitWord: true,
		},
		{
			CurLineNum:        6,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 51, End: 60},
			OrigRuneOffset:    LineOffset{Start: 51, End: 60},
			SegmentInOrig:     6,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             10,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        7,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 60, End: 70},
			OrigRuneOffset:    LineOffset{Start: 60, End: 70},
			SegmentInOrig:     7,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             10,
			EndsWithSplitWord: true,
		},
		{
			CurLineNum:        8,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 70, End: 80},
			OrigRuneOffset:    LineOffset{Start: 70, End: 80},
			SegmentInOrig:     8,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             10,
			EndsWithSplitWord: true,
		},
		{
			CurLineNum:        9,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 80, End: 90},
			OrigRuneOffset:    LineOffset{Start: 80, End: 90},
			SegmentInOrig:     9,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             10,
			EndsWithSplitWord: true,
		},
		{
			CurLineNum:        10,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 90, End: 94},
			OrigRuneOffset:    LineOffset{Start: 90, End: 94},
			SegmentInOrig:     10,
			LastSegmentInOrig: true,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             4,
			EndsWithSplitWord: false,
		},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("Wrapped String Test %d", idx+1), func(t *testing.T) {
			wrappedLine := seq.WrappedLines[idx]
			assert.Equal(t, tt, wrappedLine)
		})
	}
}
