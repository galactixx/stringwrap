package stringwrap

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stringWrapTestCase struct {
	input          string
	wrapped        string
	limit          int
	trimWhitespace bool
	splitWord      bool
}

func wrapString(tt stringWrapTestCase) (string, *WrappedStringSeq, error) {
	if tt.splitWord {
		return StringWrapSplit(tt.input, tt.limit, 4, tt.trimWhitespace)
	} else {
		return StringWrap(tt.input, tt.limit, 4, tt.trimWhitespace)
	}
}

func TestWrappingStrings(t *testing.T) {
	tests := []stringWrapTestCase{
		{
			input:          "The quick brown fox jumps over the lazy dog",
			wrapped:        "The quick\nbrown fox\njumps over\nthe lazy\ndog",
			limit:          10,
			trimWhitespace: true,
			splitWord:      false,
		},
		{
			input:          "Supercalifragilisticexpialidocious",
			wrapped:        "Supercali-\nfragilist-\nicexpiali-\ndocious",
			limit:          10,
			trimWhitespace: true,
			splitWord:      true,
		},
		{
			input:          "hello\tworld",
			wrapped:        "hello   world",
			limit:          15,
			trimWhitespace: true,
			splitWord:      false,
		},
		{
			input:          "hello\tworld",
			wrapped:        "hello\nworld",
			limit:          7,
			trimWhitespace: true,
			splitWord:      false,
		},
		{
			input:          "Pseudopseudohypoparathyroidism is a long medical term that might be split",
			wrapped:        "Pseudopseudohy-\npoparathyroidi-\nsm is a long m-\nedical term th-\nat might be sp-\nlit",
			limit:          15,
			trimWhitespace: true,
			splitWord:      true,
		},
		{
			input:          "\x1b[32m\tGreen 🍀 text with ANSI and emojis\x1b[0m alongside  plain content here",
			wrapped:        "\x1b[32m    Green 🍀 text \nwith ANSI and \nemojis\x1b[0m alongside  \nplain content here",
			limit:          18,
			trimWhitespace: false,
			splitWord:      false,
		},
		{
			input:          "\x1b[31mred\x1b[0m text normal",
			wrapped:        "\x1b[31mred\x1b[0m text\nnormal",
			limit:          10,
			trimWhitespace: true,
			splitWord:      false,
		},
		{
			input:          "Hello.",
			wrapped:        "Hell-\no.",
			limit:          5,
			trimWhitespace: true,
			splitWord:      true,
		},
		{
			input:          "\tThis is a longer example input that will wrap nicely  ",
			wrapped:        "    This is a longer\n example input that \nwill wrap nicely  ",
			limit:          20,
			trimWhitespace: false,
			splitWord:      false,
		},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("Wrapped String Test %d", idx+1), func(t *testing.T) {
			wrapped, seq, err := wrapString(tt)
			assert.Nil(t, err)
			assert.Equal(t, len(seq.WrappedLines), len(strings.Split(wrapped, "\n")))
			assert.Equal(t, tt.wrapped, wrapped)
		})
	}
}

func TestWrappedStringSeq(t *testing.T) {
	input := "Hello world!\nLine two with 🌟stars\nFinal"
	limit := 8
	tabSize := 4

	wrapped, seq, _ := StringWrap(input, limit, tabSize, true)
	assert.Equal(t, "Hello\nworld!\nLine two\nwith\n🌟stars\nFinal", wrapped)

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
			Width:             5,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        2,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 6, End: 13},
			OrigRuneOffset:    LineOffset{Start: 6, End: 13},
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
			OrigByteOffset:    LineOffset{Start: 13, End: 21},
			OrigRuneOffset:    LineOffset{Start: 13, End: 21},
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
			OrigByteOffset:    LineOffset{Start: 21, End: 27},
			OrigRuneOffset:    LineOffset{Start: 21, End: 27},
			SegmentInOrig:     2,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             4,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        5,
			OrigLineNum:       2,
			OrigByteOffset:    LineOffset{Start: 27, End: 37},
			OrigRuneOffset:    LineOffset{Start: 27, End: 34},
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
			OrigByteOffset:    LineOffset{Start: 37, End: 42},
			OrigRuneOffset:    LineOffset{Start: 34, End: 39},
			SegmentInOrig:     1,
			LastSegmentInOrig: true,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             5,
			EndsWithSplitWord: false,
		},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("Wrapped String Seq Test %d", idx+1), func(t *testing.T) {
			wrappedLine := seq.WrappedLines[idx]
			assert.Equal(t, tt, wrappedLine)
		})
	}
}

func TestWrappedStringSplitSeq(t *testing.T) {
	input := "Supercalifragilisticexpialidocious is a long word often used to test wrapping behavior."
	limit := 10
	tabSize := 4

	wrapped, seq, _ := StringWrapSplit(input, limit, tabSize, true)
	assert.Equal(
		t,
		"Supercali-\nfragilist-\nicexpiali-\ndocious is\na long wo-\nrd often\nused to t-\nest wrapp-\ning behav-\nior.",
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
			OrigByteOffset:    LineOffset{Start: 51, End: 61},
			OrigRuneOffset:    LineOffset{Start: 51, End: 61},
			SegmentInOrig:     6,
			LastSegmentInOrig: false,
			NotWithinLimit:    false,
			IsHardBreak:       false,
			Width:             8,
			EndsWithSplitWord: false,
		},
		{
			CurLineNum:        7,
			OrigLineNum:       1,
			OrigByteOffset:    LineOffset{Start: 61, End: 71},
			OrigRuneOffset:    LineOffset{Start: 61, End: 71},
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
			OrigByteOffset:    LineOffset{Start: 71, End: 81},
			OrigRuneOffset:    LineOffset{Start: 71, End: 81},
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
			OrigByteOffset:    LineOffset{Start: 81, End: 91},
			OrigRuneOffset:    LineOffset{Start: 81, End: 91},
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
			OrigByteOffset:    LineOffset{Start: 91, End: 95},
			OrigRuneOffset:    LineOffset{Start: 91, End: 95},
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
