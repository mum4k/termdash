// Copyright 2019 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package textinput

// editor.go contains data types that edit the content of the text input field.

import (
	"bytes"
	"fmt"

	"github.com/mum4k/termdash/internal/runewidth"
)

// fieldData are the data currently present inside the text input field.
type fieldData []rune

// String implements fmt.Stringer.
func (fd fieldData) String() string {
	var b bytes.Buffer
	for _, r := range fd {
		b.WriteRune(r)
	}
	return b.String()
}

// insertAt inserts rune at the specified index.
func (fd *fieldData) insertAt(idx int, r rune) {
	*fd = append(
		(*fd)[:idx],
		append(fieldData{r}, (*fd)[idx:]...)...,
	)
}

// deleteAt deletes rune at the specified index.
func (fd *fieldData) deleteAt(idx int) {
	*fd = append((*fd)[:idx], (*fd)[idx+1:]...)
}

// rangeWidth returns the rune width of all the runes in range:
//   startIdx <= idx < endIdx
func (fd *fieldData) rangeWidth(startIdx, endIdx int) int {
	var width int
	for _, r := range (*fd)[startIdx:endIdx] {
		width += runewidth.RuneWidth(r)
	}
	return width
}

// visibleRange represents a range of currently visible runes.
// Visible runes are all such runes whose index falls within:
//   startIdx <= idx < endIdx
type visibleRange struct {
	startIdx int
	endIdx   int
}

// fieldEditor maintains the cursor position and allows editing of the data in
// the text input field.
// This object isn't thread-safe.
type fieldEditor struct {
	// data are the data currently present in the text input field.
	data fieldData

	// curPos is the current position of the cursor within the data.
	curPos int

	// visible is the currently visible range.
	visible visibleRange
}

// newFieldEditor returns a new fieldEditor instance.
func newFieldEditor() *fieldEditor {
	return &fieldEditor{}
}

// viewFor returns the currently visible data inside a text field with the
// specified width and the cursor position within the field.
func (fe *fieldEditor) viewFor(width int) (string, int, error) {
	if min := 3; width < min {
		return "", -1, fmt.Errorf("width %d is too small, the minimum is %d", width, min)
	}

	/*
		case1: range is zero - initialize to width
		case2: range is set, cursor is in
		case3: range is set, cursor is to the right - shift range right, calculate left based on rune width.
		case4: range is set, cursor is to the left - shift range left, calculate right based on rune width.

		available:
		case1: data < width => width - 1 // one for the cursor
		case2: data >= width && left edge visible => width - 1 // one for the arrow
		case3: data >= width && right edge visible => width - 2 // one for the left arrow and one for the cursor on the right
		case4: data >= width && no edge visible => width -2 // two for the two arrows
	*/

	return "", fe.curPos, nil
}

// insert inserts the rune at the current position of the cursor.
func (fe *fieldEditor) insert(r rune) {
	fe.data.insertAt(fe.curPos, r)
	fe.curPos++
}

// delete deletes the rune at the current position of the cursor.
func (fe *fieldEditor) delete(r rune) {}

// deleteBefore deletes the rune that is immediately to the left of the cursor.
func (fe *fieldEditor) deleteBefore(r rune) {}

// cursorRight moves the cursor one position to the right.
func (fe *fieldEditor) cursorRight() {
}

// cursorLeft moves the cursor one position to the left.
func (fe *fieldEditor) cursorLeft() {
}

// cursorHome moves the cursor to the beginning of the data.
func (fe *fieldEditor) cursorHome() {}

// cursorEnd moves the cursor to the end of the data.
func (fe *fieldEditor) cursorEnd() {}
