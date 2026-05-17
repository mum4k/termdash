// threed/color.go

package threed

import (
	"github.com/mum4k/termdash/cell"
)

// Color represents an RGB color with values between 0 and 1.
type Color struct {
	R float64 // Red component (0.0 - 1.0)
	G float64 // Green component (0.0 - 1.0)
	B float64 // Blue component (0.0 - 1.0)
}

// Multiply multiplies the color by a scalar.
func (c Color) Multiply(factor float64) Color {
	return Color{
		R: c.R * factor,
		G: c.G * factor,
		B: c.B * factor,
	}
}

// Add adds another color to this color.
func (c Color) Add(other Color) Color {
	return Color{
		R: c.R + other.R,
		G: c.G + other.G,
		B: c.B + other.B,
	}
}

// Modulate multiplies this color by another color channel-by-channel.
func (c Color) Modulate(other Color) Color {
	return Color{
		R: c.R * other.R,
		G: c.G * other.G,
		B: c.B * other.B,
	}
}

// ToCellColor converts the Color to a cell.Color.
func (c Color) ToCellColor() cell.Color {
	r := int(clampFloat(c.R*255, 0, 255))
	g := int(clampFloat(c.G*255, 0, 255))
	b := int(clampFloat(c.B*255, 0, 255))
	return cell.ColorRGB24(r, g, b)
}
