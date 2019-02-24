// Package linestyle defines various line styles.
package linestyle

// LineStyle defines the supported line styles.Q
type LineStyle int

// String implements fmt.Stringer()
func (ls LineStyle) String() string {
	if n, ok := lineStyleNames[ls]; ok {
		return n
	}
	return "LineStyleUnknown"
}

// lineStyleNames maps LineStyle values to human readable names.
var lineStyleNames = map[LineStyle]string{
	None:   "LineStyleNone",
	Light:  "LineStyleLight",
	Double: "LineStyleDouble",
	Round:  "LineStyleRound",
}

// Supported line styles.
// See https://en.wikipedia.org/wiki/Box-drawing_character.
const (
	// None indicates that no line should be present.
	None LineStyle = iota

	// Light is line style using the '─' characters.
	Light

	// Double is line style using the '═' characters.
	Double

	// Round is line style using the rounded corners '╭' characters.
	Round
)
