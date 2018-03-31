package faketerm

import (
	"bytes"
	"reflect"
)

// diff.go provides functions that highlight differences between fake terminals.

// Diff compares the two terminals, returning an empty string if there is not
// difference. If a difference is found, returns a human readable description
// of the differences.
func Diff(want, got *Terminal) string {
	if reflect.DeepEqual(want, got) {
		return ""
	}

	var b bytes.Buffer
	b.WriteString("found differences between the two fake terminals.\n")
	b.WriteString("   got:\n")
	b.WriteString(got.String())
	b.WriteString("  want:\n")
	b.WriteString(want.String())
	b.WriteString("  diff (unexpected cells highlighted with rune '࿃'):\n")

	size := got.Size()
	for row := 0; row < size.Y; row++ {
		for col := 0; col < size.X; col++ {
			r := got.BackBuffer()[col][row].Rune
			if r != want.BackBuffer()[col][row].Rune {
				r = '࿃'
			} else if r == 0 {
				r = ' '
			}
			b.WriteRune(r)
		}
		b.WriteRune('\n')
	}
	return b.String()
}
