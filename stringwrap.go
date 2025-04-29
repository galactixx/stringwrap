package stringwrap

import (
	"bytes"
	"errors"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func isWordyGrapheme(grapheme string) bool {
	r, _ := utf8.DecodeRuneInString(grapheme)
	return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r)
}

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
	// The grapheme cluster start and end offsets of this
	// segment in the original unwrapped string.
	OrigGraphemeOffset LineOffset
	// Which segment number this is within the original line
	// (first, second, etc.).
	SegmentInOrig int
	// Whether the segment fits entirely within the wrapping
	// limit.
	NotWithinLimit bool
	// Optional reason why the segment was considered not
	// within limit (e.g., soft break, hard break).
	NotWithinLimitReason *string
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

type WrappedStringSeq struct {
	WrappedLines     []WrappedString
	WordSplitAllowed bool
	TabSize          int
	Limit            int
}

func (s *WrappedStringSeq) lastWrappedLine() WrappedString {
	return s.WrappedLines[len(s.WrappedLines)-1]
}

func (s *WrappedStringSeq) appendWrappedSeq(wrapped WrappedString) {
	s.WrappedLines = append(s.WrappedLines, wrapped)
}

type ansiRange struct {
	start int
	end   int
}

type ansiRanges struct{ ranges []*ansiRange }

func (r *ansiRanges) clearRange() { r.ranges = slices.Delete(r.ranges, 0, 1) }

func (r *ansiRanges) nextRange() *ansiRange {
	if len(r.ranges) > 0 {
		return r.ranges[0]
	}
	return nil
}

type graphemeWordIter struct {
	subWordBuffer   bytes.Buffer
	subWordWidth    int
	preLimitCluster string
	cluster         string
	graphemes       *uniseg.Graphemes
}

func (g *graphemeWordIter) needsHyphen() bool {
	return isWordyGrapheme(g.cluster) && isWordyGrapheme(g.preLimitCluster)
}

func (g *graphemeWordIter) iter(lineWidth int, limit int) {
	for g.graphemes.Next() && g.subWordWidth+lineWidth < limit {
		g.preLimitCluster = g.cluster
		g.cluster = g.graphemes.Str()
		g.subWordWidth += runewidth.StringWidth(g.cluster)
		g.subWordBuffer.WriteString(g.preLimitCluster)
	}
}

// manage the current string line number taking into account wrapping.
// thus, there is a variable that tracks the line number ignoring
// any wrapping and taking wrapping into account
type positions struct {
	curLineWidth      int
	curLineNum        int
	origLineNum       int
	curWordWidth      int
	origLineSegment   int
	origStartLineByte int
}

func (p positions) curWritePosition() int { return p.curWordWidth + p.curLineWidth }

// incrementCurLine increases the current string line number
func (p *positions) incrementCurLine() { p.curLineNum += 1 }

// incrementOrigLine increases the original line number
func (p *positions) incrementOrigLine() { p.origLineNum += 1 }

type wordWrapConfig struct {
	limit     int
	tabSize   int
	splitWord bool
}

// buffer to manage the wrapped output that results from the function and
// line and word buffers to manage the temporary states before writing
// to wrapped result buffer
//
// all runes for a specific line, wrapped or not, will be stored in
// temporary slice variable that resets after every line is
// established
type wrapStateMachine struct {
	lineBuffer bytes.Buffer
	wordBuffer bytes.Buffer
	buffer     bytes.Buffer

	pos              *positions
	wrappedStringSeq *WrappedStringSeq
	config           wordWrapConfig
	wordHasNbsp      bool
}

func (w *wrapStateMachine) writeANSIToLine(str string) {
	w.lineBuffer.WriteString(str)
}

