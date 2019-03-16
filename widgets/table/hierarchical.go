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

// hierarchical.go contains options that inherit values from parents in the
// hierarchy.

import (
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/wrap"
)

// hierarchicalOptions stores options that can be applied at multiple levels or
// hierarchy, i.e. the Content (top level), the Row or the Cell.
type hierarchicalOptions struct {
	// parent are the hierarchical options on the parent.
	parent *hierarchicalOptions

	cellOpts          []cell.Option
	horizontalPadding *int
	verticalPadding   *int
	alignHorizontal   *align.Horizontal
	alignVertical     *align.Vertical
	height            *int
	wrapMode          *wrap.Mode
}

// getCellOpts returns the cell options.
// Either the value set at this level of hierarchy or a value set in one of the
// parents.
// Returns the default cell options if the cell options weren't set anywhere in
// the hierarchy.
func (ho *hierarchicalOptions) getCellOpts() *cell.Options {
	for ; ho != nil; ho = ho.parent {
		if len(ho.cellOpts) > 0 {
			return cell.NewOptions(ho.cellOpts...)
		}
	}
	return cell.NewOptions()
}

// getHorizontalPadding returns the horizontal cell padding.
// Either the value set at this level of hierarchy or a value set in one of the
// parents.
// Returns zero if the value wasn't set anywhere in the hierarchy.
func (ho *hierarchicalOptions) getHorizontalPadding() int {
	for ; ho != nil; ho = ho.parent {
		if ho.horizontalPadding != nil {
			return *ho.horizontalPadding
		}
	}
	return 0
}

// getVerticalPadding returns the vertical cell padding.
// Either the value set at this level of hierarchy or a value set in one of the
// parents.
// Returns zero if the value wasn't set anywhere in the hierarchy.
func (ho *hierarchicalOptions) getVerticalPadding() int {
	for ; ho != nil; ho = ho.parent {
		if ho.verticalPadding != nil {
			return *ho.verticalPadding
		}
	}
	return 0
}

// getAlignHorizontal returns the horizontal alignment.
// Either the value set at this level of hierarchy or a value set in one of the
// parents.
// Returns left horizontal alignment if the value wasn't set anywhere in the
// hierarchy.
func (ho *hierarchicalOptions) getAlignHorizontal() align.Horizontal {
	for ; ho != nil; ho = ho.parent {
		if ho.alignHorizontal != nil {
			return *ho.alignHorizontal
		}
	}
	return align.HorizontalLeft
}

// getAlignVertical returns the vertical alignment.
// Either the value set at this level of hierarchy or a value set in one of the
// parents.
// Returns top vertical alignment if the value wasn't set anywhere in the
// hierarchy.
func (ho *hierarchicalOptions) getAlignVertical() align.Vertical {
	for ; ho != nil; ho = ho.parent {
		if ho.alignVertical != nil {
			return *ho.alignVertical
		}
	}
	return align.VerticalTop
}

// getHeight returns the height.
// Either the value set at this level of hierarchy or a value set in one of the
// parents.
// Returns zero if the value wasn't set anywhere in the hierarchy, zero is
// interpreted as height adjusted to content.
func (ho *hierarchicalOptions) getHeight() int {
	for ; ho != nil; ho = ho.parent {
		if ho.height != nil {
			return *ho.height
		}
	}
	return 0
}

// getWrapMode returns the wrap mode.
// Either the value set at this level of hierarchy or a value set in one of the
// parents.
// Returns wrap.Never if the value wasn't set anywhere in the hierarchy.
func (ho *hierarchicalOptions) getWrapMode() wrap.Mode {
	for ; ho != nil; ho = ho.parent {
		if ho.wrapMode != nil {
			return *ho.wrapMode
		}
	}
	return wrap.Never
}
