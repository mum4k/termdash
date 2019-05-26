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

// Package grid helps to build grid layouts.
package grid

import (
	"fmt"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/widgetapi"
)

// Builder builds grid layouts.
type Builder struct {
	elems []Element
}

// New returns a new grid builder.
func New() *Builder {
	return &Builder{}
}

// Add adds the specified elements.
// The subElements can be either a single Widget or any combination of Rows and
// Columns.
// Rows are created using functions with the RowHeight prefix and Columns are
// created using functions with the ColWidth prefix
// Can be called repeatedly, e.g. to add multiple Rows or Columns.
func (b *Builder) Add(subElements ...Element) {
	b.elems = append(b.elems, subElements...)
}

// Build builds the grid layout and returns the corresponding container
// options.
func (b *Builder) Build() ([]container.Option, error) {
	if err := validate(b.elems /* fixedSizeParent = */, false); err != nil {
		return nil, err
	}
	return build(b.elems, 100, 100), nil
}

// validate recursively validates the elements that were added to the builder.
// Validates the following per each level of Rows or Columns.:
//   The subElements are either exactly one Widget or any number of Rows and
//   Columns.
//   Each individual width or height is in the range 0 < v < 100.
//   The sum of all widths is <= 100.
//   The sum of all heights is <= 100.
// Argument fixedSizeParent indicates if any of the parent elements uses fixed
// size splitType.
func validate(elems []Element, fixedSizeParent bool) error {
	heightPercSum := 0
	widthPercSum := 0
	for _, elem := range elems {
		switch e := elem.(type) {
		case *row:
			if e.splitType == splitTypeRelative {
				if min, max := 0, 100; e.heightPerc <= min || e.heightPerc >= max {
					return fmt.Errorf("invalid row %v, must be a value in the range %d < v < %d", e, min, max)
				}
			}
			heightPercSum += e.heightPerc

			if fixedSizeParent && e.splitType == splitTypeRelative {
				return fmt.Errorf("row %v cannot use relative height when one of its parent elements uses fixed height", e)
			}

			isFixed := fixedSizeParent || e.splitType == splitTypeFixed
			if err := validate(e.subElem, isFixed); err != nil {
				return err
			}

		case *col:
			if e.splitType == splitTypeRelative {
				if min, max := 0, 100; e.widthPerc <= min || e.widthPerc >= max {
					return fmt.Errorf("invalid column %v, must be a value in the range %d < v < %d", e, min, max)
				}
			}
			widthPercSum += e.widthPerc

			if fixedSizeParent && e.splitType == splitTypeRelative {
				return fmt.Errorf("column %v cannot use relative width when one of its parent elements uses fixed height", e)
			}

			isFixed := fixedSizeParent || e.splitType == splitTypeFixed
			if err := validate(e.subElem, isFixed); err != nil {
				return err
			}

		case *widget:
			if len(elems) > 1 {
				return fmt.Errorf("when adding a widget, it must be the only added element at that level, got: %v", elems)
			}
		}
	}

	if max := 100; heightPercSum > max || widthPercSum > max {
		return fmt.Errorf("the sum of all height percentages(%d) and width percentages(%d) at one element level cannot be larger than %d", heightPercSum, widthPercSum, max)
	}
	return nil
}

// build recursively builds the container options according to the elements
// that were added to the builder.
// The parentHeightPerc and parentWidthPerc percent indicate the relative size
// of the element we are building now in the parent element. See innerPerc()
// for more details.
func build(elems []Element, parentHeightPerc, parentWidthPerc int) []container.Option {
	if len(elems) == 0 {
		return nil
	}

	elem := elems[0]
	elems = elems[1:]

	switch e := elem.(type) {
	case *row:
		if len(elems) > 0 {
			perc := innerPerc(e.heightPerc, parentHeightPerc)
			childHeightPerc := parentHeightPerc - e.heightPerc

			var splitOpts []container.SplitOption
			if e.splitType == splitTypeRelative {
				splitOpts = append(splitOpts, container.SplitPercent(perc))
			} else {
				splitOpts = append(splitOpts, container.SplitFixed(e.heightFixed))
			}

			return []container.Option{
				container.SplitHorizontal(
					container.Top(append(e.cOpts, build(e.subElem, 100, parentWidthPerc)...)...),
					container.Bottom(build(elems, childHeightPerc, parentWidthPerc)...),
					splitOpts...,
				),
			}
		}
		return append(e.cOpts, build(e.subElem, 100, parentWidthPerc)...)

	case *col:
		if len(elems) > 0 {
			perc := innerPerc(e.widthPerc, parentWidthPerc)
			childWidthPerc := parentWidthPerc - e.widthPerc

			var splitOpts []container.SplitOption
			if e.splitType == splitTypeRelative {
				splitOpts = append(splitOpts, container.SplitPercent(perc))
			} else {
				splitOpts = append(splitOpts, container.SplitFixed(e.widthFixed))
			}

			return []container.Option{
				container.SplitVertical(
					container.Left(append(e.cOpts, build(e.subElem, parentHeightPerc, 100)...)...),
					container.Right(build(elems, parentHeightPerc, childWidthPerc)...),
					splitOpts...,
				),
			}
		}
		return append(e.cOpts, build(e.subElem, parentHeightPerc, 100)...)

	case *widget:
		opts := e.cOpts
		opts = append(opts, container.PlaceWidget(e.widget))
		return opts
	}
	return nil
}

