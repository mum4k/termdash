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

// fieldEditor maintains the cursor position and allows editing of the data in
// the text input field.
// This object isn't thread-safe.
type fieldEditor struct {
	// curPos is the current position of the cursor.
	curPos int

	// dataPos is the first visible rune. This is non-zero when there are more
	// runes than the width of the text input field and the data scroll to the
	// left.
	dataPos int

	// lastWidth is the width of the text input field when viewFor was called
	// last.
	lastWidth int

	// data are the data currently present in the text input field.
	data fieldData
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

	maxPos := width - 1
	if width < fe.lastWidth && fe.curPos > maxPos {
		// Indicates a terminal resize, normalize the cursor back into the text
		// input field.
		fe.curPos = maxPos
	}
	fe.lastWidth = width

	if fe.curPos > maxPos {
		fe.dataPos += fe.curPos - maxPos
		fe.curPos = maxPos
	}

	if len(fe.data) < width { // One reserved for the cursor.
		return string(fe.data[fe.dataPos:]), fe.curPos, nil
	}

	var b bytes.Buffer
	for i, r := range fe.data[fe.dataPos:] {
		if i == 0 {
			b.WriteRune('⇦')
			continue
		}

		if i >= maxPos {
			break
		}
		b.WriteRune(r)
	}
	return b.String(), fe.curPos, nil
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
func (fe *fieldEditor) cursorRight() {}

// cursoriLeft moves the cursor one position to the left.
func (fe *fieldEditor) cursorLeft() {}

// cursorHome moves the cursor to the beginning of the data.
func (fe *fieldEditor) cursorHome() {}

// cursorEnd moves the cursor to the end of the data.
func (fe *fieldEditor) cursorEnd() {}
