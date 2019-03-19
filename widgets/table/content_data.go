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

package table

// content_data.go defines a type that represents data within a table cell.

import (
	"bytes"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/canvas/buffer"
)

// DataOption is used to provide options to NewDataWithOpts.
type DataOption interface {
	// set sets the provided option.
	set(*dataOptions)
}

// dataOptions stores the provided data options.
type dataOptions struct {
	cellOpts []cell.Option
}

// dataOption implements DataOption.
type dataOption func(*dataOptions)

// set implements Option.set.
func (do dataOption) set(dOpts *dataOptions) {
	do(dOpts)
}

// DataCellOpts sets options on the cells that contain the data.
func DataCellOpts(cellOpts ...cell.Option) DataOption {
	return dataOption(func(dOpts *dataOptions) {
		dOpts.cellOpts = cellOpts
	})
}

// Data is part of (or the full) the data that is displayed inside one Cell.
type Data struct {
	// cells contain the text and its cell options.
	cells []*buffer.Cell
}

// String implements fmt.Stringer.
func (d *Data) String() string {
	var b bytes.Buffer
	for _, c := range d.cells {
		b.WriteRune(c.Rune)
	}
	return b.String()
}

// NewData creates new Data with the provided text and applies the options.
// The text contain cannot control characters (unicode.IsControl) or space
// character (unicode.IsSpace) other than:
//   ' ', '\n'
// Any newline ('\n') characters are interpreted as newlines when displaying
// the text.
func NewData(text string, opts ...DataOption) *Data {
	dOpts := &dataOptions{}
	for _, opt := range opts {
		opt.set(dOpts)
	}

	var cells []*buffer.Cell
	for _, r := range text {
		cells = append(cells, buffer.NewCell(r, dOpts.cellOpts...))
	}
	return &Data{
		cells: cells,
	}
}

// newCombinedData returns a Data instance that combines cells from all the
// data instances passed in.
func newCombinedData(data []*Data) *Data {
	res := &Data{}
	for _, d := range data {
		res.cells = append(res.cells, d.cells...)
	}
	return res
}
