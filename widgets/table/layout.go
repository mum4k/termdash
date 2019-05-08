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

// layout.go stores layout calculated for a canvas size.

import (
	"image"

	"github.com/mum4k/termdash/internal/numbers"
)

// contentLayout determines how the content gets placed onto the canvas.
type contentLayout struct {
	// lastCvsAr is the are of the last canvas the content was drawn on.
	// This is image.ZR if the content hasn't been drawn yet.
	lastCvsAr image.Rectangle

	// columnWidths are the widths of individual columns in the table.
	columnWidths []columnWidth

	// TODO: Details about HV lines that are the borders.
}

// newContentLayout calculates new layout for the content when drawn on a
// canvas represented with the provided area.
func newContentLayout(content *Content, cvsAr image.Rectangle) (*contentLayout, error) {
	// TODO: Space for border and horizontal padding / spacing.
	cvsWidth := cvsAr.Dx()
	colWidths, err := columnWidths(content, cvsWidth)
	if err != nil {
		return nil, err
	}

	// Wrap content in all cells that have wrapping configured.
	for _, tRow := range content.rows {
		for colIdx, tCell := range tRow.cells {
			// TODO: Account for colSpan.
			cw := colWidths[colIdx]
			if err := tCell.wrapToWidth(cw); err != nil {
				return nil, err
			}
		}
	}

	return &contentLayout{
		lastCvsAr:    cvsAr,
		columnWidths: colWidths,
	}, nil
}

// cellUsable determines the usable width for the specified cell given the
// column widths. This accounts for additional space requirements when border
// or padding is used. Can return zero if no width is available for the cell.
func cellUsableWidth(content *Content, cell *Cell, cw columnWidth) int {
	sub := 0
	if content.hasBorder() {
		// Reserve one terminal cell per table cell for the left border.
		sub++
	}
	sub += 2 * cell.hierarchical.getHorizontalPadding()

	_, usable := numbers.MinMaxInts([]int{0, int(cw) - sub})
	return usable
}
