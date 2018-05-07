// Package align defines constants representing types of alignment.
package align

import (
	"fmt"
	"image"
)

// Horizontal indicates the type of horizontal alignment.
type Horizontal int

// String implements fmt.Stringer()
func (h Horizontal) String() string {
	if n, ok := horizontalNames[h]; ok {
		return n
	}
	return "HorizontalUnknown"
}

// horizontalNames maps Horizontal values to human readable names.
var horizontalNames = map[Horizontal]string{
	HorizontalLeft:   "HorizontalLeft",
	HorizontalCenter: "HorizontalCenter",
	HorizontalRight:  "HorizontalRight",
}

const (
	// HorizontalLeft is left alignment along the horizontal axis.
	HorizontalLeft Horizontal = iota
	// HorizontalCenter is center alignment along the horizontal axis.
	HorizontalCenter
	// HorizontalRight is right alignment along the horizontal axis.
	HorizontalRight
)

// Vertical indicates the type of vertical alignment.
type Vertical int

// String implements fmt.Stringer()
func (v Vertical) String() string {
	if n, ok := verticalNames[v]; ok {
		return n
	}
	return "VerticalUnknown"
}

// verticalNames maps Vertical values to human readable names.
var verticalNames = map[Vertical]string{
	VerticalTop:    "VerticalTop",
	VerticalMiddle: "VerticalMiddle",
	VerticalBottom: "VerticalBottom",
}

const (
	// VerticalTop is top alignment along the vertical axis.
	VerticalTop Vertical = iota
	// VerticalMiddle is middle alignment along the vertical axis.
	VerticalMiddle
	// VerticalBottom is bottom alignment along the vertical axis.
	VerticalBottom
)

// hAlign aligns the given area in the rectangle horizontally.
func hAlign(rect image.Rectangle, ar image.Rectangle, h Horizontal) (image.Rectangle, error) {
	gap := rect.Dx() - ar.Dx()
	switch h {
	case HorizontalRight:
		// Use gap from above.
	case HorizontalCenter:
		gap /= 2
	case HorizontalLeft:
		gap = 0
	default:
		return image.ZR, fmt.Errorf("unsupported horizontal alignment %v", h)
	}

	return image.Rect(
		rect.Min.X+gap,
		ar.Min.Y,
		rect.Min.X+gap+ar.Dx(),
		ar.Max.Y,
	), nil
}

// vAlign aligns the given area in the rectangle vertically.
func vAlign(rect image.Rectangle, ar image.Rectangle, v Vertical) (image.Rectangle, error) {
	gap := rect.Dy() - ar.Dy()
	switch v {
	case VerticalBottom:
		// Use gap from above.
	case VerticalMiddle:
		gap /= 2
	case VerticalTop:
		gap = 0
	default:
		return image.ZR, fmt.Errorf("unsupported vertical alignment %v", v)
	}

	return image.Rect(
		ar.Min.X,
		rect.Min.Y+gap,
		ar.Max.X,
		rect.Min.Y+gap+ar.Dy(),
	), nil
}

// Rectangle aligns the rectangle within the provided area returning the
// aligned area. The area must fall within the rectangle.
func Rectangle(rect image.Rectangle, ar image.Rectangle, h Horizontal, v Vertical) (image.Rectangle, error) {
	if !ar.In(rect) {
		return image.ZR, fmt.Errorf("cannot align area %v inside rectangle %v, the area falls outside of the rectangle", ar, rect)
	}

	aligned, err := hAlign(rect, ar, h)
	if err != nil {
		return image.ZR, err
	}
	aligned, err = vAlign(rect, aligned, v)
	if err != nil {
		return image.ZR, err
	}
	return aligned, nil
}
