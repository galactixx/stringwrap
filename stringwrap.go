package stringwrap

import (
	"bytes"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/galactixx/ansiwalker"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// isWordyGrapheme returns true if the first rune in the grapheme cluster
// is considered part of a word (i.e., a letter, number, or combining mark).
func isWordyGrapheme(grapheme string) bool {
	r, _ := utf8.DecodeRuneInString(grapheme)
	return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r)
}

// btoi is a simple function to convert a boolean to an integer
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// LineOffset represents a half-open interval [Start, End) that describes
// either the byte offset or rune offset range of a wrapped segment
// in the original unwrapped string.
type LineOffset struct {
	Start int
	End   int
}

type WrappedString struct {
	// The current wrapped line number (after wrapping).
	CurLineNum int
	// The original unwrapped line number this segment came
	// from.
	OrigLineNum int
	// The byte start and end offsets of this segment in the
	// original unwrapped string.
	OrigByteOffset LineOffset
	// The rune start and end offsets of this segment in the
	// original unwrapped string.
	OrigRuneOffset LineOffset
	// Which segment number this is within the original line
	// (first, second, etc.).
	SegmentInOrig int
	// Whether this segment is the last from the original
	// ilne within the unwrapped string.
	LastSegmentInOrig bool
	// Whether the segment fits entirely within the wrapping
	// limit.
	NotWithinLimit bool
	// Whether the wrap was due to a hard break (newline)
	// instead of word wrapping.
	IsHardBreak bool
	// The viewable width of the wrapped string.
	Width int
	// Whether this wrapped segment ends with a split word due
	// to reaching the wrapping limit
	// (e.g., a hyphen may be added).
	EndsWithSplitWord bool
}

// WrappedStringSeq holds the sequence of wrapped lines produced by
// the string wrapping process, along with the configuration used.
type WrappedStringSeq struct {
	// WrappedLines is the list of individual wrapped segments with
	// metadata.
	WrappedLines []WrappedString
	// WordSplitAllowed indicates whether splitting words across
	// lines is permitted.
	WordSplitAllowed bool
	// TabSize defines how many spaces a tab character expands to.
	TabSize int
	// Limit is the maximum viewable width allowed per line.
	Limit int
}

// lastWrappedLine pulls the last wrapped line that has been parsed
func (s *WrappedStringSeq) lastWrappedLine() *WrappedString {
	return &s.WrappedLines[len(s.WrappedLines)-1]
}

// appendWrappedSeq adds a new WrappedString to the existing slice
func (s *WrappedStringSeq) appendWrappedSeq(wrapped WrappedString) {
	s.WrappedLines = append(s.WrappedLines, wrapped)
}

// graphemeWordIter manages state for iterating through each word
// to determine the split point when word splitting is enabled
type graphemeWordIter struct {
	subWordBuffer   bytes.Buffer
	subWordWidth    int
	preLimitCluster string
	cluster         string
	graphemes       *uniseg.Graphemes
}

// needsHyphen returns true if a hyphen should be added when
// word splitting
func (g *graphemeWordIter) needsHyphen() bool {
	return isWordyGrapheme(g.cluster) && isWordyGrapheme(g.preLimitCluster)
}

// iter iterates through the word buffer until the limit
// is exceeded
func (g *graphemeWordIter) iter(lineWidth int, limit int) {
	for g.graphemes.Next() && g.subWordWidth+lineWidth < limit {
		g.preLimitCluster = g.cluster
		g.cluster = g.graphemes.Str()
		g.subWordWidth += runewidth.StringWidth(g.cluster)
		g.subWordBuffer.WriteString(g.preLimitCluster)
	}
}

// positions holds state for a variety of positional info
type positions struct {
	curLineWidth      int
	curLineNum        int
	origLineNum       int
	curWordWidth      int
	origLineSegment   int
	origStartLineByte int
	origStartLineRune int
	timmedWhiteSpace  int
}

