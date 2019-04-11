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
	"log"

	"github.com/mum4k/termdash/internal/numbers"
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

// width is the rune width of all the runes in the data.
func (fd *fieldData) width() int {
	return fd.rangeWidth(0, len(*fd))
}

// runesIn returns runes that are in the visible range.
func (fd *fieldData) runesIn(vr *visibleRange) string {
	var runes []rune
	for i, r := range (*fd)[vr.startIdx:] {
		if i+vr.startIdx >= vr.endIdx {
			break
		}
		runes = append(runes, r)
	}

	startVisible := vr.startIdx == 0
	endVisible := vr.endIdx >= len(*fd)
	useArrows := len(runes) > 2

	var b bytes.Buffer
	for i, r := range runes {
		switch {
		case useArrows && i == 0 && !startVisible:
			b.WriteRune('⇦')

		case useArrows && i == len(runes)-1 && !endVisible:
			b.WriteRune('⇨')

		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// visibleRange represents a range of currently visible runes.
// Visible runes are all such runes whose index falls within:
//   startIdx <= idx < endIdx
type visibleRange struct {
	startIdx int
	endIdx   int
}

// runeCount returns the number of visible runes.
func (vr *visibleRange) runeCount() int {
	return vr.endIdx - vr.startIdx
}

// setFromStart sets the visible range from the start of the data until the
// provided width.
func (vr *visibleRange) setFromStart(forRunes int) {
	vr.startIdx = 0
	vr.endIdx = forRunes
}

// set sets the visible range from the start to the end index.
func (vr *visibleRange) set(startIdx, endIdx int) {
	vr.startIdx = startIdx
	vr.endIdx = endIdx
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
	visible *visibleRange
}

// newFieldEditor returns a new fieldEditor instance.
func newFieldEditor() *fieldEditor {
	return &fieldEditor{
		visible: &visibleRange{},
	}
}

// viewFor returns the currently visible data inside a text field with the
// specified width and the cursor position within the field.
func (fe *fieldEditor) viewFor(width int) (string, int, error) {
	if min := 4; width < min { // One for left arrow, two for one full-width rune and one for the cursor.
		return "", -1, fmt.Errorf("width %d is too small, the minimum is %d", width, min)
	}
	forRunes := width - 1 // One reserved for the cursor.

	if fe.data.width() <= forRunes {
		log.Printf("Case1(all visible)")
		// Base case, all runes fit into the width.
		fe.visible.setFromStart(forRunes)
		return fe.data.runesIn(fe.visible), fe.curPos, nil
	}

	if fe.visible.runeCount() > forRunes {
		log.Printf("Case2(shrinking visible)")
		fe.visible.endIdx = fe.visible.startIdx + forRunes
	}

	log.Printf("fe.curPos:%d fe.visible:%#v", fe.curPos, fe.visible)
	if fe.curPos > fe.visible.endIdx {
		log.Printf("Case3(shifting right)")
		endIdx := fe.curPos
		startIdx := fe.data.cellsBefore(forRunes, endIdx)
		fe.visible.set(startIdx, endIdx)
		width := fe.data.rangeWidth(startIdx, endIdx)

		curPos := 0
		if width == forRunes {
			curPos = fe.visible.endIdx - fe.visible.startIdx
		} else {
			diff := forRunes - width
			curPos = fe.visible.endIdx - fe.visible.startIdx - diff
		}
		return fe.data.runesIn(fe.visible), curPos, nil
	}

	if fe.curPos < fe.visible.startIdx {
		log.Printf("Case3(shifting left)")
		startIdx := fe.curPos
		endIdx := fe.data.cellsAfter(forRunes, startIdx)
		fe.visible.set(startIdx, endIdx)

		curPos := 0
		return fe.data.runesIn(fe.visible), curPos, nil
	}

	// Case - the cursor is in.
	log.Printf("Case3(cursor is in)")
	curPos := fe.curPos - fe.visible.startIdx
	return fe.data.runesIn(fe.visible), curPos, nil
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
	fe.curPos, _ = numbers.MinMaxInts([]int{fe.curPos + 1, len(fe.data)})
}

// cursorLeft moves the cursor one position to the left.
func (fe *fieldEditor) cursorLeft() {
	_, fe.curPos = numbers.MinMaxInts([]int{fe.curPos - 1, 0})
}

// cursorHome moves the cursor to the beginning of the data.
func (fe *fieldEditor) cursorHome() {}

// cursorEnd moves the cursor to the end of the data.
func (fe *fieldEditor) cursorEnd() {}
