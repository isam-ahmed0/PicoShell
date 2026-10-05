package ui

import "unicode/utf8"

// Ellipsis is appended to text that had to be cut short.
const Ellipsis = "…"

// Fit shortens s with an ellipsis until it measures no wider than maxW in dp.
// Measurement goes through r, so truncation always matches what the backend
// will actually draw. Text that already fits is returned unchanged.
func Fit(r Rasterizer, f Font, s string, maxW float32) string {
	if r == nil || maxW <= 0 || s == "" {
		return ""
	}
	if r.Measure(f, s).W <= maxW {
		return s
	}

	dot := r.Measure(f, Ellipsis).W
	if dot >= maxW {
		return ""
	}

	// Rune-wise so multi-byte text is never split mid-character.
	cut := 0
	for i := 0; i < len(s); {
		_, size := utf8.DecodeRuneInString(s[i:])
		if r.Measure(f, s[:i+size]+Ellipsis).W > maxW {
			break
		}
		i += size
		cut = i
	}
	return s[:cut] + Ellipsis
}

// FitLeft shortens a path from the left, keeping the filename visible. Used by
// breadcrumbs and path labels, where the tail is what identifies a file.
func FitLeft(r Rasterizer, f Font, s string, maxW float32) string {
	if r == nil || maxW <= 0 || s == "" {
		return ""
	}
	if r.Measure(f, s).W <= maxW {
		return s
	}

	dot := r.Measure(f, Ellipsis).W
	if dot >= maxW {
		return ""
	}

	// Grow the kept tail one rune at a time from the end, so the last character
	// can never be split.
	best := -1
	for i := len(s); i > 0; {
		start := i - 1
		for start > 0 && !utf8Start(s[start]) {
			start--
		}
		if r.Measure(f, Ellipsis+s[start:]).W > maxW {
			break
		}
		best = start
		i = start
	}
	if best < 0 {
		return Ellipsis
	}
	return Ellipsis + s[best:]
}

func utf8Start(b byte) bool { return b&0xc0 != 0x80 }
