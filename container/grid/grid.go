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
// Rows are created using RowHeightPerc() and Columns are created using
// ColWidthPerc().
// Can be called repeatedly, e.g. to add multiple Rows or Columns.
func (b *Builder) Add(subElements ...Element) {
	b.elems = append(b.elems, subElements...)
}

// Build builds the grid layout and returns the corresponding container
// options.
func (b *Builder) Build() ([]container.Option, error) {
	if err := validate(b.elems); err != nil {
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
func validate(elems []Element) error {
	heightSum := 0
	widthSum := 0
	for _, elem := range elems {
		switch e := elem.(type) {
		case *row:
			if min, max := 0, 100; e.heightPerc <= min || e.heightPerc >= max {
				return fmt.Errorf("invalid row heightPerc(%d), must be a value in the range %d < v < %d", e.heightPerc, min, max)
			}
			heightSum += e.heightPerc
			if err := validate(e.subElem); err != nil {
				return err
			}

		case *col:
			if min, max := 0, 100; e.widthPerc <= min || e.widthPerc >= max {
				return fmt.Errorf("invalid column widthPerc(%d), must be a value in the range %d < v < %d", e.widthPerc, min, max)
			}
			widthSum += e.widthPerc
			if err := validate(e.subElem); err != nil {
				return err
			}

		case *widget:
			if len(elems) > 1 {
				return fmt.Errorf("when adding a widget, it must be the only added element at that level, got: %v", elems)
			}
		}
	}

	if max := 100; heightSum > max || widthSum > max {
		return fmt.Errorf("the sum of all height percentages(%d) and width percentages(%d) at one element level cannot be larger than %d", heightSum, widthSum, max)
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
			return []container.Option{
				container.SplitHorizontal(
					container.Top(build(e.subElem, 100, parentWidthPerc)...),
					container.Bottom(build(elems, childHeightPerc, parentWidthPerc)...),
					container.SplitPercent(perc),
				),
			}
		}
		return build(e.subElem, 100, parentWidthPerc)

	case *col:
		if len(elems) > 0 {
			perc := innerPerc(e.widthPerc, parentWidthPerc)
			childWidthPerc := parentWidthPerc - e.widthPerc
			return []container.Option{
				container.SplitVertical(
					container.Left(build(e.subElem, parentHeightPerc, 100)...),
					container.Right(build(elems, parentHeightPerc, childWidthPerc)...),
					container.SplitPercent(perc),
				),
			}
		}
		return build(e.subElem, parentHeightPerc, 100)

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

// row is a row in the grid.
// row implements Element.
type row struct {
	// heightPerc is the height percentage this row occupies.
	heightPerc int

	// subElem are the sub Rows or Columns or a single widget.
	subElem []Element
}

// isElement implements Element.isElement.
func (row) isElement() {}

// String implements fmt.Stringer.
func (r *row) String() string {
	return fmt.Sprintf("row{height:%d, sub:%v}", r.heightPerc, r.subElem)
}

// col is a column in the grid.
// col implements Element.
type col struct {
	// widthPerc is the width percentage this column occupies.
	widthPerc int

	// subElem are the sub Rows or Columns or a single widget.
	subElem []Element
}

// isElement implements Element.isElement.
func (col) isElement() {}

// String implements fmt.Stringer.
func (c *col) String() string {
	return fmt.Sprintf("col{width:%d, sub:%v}", c.widthPerc, c.subElem)
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

// RowHeightPerc creates a row of the specified height.
// The height is supplied as height percentage of the parent element.
// The sum of all heights at the same level cannot be larger than 100%. If it
// is less that 100%, the last element stretches to the edge of the screen.
// The subElements can be either a single Widget or any combination of Rows and
// Columns.
func RowHeightPerc(heightPerc int, subElements ...Element) Element {
	return &row{
		heightPerc: heightPerc,
		subElem:    subElements,
	}
}

// ColWidthPerc creates a column of the specified width.
// The width is supplied as width percentage of the parent element.
// The sum of all widths at the same level cannot be larger than 100%. If it
// is less that 100%, the last element stretches to the edge of the screen.
// The subElements can be either a single Widget or any combination of Rows and
// Columns.
func ColWidthPerc(widthPerc int, subElements ...Element) Element {
	return &col{
		widthPerc: widthPerc,
		subElem:   subElements,
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
