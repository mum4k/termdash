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

	"github.com/mum4k/termdash/internal/numbers"
	"github.com/mum4k/termdash/internal/runewidth"
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

// startVisible asserts whether the first rune is within the visible range.
func (fd *fieldData) startVisible(vr *visibleRange) bool {
	return vr.startIdx == 0
}

// endVisible asserts whether the last rune is within the visible range.
// The last position in the visible range is reserved for the cursor or an
// arrow.
func (fd *fieldData) endVisible(vr *visibleRange) bool {
	return vr.endIdx-1 >= len(*fd)
}

// minForArrows is the smallest number of cells in the window where we can
// indicate hidden text with left and right arrow.
const minForArrows = 3

// runesIn returns runes that are in the visible range.
func (fd *fieldData) runesIn(vr *visibleRange) string {
	var runes []rune
	for i, r := range (*fd)[vr.startIdx:] {
		if i+vr.startIdx > vr.endIdx-2 {
			break
		}
		runes = append(runes, r)
	}

	useArrows := vr.cells() >= minForArrows
	var b strings.Builder
	for i, r := range runes {
		switch {
		case useArrows && i == 0 && !fd.startVisible(vr):
			b.WriteRune('⇦')

		default:
			b.WriteRune(r)
		}
	}

	if useArrows && !fd.endVisible(vr) {
		b.WriteRune('⇨')
	}
	return b.String()
}

// visibleRange represents a range of currently visible cells.
// Visible cells are all cells whose index falls within:
//   startIdx <= idx < endIdx
// Not all of these cells are available for runes, the last cell is reserved
// for the cursor to append data or for an arrow indicating that the text is
// scrolling. See forRunes().
type visibleRange struct {
	startIdx int
	endIdx   int
}

// forRunes returns the number of cells that are usable for runes.
// Part of the visible range is reserved for the cursor at the end of the data.
func (vr *visibleRange) forRunes() int {
	cells := vr.cells()
	if cells < 1 {
		return 0
	}
	return cells - 1 // One cell reserved for the cursor at the end.
}

// cells returns the number of cells in the range.
func (vr *visibleRange) cells() int {
	return vr.endIdx - vr.startIdx
}

// contains asserts whether the provided index is in the range.
func (vr *visibleRange) contains(idx int) bool {
	return idx >= vr.startIdx && idx < vr.endIdx
}

// set sets the visible range from the start to the end index.
func (vr *visibleRange) set(startIdx, endIdx int) {
	vr.startIdx = startIdx
	vr.endIdx = endIdx
}

// curMinIdx returns the lowest acceptable index for cursor position that is
// still within the visible range.
func (vr *visibleRange) curMinIdx() int {
	if vr.cells() == 0 {
		return vr.startIdx
	}

	if vr.startIdx == 0 || vr.cells() < minForArrows {
		// The very first rune is visible, so the cursor can go all the way to
		// the start.
		return vr.startIdx
	}

	// When the first rune isn't visible, the cursor cannot go on the first
	// cell in the visible range since it contains the left arrow.
	return vr.startIdx + 1
}

// curMaxIdx returns the highest acceptable index for cursor position that is
// still within the visible range given the number of runes in data.
func (vr *visibleRange) curMaxIdx(runeCount int) int {
	if vr.cells() == 0 {
		return vr.startIdx
	}

	if vr.endIdx == runeCount || vr.endIdx == runeCount+1 || vr.cells() < minForArrows {
		// The last rune is visible, so the cursor can go all the way to the
		// end.
		return vr.endIdx - 1
	}

	// When the last rune isn't visible, the cursor cannot go on the last cell
	// in the window that is reserved for appending text, since it contains the
	// right arrow.
	return vr.endIdx - 2
}

// normalizeToWidth normalizes the visible range, handles cases where the width of the
// text input field changed (terminal resize).
func (vr *visibleRange) normalizeToWidth(width int) {
	switch {
	case width < vr.cells():
		diff := vr.cells() - width
		vr.startIdx += diff

	case width > vr.cells():
		diff := width - vr.cells()
		vr.startIdx -= diff
	}

	if vr.startIdx < 0 {
		vr.endIdx += -1 * vr.startIdx
		vr.startIdx = 0
	}
}

