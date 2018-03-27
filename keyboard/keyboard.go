// Package keyboard defines well known keyboard keys and shortcuts.
package keyboard

// Button represents a single button on the keyboard.
type Button rune

// String implements fmt.Stringer()
func (b Button) String() string {
	if n, ok := buttonNames[b]; ok {
		return n
	}
	return "ButtonUnknown"
}

// buttonNames maps Button values to human readable names.
var buttonNames = map[Button]string{
	ButtonArrowDown:  "ButtonArrowDown",
	ButtonArrowLeft:  "ButtonArrowLeft",
	ButtonArrowRight: "ButtonArrowRight",
	ButtonArrowUp:    "ButtonArrowUp",
	ButtonBackspace:  "ButtonBackspace",
	ButtonDelete:     "ButtonDelete",
	ButtonEnd:        "ButtonEnd",
	ButtonEnter:      "ButtonEnter",
	ButtonEsc:        "ButtonEsc",
	ButtonF1:         "ButtonF1",
	ButtonF10:        "ButtonF10",
	ButtonF11:        "ButtonF11",
	ButtonF12:        "ButtonF12",
	ButtonF2:         "ButtonF2",
	ButtonF3:         "ButtonF3",
	ButtonF4:         "ButtonF4",
	ButtonF5:         "ButtonF5",
	ButtonF6:         "ButtonF6",
	ButtonF7:         "ButtonF7",
	ButtonF8:         "ButtonF8",
	ButtonF9:         "ButtonF9",
	ButtonHome:       "ButtonHome",
	ButtonInsert:     "ButtonInsert",
	ButtonPgdn:       "ButtonPgdn",
	ButtonPgup:       "ButtonPgup",
	ButtonSpace:      "ButtonSpace",
	ButtonTab:        "ButtonTab",
	ButtonTilde:      "ButtonTilde",
}

const (
	ButtonArrowDown Button = -(iota + 1)
	ButtonArrowLeft
	ButtonArrowRight
	ButtonArrowUp
	ButtonBackspace
	ButtonDelete
	ButtonEnd
	ButtonEnter
	ButtonEsc
	ButtonF1
	ButtonF10
	ButtonF11
	ButtonF12
	ButtonF2
	ButtonF3
	ButtonF4
	ButtonF5
	ButtonF6
	ButtonF7
	ButtonF8
	ButtonF9
	ButtonHome
	ButtonInsert
	ButtonPgdn
	ButtonPgup
	ButtonSpace
	ButtonTab
	ButtonTilde
)

// Modifier represents a modified key on the keyboard, i.e. a keys that
// together with buttons can form shortcuts.
type Modifier int

// String implements fmt.Stringer()
func (m Modifier) String() string {
	if n, ok := modifierNames[m]; ok {
		return n
	}
	return "ModifierUnknown"
}

// modifierNames maps Modifier values to human readable names.
var modifierNames = map[Modifier]string{
	ModifierShift: "ModifierShift",
	ModifierCtrl:  "ModifierCtrl",
	ModifierAlt:   "ModifierAlt",
	ModifierMeta:  "ModifierMeta",
}

const (
	modifierUnknown Modifier = iota
	ModifierShift
	ModifierCtrl
	ModifierAlt

	// ModifierMeta is the platform specific key, i.e. Windows key on windows
	// or Apple (command) key on MacOS keyboard.
	ModifierMeta
)

// Shortcut is a key combination pressed on the keyboard.
type Shortcut struct {
	// Modifiers contains zero or more unique modifier keys.
	Modifiers []Modifier

	// Key is the key pressed on the keyboard.
	// Either equals to one of the defined Button values or contains the raw
	// Unicode byte sequence.
	Key Button
}
