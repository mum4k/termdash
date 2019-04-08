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
