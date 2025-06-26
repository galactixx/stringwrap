<p align="center">
  <img src="/docs/logo.png" alt="stringwrap logo" width="50%"/>
</p>

Stringwrap is a Go package for wrapping strings by visual width with optional word splitting and full ANSI + grapheme cluster support. Designed for precision line-wrapping, ideal for terminal output, formatted logs, or editors that require accurate metadata per line segment.

## ✨ **Features**

**General Wrapping**
* Ignores ANSI escape codes for width calculations while preserving them in the output.
* Correctly processes Unicode grapheme clusters.
* Supports configurable tab sizes.
* Respects hard breaks (`\n`) in the input string.
* Provides optional word splitting for finer-grained control.
* Handles non-breaking spaces (`\u00A0`) to prevent unwanted line breaks.

**Wrapped-Line Metadata**
* Byte and rune offsets within the original string.
* The visual width of the wrapped line.
* The index of the segment from the original line that this wrapped line belongs to.
* An indication of whether the line ended due to a hard break or soft wrapping.
* A flag indicating if the segment ends with a word that was split during wrapping.

## 💡 **Why Grapheme Clusters Matter**

Both `StringWrap` and `StringWrapSplit` use Unicode grapheme cluster parsing (via the `uniseg` library) rather than simple rune iteration. This is crucial for accurate width calculation with complex Unicode sequences:

* **ZWJ Emojis:** Sequences like "👩‍💻" contain multiple runes but display as a single character
* **Combining Marks:** Characters like "é" must be treated as one unit

While this approach is slower than rune-based processing, it prevents incorrect wrapping that would occur with naive rune counting.

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
