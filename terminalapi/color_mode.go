package terminalapi

// color_mode.go defines the terminal color modes.

// ColorMode represents
type ColorMode int

// String implements fmt.Stringer()
func (cm ColorMode) String() string {
	if n, ok := colorModeNames[cm]; ok {
		return n
	}
	return "ColorModeUnknown"
}

// colorModeNames maps ColorMode values to human readable names.
var colorModeNames = map[ColorMode]string{
	ColorMode8:         "ColorMode8",
	ColorMode256:       "ColorMode256",
	ColorMode216:       "ColorMode216",
	ColorModeGrayscale: "ColorModeGrayscale",
}

const (
	ColorMode8 ColorMode = iota
	ColorMode256
	ColorMode216
	ColorModeGrayscale
)