// normalizeToiData normalizes the visible range, handles cases where the
// length of the data decreased due to deletion of some runes.
func (vr *visibleRange) normalizeToData(dataLen int) {
	if dataLen >= vr.endIdx || vr.startIdx == 0 {
		// Nothing to do when data is longer than the range or the range
		// already starts all the way left.
		return
	}

	diff := vr.endIdx - dataLen
	if diff == 1 {
		// The data can be one character shorter than the visible range, since
		// this is space for the cursor when appending.
		return
	}
	diff--

	_, newStartIdx := numbers.MinMaxInts([]int{vr.startIdx - diff, 0})
	shift := vr.startIdx - newStartIdx
	vr.endIdx -= shift
	vr.startIdx = newStartIdx
}

// curRelative returns the relative position of the cursor within the visible
// range. Returns an error if the cursos isn't inside the visible range.
func (vr *visibleRange) curRelative(curDataPos int) (int, error) {
	if !vr.contains(curDataPos) {
		return 0, fmt.Errorf("curDataPos %d isn't inside %#v", curDataPos, *vr)
	}
	return curDataPos - vr.startIdx, nil
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

	// visible is the currently visible range.
	visible *visibleRange
}

// newFieldEditor returns a new fieldEditor instance.
func newFieldEditor() *fieldEditor {
	return &fieldEditor{
		visible: &visibleRange{},
	}
}

// shiftLeft shifts the visible range left so that it again contains the
// cursor.
func (fe *fieldEditor) shiftLeft() {
	var startIdx int
	switch {
	case fe.curDataPos == 0 || fe.visible.cells() < minForArrows:
		startIdx = fe.curDataPos

	default:
		startIdx = fe.curDataPos - 1
	}
	endIdx := fe.data.cellsAfter(fe.visible.forRunes(), startIdx)
	endIdx++ // Space for the cursor.

	gotCells := endIdx - startIdx
	if fe.visible.cells() >= minForArrows && gotCells < minForArrows {
		// The plan was to hide the first rune under an arrow.
		// However after looking at the actual runes in the range, some took
		// more space than one cell (full-width runes) and we have lost the
		// space for the arrow, so shift the range by one.
		startIdx++
		endIdx++
	}
	fe.visible.set(startIdx, endIdx)
}

// shiftRight shifts the visible range right so that it again contains the
// cursor.
func (fe *fieldEditor) shiftRight() {
	var endIdx int
	switch dataLen := len(fe.data); {
	case fe.curDataPos == dataLen:
		// Cursor is in the empty space after the data.
		// Print all runes until the end of data.
		endIdx = dataLen

	default:
		// Cursor is within the data, print all runes including the one the
		// cursor is on.
		endIdx = fe.curDataPos + 1
	}

	startIdx := fe.data.cellsBefore(fe.visible.forRunes(), endIdx)
	endIdx++ // Space for the cursor within the visible range.
	fe.visible.set(startIdx, endIdx)
}

// toCursor shifts the visible range to the cursor if it scrolled out of view.
// This is a no-op if the cursor is inside the range.
func (fe *fieldEditor) toCursor() {
	switch {
	case fe.curDataPos < fe.visible.curMinIdx():
		fe.shiftLeft()
	case fe.curDataPos > fe.visible.curMaxIdx(len(fe.data)):
		fe.shiftRight()
	}
}

// viewFor returns the currently visible data inside a text field with the
// specified width and the cursor position within the field.
func (fe *fieldEditor) viewFor(width int) (string, int, error) {
	if min := 4; width < min { // One for left arrow, two for one full-width rune and one for the cursor.
		return "", -1, fmt.Errorf("width %d is too small, the minimum is %d", width, min)
	}
	fe.visible.normalizeToWidth(width)
	fe.visible.normalizeToData(len(fe.data))
	fe.toCursor()

	cur, err := fe.visible.curRelative(fe.curDataPos)
	if err != nil {
		return "", 0, err
	}
	return fe.data.runesIn(fe.visible), cur, nil
}

// insert inserts the rune at the current position of the cursor.
func (fe *fieldEditor) insert(r rune) {
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