// endLineCalc calculates the end byte/rune index
func (p positions) endCalc(count int, lineCount int, hardBreak bool) int {
	origEndLine := count + lineCount - 1 + btoi(hardBreak)
	return origEndLine + p.timmedWhiteSpace
}

// getEndLineByte calculates the end byte index and offset
func (p positions) endByte(line string, hardBreak bool) (int, LineOffset) {
	endLine := p.endCalc(p.origStartLineByte, len(line), hardBreak)
	return endLine, LineOffset{Start: p.origStartLineByte, End: endLine}
}

// getEndLineRune calculates the end rune index and offset
func (p positions) endRune(line string, hardBreak bool) (int, LineOffset) {
	endLine := p.endCalc(
		p.origStartLineRune,
		utf8.RuneCountInString(line),
		hardBreak,
	)
	return endLine, LineOffset{Start: p.origStartLineRune, End: endLine}
}

// returns the current viewable width (word + line)
func (p positions) curWritePosition() int { return p.curWordWidth + p.curLineWidth }

// incrementCurLine increases the current string line number
func (p *positions) incrementCurLine() { p.curLineNum += 1 }

// incrementOrigLine increases the original line number
func (p *positions) incrementOrigLine() { p.origLineNum += 1 }

// a struct to hold all configuration information
type wordWrapConfig struct {
	limit          int
	tabSize        int
	trimWhitespace bool
	splitWord      bool
}

// buffer to manage the wrapped output that results from the function and
// line and word buffers to manage the temporary states before writing
// to wrapped result buffer
type wrapStateMachine struct {
	lineBuffer bytes.Buffer
	wordBuffer bytes.Buffer
	buffer     bytes.Buffer

	pos              *positions
	wrappedStringSeq *WrappedStringSeq
	config           wordWrapConfig
	wordHasNbsp      bool
}

// writeANSIToLine writes ANSI to the line buffer
func (w *wrapStateMachine) writeANSIToLine(str string) {
	w.lineBuffer.WriteString(str)
}

// writeRuneToLine appends the given string directly to the lineBuffer.
func (w *wrapStateMachine) writeSpaceToLine(r rune) {
	w.flushLineBuffer(1)
	if !w.config.trimWhitespace || w.pos.curLineWidth > 0 {
		w.lineBuffer.WriteRune(r)
		w.pos.curLineWidth += 1
	} else {
		w.pos.timmedWhiteSpace += 1
	}
}

// writeRuneToWord appends a rune to the wordBuffer.
func (w *wrapStateMachine) writeStrToWord(str string) {
	w.wordBuffer.WriteString(str)
}

// writeRuneToWord appends a rune to the wordBuffer.
func (w *wrapStateMachine) writeRuneToWord(r rune) {
	w.wordBuffer.WriteRune(r)
}

// writeTabToLine appends the given tab size in spaces to the lineBuffer.
func (w *wrapStateMachine) writeTabToLine() int {
	adjTabSize := w.config.tabSize - (w.pos.curLineWidth % w.config.tabSize)
	w.flushLineBuffer(adjTabSize)

	if w.lineBuffer.Len() == 0 {
		if w.config.trimWhitespace {
			adjTabSize = 0
			w.pos.timmedWhiteSpace += 1
		} else {
			adjTabSize = w.config.tabSize
		}
	}
	tabSpaces := strings.Repeat(" ", adjTabSize)
	w.lineBuffer.WriteString(tabSpaces)
	return adjTabSize
}

// writeHardLine is used to write a hard break
func (w *wrapStateMachine) writeHardLine() { w.writeLine(true, false) }

// writeSoftLine is used to write a soft break
func (w *wrapStateMachine) writeSoftLine(endsSplit bool) {
	w.writeLine(false, endsSplit)
}

