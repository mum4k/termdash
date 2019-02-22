// runewidth is a wrapper over github.com/mattn/go-runewidth which gives
// different treatment to certain runes with ambiguous width.
package runewidth

import runewidth "github.com/mattn/go-runewidth"

// RuneWidth returns the number of cells needed to draw r.
// Background in http://www.unicode.org/reports/tr11/.
//
// Treats runes used internally by termdash as single-cell (half-width) runes
// regardless of the locale. I.e. runes that are used to draw lines, boxes,
// indicate resize or text trimming was needed and runes used by the braille
// canvas.
//
// This should be safe, since even in locales where these runes have ambiguous
// width, we still place all the character content around them so they should
// have be half-width.
func RuneWidth(r rune) int {
	return runewidth.RuneWidth(r)
}

// StringWidth is like RuneWidth, but returns the number of cells occupied by
// all the runes in the string.
func StringWidth(s string) int {
	var width int
	for _, r := range []rune(s) {
		width += RuneWidth(r)
	}
	return width
}