// innerPerc translates the outer split percentage into the inner one.
// E.g. multiple rows would specify that they want the outer split percentage
// of 25% each, but we are representing them in a tree of containers so the
// inner splits vary:
//     ╭─────────╮
// 25% │   25%   │
//     │╭───────╮│ ---
// 25% ││  33%  ││
//     ││╭─────╮││
// 25% │││ 50% │││
//     ││├─────┤││ 75%
// 25% │││ 50% │││
//     ││╰─────╯││
//     │╰───────╯│
//     ╰─────────╯ ---
//
// Argument outerPerc is the user specified percentage for the split, i.e. the
// 25% in the example above.
// Argument parentPerc is the percentage this container has in the parent, i.e.
// 75% for the first inner container in the example above.
func innerPerc(outerPerc, parentPerc int) int {
	// parentPerc * parentHeightCells = childHeightCells
	// innerPerc * childHeightCells = outerPerc * parentHeightCells
	// innerPerc * parentPerc * parentHeightCells = outerPerc * parentHeightCells
	// innerPerc * parentPerc = outerPerc
	// innerPerc = outerPerc / parentPerc
	return int(float64(outerPerc) / float64(parentPerc) * 100)
}

// Element is an element that can be added to the grid.
type Element interface {
	isElement()
}

// splitType represents
type splitType int

// String implements fmt.Stringer()
func (st splitType) String() string {
	if n, ok := splitTypeNames[st]; ok {
		return n
	}
	return "splitTypeUnknown"
}

// splitTypeNames maps splitType values to human readable names.
var splitTypeNames = map[splitType]string{
	splitTypeRelative: "splitTypeRelative",
	splitTypeFixed:    "splitTypeFixed",
}

const (
	splitTypeRelative splitType = iota
	splitTypeFixed
)

// row is a row in the grid.
// row implements Element.
type row struct {
	// splitType identifies how the size of the split is determined.
	splitType splitType

	// heightPerc is the height percentage this row occupies.
	// Only set when splitType is splitTypeRelative.
	heightPerc int

	// heightFixed is the height in cells this row occupies.
	// Only set when splitType is splitTypeFixed.
	heightFixed int

	// subElem are the sub Rows or Columns or a single widget.
	subElem []Element

	// cOpts are the options for the row's container.
	cOpts []container.Option
}

// isElement implements Element.isElement.
func (row) isElement() {}

// String implements fmt.Stringer.
func (r *row) String() string {
	return fmt.Sprintf("row{splitType:%v, heightPerc:%d, heightFixed:%d, sub:%v}", r.splitType, r.heightPerc, r.heightFixed, r.subElem)
}

// col is a column in the grid.
// col implements Element.
type col struct {
	// splitType identifies how the size of the split is determined.
	splitType splitType

	// widthPerc is the width percentage this column occupies.
	// Only set when splitType is splitTypeRelative.
	widthPerc int

	// widthFixed is the width in cells thiw column occupies.
	// Only set when splitType is splitTypeRelative.
	widthFixed int

	// subElem are the sub Rows or Columns or a single widget.
	subElem []Element

	// cOpts are the options for the column's container.
	cOpts []container.Option
}

// isElement implements Element.isElement.
func (col) isElement() {}

// String implements fmt.Stringer.
func (c *col) String() string {
	return fmt.Sprintf("col{splitType:%v, widthPerc:%d, widthFixed:%d, sub:%v}", c.splitType, c.widthPerc, c.widthFixed, c.subElem)
}