// writeLine writes the current lineBuffer to the buffer with a
// newline, then resets it.
func (w *wrapStateMachine) writeLine(hardBreak bool, endsSplit bool) {
	newLine := w.lineBuffer.String()
	if w.config.trimWhitespace {
		newLine = strings.TrimRightFunc(newLine, unicode.IsSpace)
		trimWidth := runewidth.StringWidth(newLine)
		w.pos.timmedWhiteSpace += w.pos.curLineWidth - trimWidth
		w.pos.curLineWidth = trimWidth
	}
	newLine += "\n"

	w.buffer.WriteString(newLine)
	w.pos.origLineSegment += 1
	w.lineBuffer.Reset()
	origEndLineByte, origByteOffset := w.pos.endByte(newLine, hardBreak)
	origEndLineRune, origRuneOffset := w.pos.endRune(newLine, hardBreak)

	wrappedString := WrappedString{
		OrigLineNum:       w.pos.origLineNum,
		CurLineNum:        w.pos.curLineNum,
		OrigByteOffset:    origByteOffset,
		OrigRuneOffset:    origRuneOffset,
		SegmentInOrig:     w.pos.origLineSegment,
		LastSegmentInOrig: hardBreak,
		NotWithinLimit:    w.pos.curLineWidth > w.config.limit,
		IsHardBreak:       hardBreak,
		Width:             w.pos.curLineWidth,
		EndsWithSplitWord: endsSplit,
	}
	w.wrappedStringSeq.appendWrappedSeq(wrappedString)
	w.pos.incrementCurLine()
	w.pos.origStartLineByte = origEndLineByte
	w.pos.origStartLineRune = origEndLineRune

	// since coming to end of a line, reset char counter to zero
	w.pos.curLineWidth = 0
	w.pos.timmedWhiteSpace = 0
}

// writeWord moves the contents of the wordBuffer into the lineBuffer,
// then resets the wordBuffer.
func (w *wrapStateMachine) writeWord() {
	w.lineBuffer.WriteString(w.wordBuffer.String())
	w.wordBuffer.Reset()
	w.pos.curLineWidth += w.pos.curWordWidth
	w.pos.curWordWidth = 0
}

// flushLineBuffer writes the current line if adding the next content
// would exceed the wrapping limit.
func (w *wrapStateMachine) flushLineBuffer(length int) {
	if w.pos.curLineWidth+length > w.config.limit {
		w.writeSoftLine(false)
	}
}

// flushes the word buffer when a word has been written
func (w *wrapStateMachine) flushWordBuffer() {
	exceedsLimit := w.pos.curWritePosition() > w.config.limit
	if exceedsLimit && w.pos.curWordWidth == 0 {
		w.writeSoftLine(false)
		return
	}

	if exceedsLimit {
		if w.config.splitWord && !w.wordHasNbsp {
			gIter := graphemeWordIter{
				graphemes: uniseg.NewGraphemes(w.wordBuffer.String()),
			}
			gIter.iter(w.pos.curLineWidth, w.config.limit)

			w.lineBuffer.WriteString(gIter.subWordBuffer.String())
			if gIter.needsHyphen() {
				w.lineBuffer.WriteRune('-')
			}
			w.pos.curLineWidth += gIter.subWordWidth
			w.writeSoftLine(gIter.needsHyphen())
			w.wordBuffer.Next(gIter.subWordBuffer.Len())
			w.pos.curWordWidth = runewidth.StringWidth(w.wordBuffer.String())
			w.flushWordBuffer()
		} else {
			if w.pos.curLineWidth > 0 {
				w.writeSoftLine(false)
			}
			w.writeWord()
		}
	} else {
		w.writeWord()
	}
	w.wordHasNbsp = false
}

