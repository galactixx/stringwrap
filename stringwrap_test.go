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
	assert.NotEqual(t, wrapped, "")

	lines := strings.Split(wrapped, "\n")
	for _, line := range lines {
		assert.LessOrEqual(t, runewidth.StringWidth(line), limit)
	}

	assert.Equal(t, len(seq.WrappedLines), len(lines))
}

func TestStringWrapSplitLongWord(t *testing.T) {
	input := "Supercalifragilisticexpialidocious"
	limit := 10
	tabSize := 4

	wrapped, seq, err := StringWrapSplit(input, limit, tabSize)
	assert.Nil(t, err)
	assert.NotEqual(t, wrapped, "")

	lines := strings.Split(wrapped, "\n")
	for _, line := range lines {
		assert.LessOrEqual(t, runewidth.StringWidth(line), limit)
	}

	assert.Equal(t, len(seq.WrappedLines), 4)
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
	assert.Equal(t, wrapped, "hello   world")
}

func TestStringWrapANSIHandling(t *testing.T) {
	input := "\x1b[31mred\x1b[0m text normal"
	limit := 10
	tabSize := 4

	wrapped, _, err := StringWrap(input, limit, tabSize)
	assert.Nil(t, err)
	assert.Equal(t, wrapped, "\x1b[31mred\x1b[0m text \nnormal")
}

func TestWrappedStringSeq(t *testing.T) {
	input := "Hello world!\nLine two with 🌟stars\nFinal"
	limit := 8
	tabSize := 4

	wrapped, seq, _ := StringWrap(input, limit, tabSize)
	assert.Equal(t, wrapped, "Hello \nworld!\nLine two\nwith \n🌟stars\nFinal")

	lines := strings.Split(wrapped, "\n")
	assert.Equal(t, len(seq.WrappedLines), len(lines))
	tests := []WrappedString{
		{
			CurLineNum:        1,
			OrigLineNum:       1,
			SegmentInOrig:     1,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             6,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        2,
			OrigLineNum:       1,
			SegmentInOrig:     2,
			NotWithinLimit:    false,
			IsHardBreak:       true,
			Width:             6,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        3,
			OrigLineNum:       2,
			SegmentInOrig:     1,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             8,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        4,
			OrigLineNum:       2,
			SegmentInOrig:     2,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             6, // need to look into this number
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        5,
			OrigLineNum:       2,
			SegmentInOrig:     3,
			NotWithinLimit:    false,
			IsHardBreak:       true,
			Width:             7,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        6,
			OrigLineNum:       3,
			SegmentInOrig:     1,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             5,
			EndsWithSplitWord: false,
		},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("Wrapped String Test %d", idx+1), func(t *testing.T) {
			wrappedLine := seq.WrappedLines[idx]
			assert.Equal(t, tt.CurLineNum, wrappedLine.CurLineNum)
			assert.Equal(t, tt.OrigLineNum, wrappedLine.OrigLineNum)
			assert.Equal(t, tt.SegmentInOrig, wrappedLine.SegmentInOrig)
			assert.Equal(t, tt.NotWithinLimit, wrappedLine.NotWithinLimit)
			assert.Equal(t, tt.IsHardBreak, wrappedLine.IsHardBreak)
			assert.Equal(t, tt.EndsWithSplitWord, wrappedLine.EndsWithSplitWord)
			assert.Equal(t, tt.Width, wrappedLine.Width)
		})
	}
}
