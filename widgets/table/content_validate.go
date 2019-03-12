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

// contant_validate.go contains code that validates the user provided Content.

import (
	"fmt"

	"github.com/mum4k/termdash/internal/numbers"
)

// validateContent validates the content instance.
func validateContent(content *Content) error {
	if got, min := int(content.cols), 1; got < min {
		return fmt.Errorf("invalid number of columns %d, must be a value in range %d <= v", got, min)
	}

	if got := len(content.opts.columnWidthsPercent); got > 0 {
		if want := int(content.cols); got != want {
			return fmt.Errorf("invalid number of widths in ColumnWidthsPercent %d, must be equal to the number of columns %d", got, want)
		}
		if sum, want := numbers.SumInts(content.opts.columnWidthsPercent), 100; sum != want {
			return fmt.Errorf("invalid sum of widths in ColumnWidthsPercent %d, must be %d", sum, want)
		}

	}

	for _, row := range content.rows {
		for _, c := range row.cells {
			if got, min := c.colSpan, 1; got < min {
				return fmt.Errorf("invalid CellColSpan %d, must be a value in range %d <= v", got, min)
			}
			if got, min := c.rowSpan, 1; got < min {
				return fmt.Errorf("invalid CellRowSpan %d, must be a value in range %d <= v", got, min)
			}
		}

		if got, want := row.effectiveColumns(), int(content.cols); got != want {
			return fmt.Errorf("content has %d columns, but row %v has %d, all rows must occupy the same amount of columns", want, row, got)
		}
	}
	return nil
}
