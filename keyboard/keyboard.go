// Package keyboard defines well known keyboard keys and shortcuts.
package keyboard

// Key represents a single button on the keyboard.
// Printable characters are set to their ASCII/Unicode rune value.
// Non-printable (control) characters are equal to one of the constants defined
// below.
type Key rune

// String implements fmt.Stringer()
func (b Key) String() string {
	if n, ok := buttonNames[b]; ok {
		return n
	} else if b >= 0 {
		return string(b)
	}
	return "KeyUnknown"
}

// buttonNames maps Key values to human readable names.
var buttonNames = map[Key]string{
	KeyF1:         "KeyF1",
	KeyF2:         "KeyF2",
	KeyF3:         "KeyF3",
	KeyF4:         "KeyF4",
	KeyF5:         "KeyF5",
	KeyF6:         "KeyF6",
	KeyF7:         "KeyF7",
	KeyF8:         "KeyF8",
	KeyF9:         "KeyF9",
	KeyF10:        "KeyF10",
	KeyF11:        "KeyF11",
	KeyF12:        "KeyF12",
	KeyInsert:     "KeyInsert",
	KeyDelete:     "KeyDelete",
	KeyHome:       "KeyHome",
	KeyEnd:        "KeyEnd",
	KeyPgUp:       "KeyPgUp",
	KeyPgDn:       "KeyPgDn",
	KeyArrowUp:    "KeyArrowUp",
	KeyArrowDown:  "KeyArrowDown",
	KeyArrowLeft:  "KeyArrowLeft",
	KeyArrowRight: "KeyArrowRight",
	KeyBackspace:  "KeyBackspace",
	KeyTab:        "KeyTab",
	KeyEnter:      "KeyEnter",
	KeyEsc:        "KeyEsc",
	KeyCtrl:       "KeyCtrl",
}

// Printable characters, but worth having constants for them.
const (
	KeySpace = ' '
)

// Negative values for non-printable characters.
const (
	KeyF1 Key = -(iota + 1)
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyInsert
	KeyDelete
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDn
	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeyBackspace
	KeyTab
	KeyEnter
	KeyEsc
	KeyCtrl
)