// widget is a widget placed into the grid.
// widget implements Element.
type widget struct {
	// widget is the widget instance.
	widget widgetapi.Widget
	// cOpts are the options for the widget's container.
	cOpts []container.Option
}

// String implements fmt.Stringer.
func (w *widget) String() string {
	return fmt.Sprintf("widget{type:%T}", w.widget)
}

// isElement implements Element.isElement.
func (widget) isElement() {}

// RowHeightPerc creates a row of the specified relative height.
// The height is supplied as height percentage of the parent element.
// The sum of all heights at the same level cannot be larger than 100%. If it
// is less that 100%, the last element stretches to the edge of the screen.
// The subElements can be either a single Widget or any combination of Rows and
// Columns.
func RowHeightPerc(heightPerc int, subElements ...Element) Element {
	return &row{
		splitType:  splitTypeRelative,
		heightPerc: heightPerc,
		subElem:    subElements,
	}
}

// RowHeightFixed creates a row of the specified fixed height.
// The height is supplied as a number of cells on the terminal.
// If the actual terminal size leaves the container with less than the
// specified amount of cells, the container will be created with zero cells and
// won't be drawn until the terminal size increases. If the sum of all the
// heights is less than 100% of the screen height, the last element stretches
// to the edge of the screen.
// The subElements can be either a single Widget or any combination of Rows and
// Columns.
// A row with fixed height cannot contain any sub-elements with relative size.
func RowHeightFixed(heightCells int, subElements ...Element) Element {
	return &row{
		splitType:   splitTypeFixed,
		heightFixed: heightCells,
		subElem:     subElements,
	}
}

// RowHeightPercWithOpts is like RowHeightPerc, but also allows to apply
// additional options to the container that represents the row.
func RowHeightPercWithOpts(heightPerc int, cOpts []container.Option, subElements ...Element) Element {
	return &row{
		splitType:  splitTypeRelative,
		heightPerc: heightPerc,
		subElem:    subElements,
		cOpts:      cOpts,
	}
}

// RowHeightFixedWithOpts is like RowHeightFixed, but also allows to apply
// additional options to the container that represents the row.
func RowHeightFixedWithOpts(heightCells int, cOpts []container.Option, subElements ...Element) Element {
	return &row{
		splitType:   splitTypeFixed,
		heightFixed: heightCells,
		subElem:     subElements,
		cOpts:       cOpts,
	}
}

// ColWidthPerc creates a column of the specified relative width.
// The width is supplied as width percentage of the parent element.
// The sum of all widths at the same level cannot be larger than 100%. If it
// is less that 100%, the last element stretches to the edge of the screen.
// The subElements can be either a single Widget or any combination of Rows and
// Columns.
func ColWidthPerc(widthPerc int, subElements ...Element) Element {
	return &col{
		splitType: splitTypeRelative,
		widthPerc: widthPerc,
		subElem:   subElements,
	}
}

// ColWidthFixed creates a column of the specified fixed width.
// The width is supplied as a number of cells on the terminal.
// If the actual terminal size leaves the container with less than the
// specified amount of cells, the container will be created with zero cells and
// won't be drawn until the terminal size increases. If the sum of all the
// widths is less than 100% of the screen width, the last element stretches
// to the edge of the screen.
// The subElements can be either a single Widget or any combination of Rows and
// Columns.
// A column with fixed width cannot contain any sub-elements with relative size.
func ColWidthFixed(widthCells int, subElements ...Element) Element {
	return &col{
		splitType:  splitTypeFixed,
		widthFixed: widthCells,
		subElem:    subElements,
	}
}

// ColWidthPercWithOpts is like ColWidthPerc, but also allows to apply
// additional options to the container that represents the column.
func ColWidthPercWithOpts(widthPerc int, cOpts []container.Option, subElements ...Element) Element {
	return &col{
		splitType: splitTypeRelative,
		widthPerc: widthPerc,
		subElem:   subElements,
		cOpts:     cOpts,
	}
}

// ColWidthFixedWithOpts is like ColWidthFixed, but also allows to apply
// additional options to the container that represents the column.
func ColWidthFixedWithOpts(widthCells int, cOpts []container.Option, subElements ...Element) Element {
	return &col{
		splitType:  splitTypeFixed,
		widthFixed: widthCells,
		subElem:    subElements,
		cOpts:      cOpts,
	}
}

// Widget adds a widget into the Row or Column.
// The options will be applied to the container that directly holds this
// widget.
func Widget(w widgetapi.Widget, cOpts ...container.Option) Element {
	return &widget{
		widget: w,
		cOpts:  cOpts,
	}
}
