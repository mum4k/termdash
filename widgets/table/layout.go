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

/*
TODO:
- given cvsWidth, determine space available to cells (minus border and padding).
- columnWidths to account for padding
- calculate column widths

- each row should report its required height - accounting for content of cells + padding
- process to draw row by row - top down and bottom up, account for when doesn't fit.

- func: given a cell - give me its area
- each cell - draw itself on a canvas, best effort on size.

- ensure content copies data given by the user to prevent races.
*/
