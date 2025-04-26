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

type WrappedString struct {
	OrigLineNum int
	CurLineNum  int
}

type WrappedStringSeq struct{ WrappedLines []WrappedString }

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
	// the current character number within a string line
	curLineWidth int
	curLineNum   int
	origLineNum  int
	curWordWidth int
}

func (p positions) curWritePosition() int { return p.curWordWidth + p.curLineWidth }

// incrementCurLine increases the current string line number
func (p *positions) incrementCurLine() { p.curLineNum += 1 }

// incrementOrigLine increases the original line number
func (p *positions) incrementOrigLine() { p.origLineNum += 1 }

// newWrappedString creates a WrappedString with both original and current line.
func (p positions) newWrappedString() WrappedString {
	return WrappedString{
		OrigLineNum: p.origLineNum, CurLineNum: p.curLineNum,
	}
}

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
type stringWrapStateMachine struct {
	lineBuffer    bytes.Buffer
	wordBuffer    bytes.Buffer
	wrappedBuffer bytes.Buffer

	pos              *positions
	wrappedStringSeq *WrappedStringSeq
	config           wordWrapConfig
	wordHasNbsp      bool
}

func (w *stringWrapStateMachine) writeANSIToLine(r rune) {
	w.lineBuffer.WriteRune(r)
}

// writeRuneToLine appends the given string directly to the lineBuffer.
func (w *stringWrapStateMachine) writeSpaceToLine(r rune) {
	runeLength := runewidth.RuneWidth(r)
	w.flushLineBuffer(runeLength)

	if w.pos.curLineWidth > 0 {
		w.lineBuffer.WriteRune(r)
	}
}

// writeRuneToWord appends a rune to the wordBuffer.
func (w *stringWrapStateMachine) writeStrToWord(str string) {
	w.wordBuffer.WriteString(str)
}

// writeRuneToWord appends a rune to the wordBuffer.
func (w *stringWrapStateMachine) writeRuneToWord(r rune) {
	w.wordBuffer.WriteRune(r)
}

// writeStrToLine appends the given string directly to the lineBuffer.
func (w *stringWrapStateMachine) writeStrToLine(str string) {
	w.flushLineBuffer(len(str))
	w.lineBuffer.WriteString(str)
}

// writeLine writes the current lineBuffer to the wrappedBuffer with a
// newline, then resets it.
func (w *stringWrapStateMachine) writeLine() {
	w.wrappedBuffer.WriteString(w.lineBuffer.String() + "\n")
	w.lineBuffer.Reset()
	wrappedString := w.pos.newWrappedString()
	w.wrappedStringSeq.appendWrappedSeq(wrappedString)
	w.pos.incrementCurLine()

	// since coming to end of a line, reset char counter to zero
	w.pos.curLineWidth = 0
}

// writeWord moves the contents of the wordBuffer into the lineBuffer,
// then resets the wordBuffer.
func (w *stringWrapStateMachine) writeWord() {
	w.lineBuffer.WriteString(w.wordBuffer.String())
	w.wordBuffer.Reset()
	w.pos.curLineWidth += w.pos.curWordWidth
	w.pos.curWordWidth = 0
}

// flushLineBuffer writes the current line if adding the next content
// would exceed the wrapping limit.
func (w *stringWrapStateMachine) flushLineBuffer(length int) {
	if w.pos.curLineWidth+length > w.config.limit {
		w.writeLine()
	}
}

func (w *stringWrapStateMachine) flushWordBuffer() {
	exceedsLimit := w.pos.curWritePosition() > w.config.limit
	if exceedsLimit && w.pos.curWordWidth == 0 {
		w.writeLine()
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
			w.writeLine()
			w.wordBuffer.Next(gIter.subWordBuffer.Len())
			w.pos.curWordWidth = runewidth.StringWidth(w.wordBuffer.String())
			w.flushWordBuffer()
		} else {
			if w.pos.curLineWidth > 0 {
				w.writeLine()
			}
			w.writeWord()
		}
	} else {
		w.writeWord()
	}
	w.wordHasNbsp = false
}

func StringWrap(
	str string, limit int, tabSize int, splitWord bool,
) (string, *WrappedStringSeq, error) {
	if limit < 2 {
		return "", nil, errors.New("limit must be greater than one")
	}

	var wrappedStringSeq WrappedStringSeq = WrappedStringSeq{
		WrappedLines: make([]WrappedString, 0),
	}

	// manage the current string line number taking into account wrapping
	var positions positions = positions{
		curLineWidth: 0,
		curLineNum:   1,
		origLineNum:  1,
		curWordWidth: 0,
	}

	// buffer to manage the wrapped output that results from the function
	stateMachine := stringWrapStateMachine{
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

		if rng := ansiRanges.nextRange(); rng != nil {
			stateMachine.writeANSIToLine(r)
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
				stateMachine.writeLine()
				positions.incrementOrigLine()
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
	stateMachine.writeLine()
	return stateMachine.wrappedBuffer.String(), &wrappedStringSeq, nil
}
