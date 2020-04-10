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
	"fmt"
	"strings"

	"github.com/mum4k/termdash/private/numbers"
	"github.com/mum4k/termdash/private/runewidth"
)

// fieldData are the data currently present inside the text input field.
type fieldData []rune

// String implements fmt.Stringer.
func (fd fieldData) String() string {
	var b strings.Builder
	for _, r := range fd {
		b.WriteRune(r)
	}
	return fmt.Sprintf("%q", b.String())
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

// cellsBefore given an endIdx calculates startIdx that results in range that
// will take at most the provided number of cells to print on the screen.
func (fd *fieldData) cellsBefore(cells, endIdx int) int {
	if endIdx == 0 {
		return 0
	}

	usedCells := 0
	for i := endIdx; i > 0; i-- {
		prev := (*fd)[i-1]
		width := runewidth.RuneWidth(prev)

		if usedCells+width > cells {
			return i
		}
		usedCells += width
	}
	return 0
}

// cellsAfter given a startIdx calculates endIdx that results in range that
// will take at most the provided number of cells to print on the screen.
func (fd *fieldData) cellsAfter(cells, startIdx int) int {
	if startIdx >= len(*fd) || cells == 0 {
		return startIdx
	}

	first := (*fd)[startIdx]
	usedCells := runewidth.RuneWidth(first)
	for i := startIdx + 1; i < len(*fd); i++ {
		r := (*fd)[i]
		width := runewidth.RuneWidth(r)
		if usedCells+width > cells {
			return i
		}
		usedCells += width
	}
	return len(*fd)
}

// minForArrows is the smallest number of cells in the window where we can
// indicate hidden text with left and right arrow.
const minForArrows = 3

// curMinIdx returns the lowest acceptable index for cursor position that is
// still within the visible range.
func curMinIdx(start, cells int) int {
	if start == 0 || cells < minForArrows {
		// The very first rune is visible, so the cursor can go all the way to
		// the start.
		return start
	}

	// When the first rune isn't visible, the cursor cannot go on the first
	// cell in the visible range since it contains the left arrow.
	return start + 1
}

// curMaxIdx returns the highest acceptable index for cursor position that is
// still within the visible range given the number of runes in data.
func curMaxIdx(start, end, cells, runeCount int) int {
	if end == runeCount+1 || cells < minForArrows {
		// The last rune is visible, so the cursor can go all the way to the
		// end.
		return end - 1
	}

	// When the last rune isn't visible, the cursor cannot go on the last cell
	// in the window that is reserved for appending text, since it contains the
	// right arrow.
	return end - 2
}

// shiftLeft shifts the visible range left so that it again contains the
// cursor.
// The visible range includes all fieldData indexes
// in range start <= idx < end.
func (fd *fieldData) shiftLeft(start, cells, curDataPos int) (int, int) {
	var startIdx int
	switch {
	case curDataPos == 0 || cells < minForArrows:
		startIdx = curDataPos

	default:
		startIdx = curDataPos - 1
	}
	forRunes := cells - 1
	endIdx := fd.cellsAfter(forRunes, startIdx)
	endIdx++ // Space for the cursor.

	return startIdx, endIdx
}

// shiftRight shifts the visible range right so that it again contains the
// cursor.
// The visible range includes all fieldData indexes
// in range start <= idx < end.
func (fd *fieldData) shiftRight(start, cells, curDataPos int) (int, int) {
	var endIdx int
	switch dataLen := len(*fd); {
	case curDataPos == dataLen:
		// Cursor is in the empty space after the data.
		// Print all runes until the end of data.
		endIdx = dataLen

	default:
		// Cursor is within the data, print all runes including the one the
		// cursor is on.
		endIdx = curDataPos + 1
	}

	forRunes := cells - 1
	startIdx := fd.cellsBefore(forRunes, endIdx)

	// Invariant, if counting form the back ends in the middle of a full-width
	// rune, cellsAfter doesn't include the full-width rune. This means that we
	// might have recovered space for one half-with rune at the end if there is
	// one.
	endIdx = fd.cellsAfter(forRunes, startIdx)
	endIdx++ // Space for the cursor.

	return startIdx, endIdx
}

// lastVisible given an end index of visible range asserts whether the last
// rune in the data is visible.
// The visible range includes all fieldData indexes
// in range start <= idx < end.
func (fd *fieldData) lastVisible(end int) bool {
	return end-1 >= len(*fd)
}

// runesIn returns all the runes in the visible range.
// The visible range includes all fieldData indexes
// in range start <= idx < end.
func (fd *fieldData) runesIn(start, end int) []rune {
	var runes []rune
	for i, r := range (*fd)[start:] {
		if i+start > end-2 { // One last space is for the cursor after the text.
			break
		}
		runes = append(runes, r)
	}
	return runes
}

// fitRunes starting from the firstRune index returns runes that take at most
// the specified number of cells. The last cell is reserved for a cursor
// position used for appending new runes.
// This might return smaller number of runes than the size of the range,
// depending on the width of the individual runes.
// Returns the text and the start and end positions within the data.
func (fd *fieldData) fitRunes(firstRune, curPos, cells int) (string, int, int) {
	forRunes := cells - 1 // One cell reserved for the cursor when appending.

	// Determine how many runes fit from the start.
	start := firstRune
	end := fd.cellsAfter(forRunes, start)
	end++

	if start > 0 && fd.lastVisible(end) {
		// Start is in the middle, end is visible.
		// Fit runes from the end.
		end = len(*fd)
		start = fd.cellsBefore(forRunes, end)
		end++ // Space for the cursor within the visible range.
	}

	// The fitting of runes might have resulted in a visible range that no
	// longer contains the cursor (it became shorter) or the cursor was outside
	// to begin with (due to cursorLeft() or cursorRight() calls).
	// Shift the range so the cursor is again inside.
	if curPos < curMinIdx(start, cells) {
		start, end = fd.shiftLeft(start, cells, curPos)
	} else if curPos > curMaxIdx(start, end, cells, len(*fd)) {
		start, end = fd.shiftRight(start, cells, curPos)
	}

	runes := fd.runesIn(start, end)
	useArrows := cells >= minForArrows
	var b strings.Builder
	for i, r := range runes {
		switch {
		case useArrows && i == 0 && start > 0:
			// Indicate that start is hidden by replacing the first visible
			// rune with an arrow.
			b.WriteRune('⇦')
			if rw := runewidth.RuneWidth(r); rw == 2 {
				// If the replaced rune was a full-width rune, place two arrows
				// to keep the same space allocation as pre-calculated.
				b.WriteRune('⇦')
			}

		default:
			b.WriteRune(r)
		}
	}

	if useArrows && !fd.lastVisible(end) {
		// Indicate that end is hidden by placing an arrow at the end.
		// THis has no impact on space allocation, since the last cell is
		// always reserved for the cursor or the arrow.
		b.WriteRune('⇨')
	}
	return b.String(), start, end
}

// fieldEditor maintains the cursor position and allows editing of the data in
// the text input field.
// This object isn't thread-safe.
type fieldEditor struct {
	// data are the data currently present in the text input field.
	data fieldData

	// curDataPos is the current position of the cursor within the data.
	// The cursor is allowed to go one cell beyond the data so appending is
	// possible.
	curDataPos int

	// firstRune is the index of the first displayed rune in the text input
	// field.
	firstRune int

	// width is the width of the text input field last time viewFor was called.
	width int
}

// newFieldEditor returns a new fieldEditor instance.
func newFieldEditor() *fieldEditor {
	return &fieldEditor{}
}

// minFieldWidth is the minimum supported width of the text input field.
const minFieldWidth = 4

// curCell returns the index of the cell the cursor is in within the text input field.
func (fe *fieldEditor) curCell(width int) int {
	if width == 0 {
		return 0
	}
	// The index of rune within the visible range the cursor is at.
	runeNum := fe.curDataPos - fe.firstRune

	cellNum := 0
	rn := 0
	for i, r := range fe.data {
		if i < fe.firstRune {
			continue
		}
		if rn >= runeNum {
			break
		}
		rn++
		cellNum += runewidth.RuneWidth(r)
	}
	return cellNum
}

// viewFor returns the currently visible data inside a text field with the
// specified width and the cursor position within the field.
func (fe *fieldEditor) viewFor(width int) (string, int, error) {
	if min := minFieldWidth; width < min { // One for left arrow, two for one full-width rune and one for the cursor.
		return "", -1, fmt.Errorf("width %d is too small, the minimum is %d", width, min)
	}
	runes, start, _ := fe.data.fitRunes(fe.firstRune, fe.curDataPos, width)
	fe.firstRune = start
	fe.width = width
	return runes, fe.curCell(width), nil
}

// content returns the string content in the field editor.
func (fe *fieldEditor) content() string {
	return string(fe.data)
}

// reset resets the content back to zero.
func (fe *fieldEditor) reset() {
	*fe = *newFieldEditor()
}

// insert inserts the rune at the current position of the cursor.
func (fe *fieldEditor) insert(r rune) {
	rw := runewidth.RuneWidth(r)
	if rw == 0 {
		// Don't insert invisible runes.
		return
	}
	fe.data.insertAt(fe.curDataPos, r)
	fe.curDataPos++
}

// delete deletes the rune at the current position of the cursor.
func (fe *fieldEditor) delete() {
	if fe.curDataPos >= len(fe.data) {
		// Cursor not on a rune, nothing to do.
		return
	}
	fe.data.deleteAt(fe.curDataPos)
}

// deleteBefore deletes the rune that is immediately to the left of the cursor.
func (fe *fieldEditor) deleteBefore() {
	if fe.curDataPos == 0 {
		// Cursor at the beginning, nothing to do.
		return
	}
	fe.cursorLeft()
	fe.delete()
}

// cursorRight moves the cursor one position to the right.
func (fe *fieldEditor) cursorRight() {
	fe.curDataPos, _ = numbers.MinMaxInts([]int{fe.curDataPos + 1, len(fe.data)})
}

// cursorLeft moves the cursor one position to the left.
func (fe *fieldEditor) cursorLeft() {
	_, fe.curDataPos = numbers.MinMaxInts([]int{fe.curDataPos - 1, 0})
}

// cursorStart moves the cursor to the beginning of the data.
func (fe *fieldEditor) cursorStart() {
	fe.curDataPos = 0
}

// cursorEnd moves the cursor to the end of the data.
func (fe *fieldEditor) cursorEnd() {
	fe.curDataPos = len(fe.data)
}

// cursorRelCell sets the cursor onto the cell index within the visible
// area.
// If the index falls before the window, the cursor is moved onto the first
// visible position.
// If the pos falls after the end of data, the cursor is moved onto the last
// visible position.
func (fe *fieldEditor) cursorRelCell(cellIdx int) {
	runes, start, end := fe.data.fitRunes(fe.firstRune, fe.curDataPos, fe.width)
	minDataIdx := curMinIdx(start, fe.width)
	maxDataIdx := curMaxIdx(start, end, fe.width, len(fe.data))

	// Index of the rune we should move the cursor to relative to the visible
	// range.
	var relRuneIdx int
	var cell int
	for _, r := range runes {
		cell += runewidth.RuneWidth(r)
		if cell > cellIdx {
			break
		}
		relRuneIdx++
	}

	// Absolute index of the rune we should move the cursor to.
	dataIdx := fe.firstRune + relRuneIdx
	switch {
	case dataIdx < minDataIdx:
		fe.curDataPos = minDataIdx

	case dataIdx > maxDataIdx:
		fe.curDataPos = maxDataIdx

	default:
		fe.curDataPos = dataIdx
	}
}