// general function that implements the core string wrap logic
func stringWrap(
	str string, limit int, tabSize int, trimWhitespace bool, splitWord bool,
) (string, *WrappedStringSeq, error) {
	if limit < 2 {
		return "", nil, errors.New("limit must be greater than one")
	}

	var wrappedStringSeq WrappedStringSeq = WrappedStringSeq{
		WordSplitAllowed: splitWord,
		TabSize:          tabSize,
		Limit:            limit,
	}

	// manage the current string line number taking into account wrapping
	var positions positions = positions{
		curLineNum:  1,
		origLineNum: 1,
	}

	// buffer to manage the wrapped output that results from the function
	stateMachine := wrapStateMachine{
		pos:              &positions,
		wrappedStringSeq: &wrappedStringSeq,
		config: wordWrapConfig{
			limit:          limit,
			tabSize:        tabSize,
			trimWhitespace: trimWhitespace,
			splitWord:      splitWord,
		},
	}

	state := -1
	idx := 0

	// iterate through each rune in the string
	for idx < len(str) {
		r, rSize, next, ok := ansiwalker.ANSIWalk(str, idx)
		rIdx := next - rSize
		if ok && rIdx > idx {
			stateMachine.flushWordBuffer()
			stateMachine.writeANSIToLine(str[idx:rIdx])
			state = -1
		}
		idx = rIdx

		if r == '\u00A0' {
			stateMachine.wordHasNbsp = true
			stateMachine.writeRuneToWord(r)
			positions.curWordWidth += 1
			idx += rSize
		} else if unicode.IsSpace(r) {
			stateMachine.flushWordBuffer()

			// All legacy whitespace is ignored and not written
			switch r {
			case ' ':
				stateMachine.writeSpaceToLine(r)
			case '\n':
				stateMachine.writeHardLine()
				positions.incrementOrigLine()
				positions.origLineSegment = 0
			case '\t':
				adjTabSize := stateMachine.writeTabToLine()
				positions.curLineWidth += adjTabSize
			}
			state = -1
			idx += rSize
		} else {
			cluster, _, _, st := uniseg.StepString(str[idx:], state)
			state = st

			if cluster != "" {
				clusterWidth := runewidth.StringWidth(cluster)
				positions.curWordWidth += clusterWidth

				// Writer cluster string to word and then check word buffer
				stateMachine.writeStrToWord(cluster)
				idx += len(cluster)
			} else {
				idx += rSize
			}
		}
	}

	// write word and line buffers after iteration is done
	stateMachine.flushWordBuffer()
	if stateMachine.lineBuffer.Len() > 0 {
		stateMachine.writeSoftLine(false)
	}

	// remove the last new line from the wrapped buffer
	lastWrappedLine := wrappedStringSeq.lastWrappedLine()
	if !lastWrappedLine.IsHardBreak {
		stateMachine.buffer.Truncate(stateMachine.buffer.Len() - 1)
		lastWrappedLine.LastSegmentInOrig = true
	}
	return stateMachine.buffer.String(), &wrappedStringSeq, nil
}

// StringWrap wraps the input string to the specified viewable width limit,
// expanding tabs using the given tab size. It preserves word boundaries
// and does not split words across lines.
//
// If trimWhitespace is true, leading and trailing whitespace on each wrapped
// line will be stripped before the newline is appended.
//
// ANSI escape sequences are preserved without contributing to visual width.
// Returns the wrapped string and a metadata sequence describing each
// wrapped line.
func StringWrap(str string, limit int, tabSize int, trimWhitespace bool) (
	string, *WrappedStringSeq, error,
) {
	return stringWrap(str, limit, tabSize, trimWhitespace, false)
}

// StringWrapSplit wraps the input string to the specified viewable width
// limit, expanding tabs using the given tab size. Unlike StringWrap, this
// function allows words to be split across lines if they exceed the
// wrapping limit.
//
// If trimWhitespace is true, leading and trailing whitespace on each wrapped
// line will be stripped before the newline is appended.
//
// ANSI escape sequences are preserved without contributing to visual width.
// Returns the wrapped string and a metadata sequence describing each
// wrapped line.
func StringWrapSplit(str string, limit int, tabSize int, trimWhitespace bool) (
	string, *WrappedStringSeq, error,
) {
	return stringWrap(str, limit, tabSize, trimWhitespace, true)
}