// writeRuneToLine appends the given string directly to the lineBuffer.
func (w *wrapStateMachine) writeSpaceToLine(r rune) {
	runeLength := runewidth.RuneWidth(r)
	w.flushLineBuffer(runeLength)

	if w.pos.curLineWidth > 0 {
		w.lineBuffer.WriteRune(r)
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

// writeStrToLine appends the given string directly to the lineBuffer.
func (w *wrapStateMachine) writeStrToLine(str string) {
	w.flushLineBuffer(len(str))
	w.lineBuffer.WriteString(str)
}

func (w *wrapStateMachine) writeHardLine() { w.writeLine(true, false) }

func (w *wrapStateMachine) writeSoftLine(endsSplit bool) {
	w.writeLine(false, endsSplit)
}

// writeLine writes the current lineBuffer to the buffer with a
// newline, then resets it.
func (w *wrapStateMachine) writeLine(hardBreak bool, endsSplit bool) {
	lineToWrite := w.lineBuffer.String() + "\n"
	w.buffer.WriteString(lineToWrite)
	w.pos.origLineSegment += 1
	w.lineBuffer.Reset()
	origEndLineByte := w.pos.origStartLineByte + len(lineToWrite)
	origByteOffset := LineOffset{
		Start: w.pos.origStartLineByte, End: origEndLineByte,
	}

	wrappedString := WrappedString{
		OrigLineNum:       w.pos.origLineNum,
		CurLineNum:        w.pos.curLineNum,
		OrigByteOffset:    origByteOffset,
		SegmentInOrig:     w.pos.origLineSegment,
		NotWithinLimit:    w.pos.curLineWidth > w.config.limit,
		IsHardBreak:       hardBreak,
		Width:             w.pos.curLineWidth,
		EndsWithSplitWord: endsSplit,
	}
	w.wrappedStringSeq.appendWrappedSeq(wrappedString)
	w.pos.incrementCurLine()
	w.pos.origStartLineByte = origEndLineByte

	// since coming to end of a line, reset char counter to zero
	w.pos.curLineWidth = 0
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

func stringWrap(
	str string, limit int, tabSize int, splitWord bool,
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
			limit:     limit,
			tabSize:   tabSize,
			splitWord: splitWord,
		},
	}

	// precompute all ANSI codes ahead of time using regex.
	ansiRangesFound := ansiRegexp.FindAllStringIndex(str, -1)
	ansiRanges := ansiRanges{ranges: []*ansiRange{}}

	// turn all index ranges into a consolidated map for quick lookup
	for _, idxRange := range ansiRangesFound {
		ansiRanges.ranges = append(ansiRanges.ranges, &ansiRange{
			start: idxRange[0], end: idxRange[1],
		})
	}

	state := -1
	idx := 0

	// iterate through each rune in the string
	for idx < len(str) {
		r, rSize := utf8.DecodeRuneInString(str[idx:])
		if rng := ansiRanges.nextRange(); rng != nil && idx == rng.start {
			stateMachine.flushWordBuffer()
			stateMachine.writeANSIToLine(str[rng.start:rng.end])
			state = -1
			idx = rng.end
			ansiRanges.clearRange()
		} else if r == '\u00A0' {
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
				positions.curLineWidth += 1
			case '\n':
				stateMachine.writeHardLine()
				positions.incrementOrigLine()
				positions.origStartLineByte = 0
				positions.origLineSegment = 0
			case '\t':
				adjTabSize := tabSize - (positions.curLineWidth % tabSize)
				stateMachine.writeStrToLine(strings.Repeat(" ", adjTabSize))
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

	// Remove the last new line from the wrapped buffer
	lastWrappedLine := wrappedStringSeq.lastWrappedLine()
	if !lastWrappedLine.IsHardBreak {
		stateMachine.buffer.Truncate(stateMachine.buffer.Len() - 1)
	}
	return stateMachine.buffer.String(), &wrappedStringSeq, nil
}

func StringWrap(
	str string, limit int, tabSize int,
) (string, *WrappedStringSeq, error) {
	return stringWrap(str, limit, tabSize, false)
}

func StringWrapSplit(
	str string, limit int, tabSize int,
) (string, *WrappedStringSeq, error) {
	return stringWrap(str, limit, tabSize, true)
}
