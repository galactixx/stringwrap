package stringwrap

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// stringWrapTestCase is a struct that contains the input string, the expected
// wrapped string, the limit, the trimWhitespace flag, and the splitWord flag.
type stringWrapTestCase struct {
	input          string
	wrapped        string
	limit          int
	trimWhitespace bool
	splitWord      bool
}

// wrapString is a helper function that wraps a string using the StringWrap
// or StringWrapSplit function based on the splitWord flag.
func wrapString(tt stringWrapTestCase) (string, *WrappedStringSeq, error) {
	if tt.splitWord {
		return StringWrapSplit(tt.input, tt.limit, 4, tt.trimWhitespace)
	} else {
		return StringWrap(tt.input, tt.limit, 4, tt.trimWhitespace)
	}
}

// TestStringWrap tests the StringWrap function with a variety of test cases.
func TestStringWrap(t *testing.T) {
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
		{
			input:          "e\u0301clair",
			wrapped:        "é-\nc-\nl-\na-\nir",
			limit:          2,
			trimWhitespace: false,
			splitWord:      true,
		},
		{
			input:          "hello\rworld",
			wrapped:        "hello\nworld",
			limit:          10,
			trimWhitespace: true,
			splitWord:      false,
		},
		{
			input:          "foo\u2028bar baz",
			wrapped:        "foo\nbar\nbaz",
			limit:          5,
			trimWhitespace: true,
			splitWord:      false,
		},
		{
			input:          "Supercalifragilisticexpialidocious\vness",
			wrapped:        "Supercali-\nfragilist-\nicexpiali-\ndociousne-\nss",
			limit:          10,
			trimWhitespace: true,
			splitWord:      true,
		},
		{
			input:          "foo\u2028barbazbaz",
			wrapped:        "foo\nbarb-\nazbaz",
			limit:          5,
			trimWhitespace: true,
			splitWord:      true,
		},
		{
			input:          "",
			wrapped:        "",
			limit:          5,
			trimWhitespace: true,
			splitWord:      true,
		},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("Wrapped String Test %d", idx+1), func(t *testing.T) {
			wrapped, seq, err := wrapString(tt)
			expectedLines := 0
			if wrapped != "" {
				expectedLines = len(strings.Split(wrapped, "\n"))
			}
			assert.Nil(t, err)
			assert.Equal(t, expectedLines, len(seq.WrappedLines))
			assert.Equal(t, tt.wrapped, wrapped)
		})
	}
}

// TestStringWrap_WrappedStringSeq tests the WrappedStringSeq struct with a
// variety of test cases.
func TestStringWrap_WrappedStringSeq(t *testing.T) {
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

// TestStringWrapSplit_WrappedStringSeq tests the WrappedStringSeq struct with
// a variety of test cases.
func TestStringWrapSplit_WrappedStringSeq(t *testing.T) {
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
			OrigByteOffset:    LineOffset{Start: 0, End: 9},
			OrigRuneOffset:    LineOffset{Start: 0, End: 9},
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
			OrigByteOffset:    LineOffset{Start: 9, End: 18},
			OrigRuneOffset:    LineOffset{Start: 9, End: 18},
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
			OrigByteOffset:    LineOffset{Start: 18, End: 27},
			OrigRuneOffset:    LineOffset{Start: 18, End: 27},
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
			OrigByteOffset:    LineOffset{Start: 27, End: 37},
			OrigRuneOffset:    LineOffset{Start: 27, End: 37},
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
			OrigByteOffset:    LineOffset{Start: 37, End: 47},
			OrigRuneOffset:    LineOffset{Start: 37, End: 47},
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
			OrigByteOffset:    LineOffset{Start: 47, End: 56},
			OrigRuneOffset:    LineOffset{Start: 47, End: 56},
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
			OrigByteOffset:    LineOffset{Start: 56, End: 65},
			OrigRuneOffset:    LineOffset{Start: 56, End: 65},
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
			OrigByteOffset:    LineOffset{Start: 65, End: 74},
			OrigRuneOffset:    LineOffset{Start: 65, End: 74},
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
			OrigByteOffset:    LineOffset{Start: 74, End: 83},
			OrigRuneOffset:    LineOffset{Start: 74, End: 83},
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
			OrigByteOffset:    LineOffset{Start: 83, End: 87},
			OrigRuneOffset:    LineOffset{Start: 83, End: 87},
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
