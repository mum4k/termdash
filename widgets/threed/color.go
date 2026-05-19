// threed/color.go

package threed

import (
	"github.com/mum4k/termdash/cell"
)

var (
	// NeonCyan is a crisp technical accent for wireframes and boards.
	NeonCyan = Color{R: 0.50, G: 0.92, B: 0.96}
	// NeonGreen is a bright signal color for nodes and highlights.
	NeonGreen = Color{R: 0.52, G: 0.98, B: 0.38}
	// Amber is a warm graph and warning accent.
	Amber = Color{R: 0.98, G: 0.78, B: 0.28}
	// Rose is a saturated block/module accent.
	Rose = Color{R: 0.96, G: 0.36, B: 0.52}
	// SoftWhite is readable foreground on dark terminals.
	SoftWhite = Color{R: 0.86, G: 0.88, B: 0.90}
)

// Color represents an RGB color with values between 0 and 1.
type Color struct {
	R float64 // Red component (0.0 - 1.0)
	G float64 // Green component (0.0 - 1.0)
	B float64 // Blue component (0.0 - 1.0)
}

// RGB converts 8-bit RGB channel values to a ThreeD color.
func RGB(r, g, b uint8) Color {
	return Color{
		R: float64(r) / 255,
		G: float64(g) / 255,
		B: float64(b) / 255,
	}
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
