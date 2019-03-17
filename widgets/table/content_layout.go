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

// content_layout.go stores layout calculated for a canvas size.

import (
	"errors"
	"image"
	"log"
	"math"

	"github.com/mum4k/termdash/internal/wrap"
)

// columnWidth is the width of a column in cells.
// This excludes any border, padding or spacing, i.e. this is the data portion
// only.
type columnWidth int

// contentLayout determines how the content gets placed onto the canvas.
type contentLayout struct {
	// lastCvsAr is the are of the last canvas the content was drawn on.
	// This is image.ZR if the content hasn't been drawn yet.
	lastCvsAr image.Rectangle

	// columnWidths are the widths of individual columns in the table.
	columnWidths []columnWidth

	// Details about HV lines that are the borders.
}

// newContentLayout calculates new layout for the content when drawn on a
// canvas represented with the provided area.
func newContentLayout(content *Content, cvsAr image.Rectangle) (*contentLayout, error) {
	return nil, errors.New("unimplemented")
}

// columnWidths given the content and the available canvas width returns the
// widths of individual columns.
// The argument cvsWidth is assumed to exclude space required for any border,
// padding or spacing.
func columnWidths(content *Content, cvsWidth int) []columnWidth {
	// This is similar to the rod-cutting problem, except instead of maximizing
	// the price, we're minimizing the number of rows that would have their
	// content trimmed.

	idxColumnCosts := columnCosts(content, cvsWidth)
	log.Printf("idxColumnCosts: %v", idxColumnCosts)
	minCost, minCuts := cutCanvas(idxColumnCosts, cvsWidth, cvsWidth, int(content.cols), 0, nil)
	log.Printf("minCost: %v", minCost)
	log.Printf("minCuts: %v", minCuts)

	var res []columnWidth
	last := 0
	for _, cut := range minCuts {
		res = append(res, columnWidth(cut-last))
		last = cut
	}
	res = append(res, columnWidth(cvsWidth-last))
	return res
}

func cutCanvas(idxColumnCosts map[int]widthCost, cvsWidth, remWidth, columns, colIdx int, cuts []int) (int, []int) {
	log.Printf("cutCanvas remWidth:%d, columns:%d, colIdx:%d, cuts:%v", remWidth, columns, colIdx, cuts)
	if remWidth <= 0 {
		log.Printf("  -> 0")
		return 0, cuts
	}

	minCost := math.MaxInt32
	var minCuts []int

	widthCosts := idxColumnCosts[colIdx]
	nextColIdx := colIdx + 1
	if nextColIdx > columns-1 {
		log.Printf("  -> no more cuts remWidth:%d cost:%d", remWidth, widthCosts[remWidth])
		return widthCosts[remWidth], cuts
	}

	for colWidth := 1; colWidth < remWidth; colWidth++ {
		diff := cvsWidth - remWidth
		idxThisCut := diff + colWidth
		costThisCut := widthCosts[colWidth]
		nextCost, nextCuts := cutCanvas(
			idxColumnCosts,
			cvsWidth,
			remWidth-colWidth,
			columns,
			nextColIdx,
			//append(cuts, colWidth+colIdx),
			append(cuts, idxThisCut),
		)

		if newMinCost := costThisCut + nextCost; newMinCost < minCost {
			log.Printf("at cuts %v, costThisCut from widthCosts:%v, at width %d:%d, nextCost:%d, minCost:%d", cuts, widthCosts, colWidth, costThisCut, nextCost, minCost)
			minCost = newMinCost
			minCuts = nextCuts
			log.Printf("new minCost:%d minCuts:%v", minCost, minCuts)
		}
	}
	log.Printf("cutCanvas remWidth:%d, columns:%d, colIdx:%d, cuts:%v -> minCost:%d, minCuts:%d", remWidth, columns, colIdx, cuts, minCost, minCuts)
	return minCost, minCuts
}

// widthCost maps column widths to the number of rows that would be trimmed on
// that width because data in the column are longer than the width.
type widthCost map[int]int

// columnCosts calculates the costs of cutting the columns to various widths.
// Returns a map of column indexes to their cut costs.
func columnCosts(content *Content, cvsWidth int) map[int]widthCost {
	idxColumnCosts := map[int]widthCost{}
	maxColWidth := cvsWidth - (int(content.cols) - 1)
	for colIdx := 0; colIdx < int(content.cols); colIdx++ {
		for colWidth := 1; colWidth <= maxColWidth; colWidth++ {
			wc, ok := idxColumnCosts[colIdx]
			if !ok {
				wc = widthCost{}
				idxColumnCosts[colIdx] = wc
			}
			wc[colWidth] = trimmedRows(content, colIdx, colWidth)
		}
	}
	return idxColumnCosts
}

// trimmedRows returns the number of rows that will have data cells with
// trimmed content in column of the specified index if the assigned width of
// the column is colWidth.
func trimmedRows(content *Content, colIdx int, colWidth int) int {
	trimmed := 0
	for _, row := range content.rows {
		tgtCell := row.cells[colIdx]
		if tgtCell.hierarchical.getWrapMode() != wrap.Never {
			// Cells that have wrapping enabled are never trimmed.
			continue
		}
		if tgtCell.width() > colWidth {
			trimmed++
		}
	}
	return trimmed
}
