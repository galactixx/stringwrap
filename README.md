<p align="center">
  <img src="/docs/logo.png" alt="stringwrap logo" width="75%"/>
</p>

Stringwrap is a Go package for wrapping strings by visual width with optional word splitting and full ANSI + grapheme cluster support. Designed for precision line-wrapping, ideal for terminal output, formatted logs, or editors that require accurate metadata per line segment.

## ✨ **Features**

This library offers robust string wrapping capabilities with a focus on accurate visual representation and detailed metadata:

* **Intelligent Wrapping:** Adapts to the **viewable width** by leveraging the `runewidth` library for precise character width calculation.
* **Comprehensive Text Handling:**
    * Ignores ANSI escape codes for width calculations while preserving them in the output.
    * Correctly processes Unicode grapheme clusters, ensuring accurate wrapping of emojis and accented characters.
    * Supports configurable tab sizes.
    * Respects hard breaks (`\n`) in the input string.
    * Provides optional word splitting for finer-grained control.
    * Handles non-breaking spaces (`\u00A0`) to prevent unwanted line breaks.
* **Detailed Line Metadata:** For each wrapped line, the library provides valuable information:
    * Byte and rune offsets within the original string.
    * The visual width of the wrapped line.
    * The index of the segment from the original line that this wrapped line belongs to.
    * An indication of whether the line ended due to a hard break or soft wrapping.
    * A flag indicating if the segment ends with a word that was split during wrapping.

## 🚀 **Getting Started**

```bash
go get github.com/galactixx/stringwrap@latest
```

## 📚 **Usage**

### Regular String Wrapping

```go
import "github.com/galactixx/stringwrap"

wrapped, meta, err := stringwrap.StringWrap("Hello world! 🌟", 10, 4)

fmt.Println(wrapped)
```

#### Output:
```text
Hello 
world! 🌟
```

### String Wrapping with Word Splitting

```go
wrapped, meta, err := stringwrap.StringWrapSplit("Supercalifragilisticexpialidocious", 10, 4)

fmt.Println(wrapped)
```

#### Output:
```text
Supercali-
fragilist-
icexpiali-
docious
```

### Accessing the Metadata

```go
for _, line := range meta.WrappedLines {
	fmt.Printf(
        "Line %d: width=%d, byteOffset=%v\n",
		line.CurLineNum,
        line.Width,
        line.OrigByteOffset
    )
}
```

## 🔍 **API**

### `func StringWrap(str string, limit int, tabSize int) (string, *WrappedStringSeq, error)`
Wraps a string at a visual width limit. Words are not split.

### `func StringWrapSplit(str string, limit int, tabSize int) (string, *WrappedStringSeq, error)`
Same as `StringWrap`, but allows splitting words across lines if needed.

### `type WrappedString struct`
Metadata for one wrapped segment.

```go
type WrappedString struct {
	CurLineNum        int
	OrigLineNum       int
	OrigByteOffset    LineOffset
	OrigRuneOffset    LineOffset
	SegmentInOrig     int
	NotWithinLimit    bool
	IsHardBreak       bool
	Width             int
	EndsWithSplitWord bool
}
```

### `type WrappedStringSeq struct`
Contains all wrapped lines and wrap configuration.

```go
type WrappedStringSeq struct {
	WrappedLines     []WrappedString
	WordSplitAllowed bool
	TabSize          int
	Limit            int
}
```

## 🤝 **License**

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

---

## 📞 **Contact**

If you have any questions or need support, feel free to reach out by opening an issue on the [GitHub repository](#).
