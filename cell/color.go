package cell

// color.go defines constants for cell colors.

// Color is the color of a cell.
type Color int

// String implements fmt.Stringer()
func (cc Color) String() string {
	if n, ok := colorNames[cc]; ok {
		return n
	}
	return "ColorUnknown"
}

// colorNames maps Color values to human readable names.
var colorNames = map[Color]string{
	ColorDefault: "ColorDefault",
	ColorBlack:   "ColorBlack",
	ColorRed:     "ColorRed",
	ColorGreen:   "ColorGreen",
	ColorYellow:  "ColorYellow",
	ColorBlue:    "ColorBlue",
	ColorMagenta: "ColorMagenta",
	ColorCyan:    "ColorCyan",
	ColorWhite:   "ColorWhite",
}

const (
	ColorDefault Color = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)
