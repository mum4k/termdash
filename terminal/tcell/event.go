// Copyright 2020 Google Inc.
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

package tcell

import (
	"image"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// tcell representation of the space key
var tcellSpaceKey = tcell.Key(' ')

// tcellToTd maps tcell key values to the termdash format.
var tcellToTd = map[tcell.Key]keyboard.Key{
	tcellSpaceKey:           keyboard.KeySpace,
	tcell.KeyF1:             keyboard.KeyF1,
	tcell.KeyF2:             keyboard.KeyF2,
	tcell.KeyF3:             keyboard.KeyF3,
	tcell.KeyF4:             keyboard.KeyF4,
	tcell.KeyF5:             keyboard.KeyF5,
	tcell.KeyF6:             keyboard.KeyF6,
	tcell.KeyF7:             keyboard.KeyF7,
	tcell.KeyF8:             keyboard.KeyF8,
	tcell.KeyF9:             keyboard.KeyF9,
	tcell.KeyF10:            keyboard.KeyF10,
	tcell.KeyF11:            keyboard.KeyF11,
	tcell.KeyF12:            keyboard.KeyF12,
	tcell.KeyInsert:         keyboard.KeyInsert,
	tcell.KeyDelete:         keyboard.KeyDelete,
	tcell.KeyHome:           keyboard.KeyHome,
	tcell.KeyEnd:            keyboard.KeyEnd,
	tcell.KeyPgUp:           keyboard.KeyPgUp,
	tcell.KeyPgDn:           keyboard.KeyPgDn,
	tcell.KeyUp:             keyboard.KeyArrowUp,
	tcell.KeyDown:           keyboard.KeyArrowDown,
	tcell.KeyLeft:           keyboard.KeyArrowLeft,
	tcell.KeyRight:          keyboard.KeyArrowRight,
	tcell.KeyEnter:          keyboard.KeyEnter,
	tcell.KeyCtrlA:          keyboard.KeyCtrlA,
	tcell.KeyCtrlB:          keyboard.KeyCtrlB,
	tcell.KeyCtrlC:          keyboard.KeyCtrlC,
	tcell.KeyCtrlD:          keyboard.KeyCtrlD,
	tcell.KeyCtrlE:          keyboard.KeyCtrlE,
	tcell.KeyCtrlF:          keyboard.KeyCtrlF,
	tcell.KeyCtrlG:          keyboard.KeyCtrlG,
	tcell.KeyCtrlJ:          keyboard.KeyCtrlJ,
	tcell.KeyCtrlK:          keyboard.KeyCtrlK,
	tcell.KeyCtrlL:          keyboard.KeyCtrlL,
	tcell.KeyCtrlN:          keyboard.KeyCtrlN,
	tcell.KeyCtrlO:          keyboard.KeyCtrlO,
	tcell.KeyCtrlP:          keyboard.KeyCtrlP,
	tcell.KeyCtrlQ:          keyboard.KeyCtrlQ,
	tcell.KeyCtrlR:          keyboard.KeyCtrlR,
	tcell.KeyCtrlS:          keyboard.KeyCtrlS,
	tcell.KeyCtrlT:          keyboard.KeyCtrlT,
	tcell.KeyCtrlU:          keyboard.KeyCtrlU,
	tcell.KeyCtrlV:          keyboard.KeyCtrlV,
	tcell.KeyCtrlW:          keyboard.KeyCtrlW,
	tcell.KeyCtrlX:          keyboard.KeyCtrlX,
	tcell.KeyCtrlY:          keyboard.KeyCtrlY,
	tcell.KeyCtrlZ:          keyboard.KeyCtrlZ,
	tcell.KeyBackspace:      keyboard.KeyBackspace,
	tcell.KeyTab:            keyboard.KeyTab,
	tcell.KeyEscape:         keyboard.KeyEsc,
	tcell.KeyCtrlBackslash:  keyboard.KeyCtrlBackslash,
	tcell.KeyCtrlRightSq:    keyboard.KeyCtrlRsqBracket,
	tcell.KeyCtrlUnderscore: keyboard.KeyCtrlUnderscore,
	tcell.KeyBackspace2:     keyboard.KeyBackspace2,
	tcell.KeyCtrlSpace:      keyboard.KeyCtrlSpace,
}

// convKey converts a tcell keyboard event to the termdash format.
func convKey(event *tcell.EventKey) terminalapi.Event {
	tcellKey := event.Key()

	if tcellKey == tcell.KeyRune {
		ch := event.Rune()
		return &terminalapi.Keyboard{
			Key: keyboard.Key(ch),
		}
	}

	k, ok := tcellToTd[tcellKey]
	if !ok {
		return terminalapi.NewErrorf("unknown keyboard key '%v' in a keyboard event %v", tcellKey, event.Name())
	}

	return &terminalapi.Keyboard{
		Key: k,
	}
}

// convMouse converts a tcell mouse event to the termdash format.
// Since tcell supports many combinations of mouse events, such as multiple mouse buttons pressed at the same time,
// this function returns nil if the event is unsupported by termdash.
func convMouse(event *tcell.EventMouse) terminalapi.Event {
	var button mouse.Button
	x, y := event.Position()

	tcellBtn := event.Buttons()

	// tcell uses signed int16 for button masks, and negative values are invalid
	if tcellBtn < 0 {
		return terminalapi.NewErrorf("unknown mouse key %v in a mouse event", tcellBtn)
	}

	// Get wheel events
	if tcellBtn&tcell.WheelUp != 0 {
		button = mouse.ButtonWheelUp
	} else if tcellBtn&tcell.WheelDown != 0 {
		button = mouse.ButtonWheelDown
	}

	// Return wheel event if found
	if button > 0 {
		return &terminalapi.Mouse{
			Position: image.Point{X: x, Y: y},
			Button:   button,
		}
	}

	switch tcellBtn = event.Buttons(); tcellBtn {
	case tcell.ButtonNone:
		button = mouse.ButtonRelease
	case tcell.Button1:
		button = mouse.ButtonLeft
	case tcell.Button2:
		button = mouse.ButtonRight
	case tcell.Button3:
		button = mouse.ButtonMiddle
	default:
		// Unknown event to termdash
		return nil
	}

	return &terminalapi.Mouse{
		Position: image.Point{X: x, Y: y},
		Button:   button,
	}
}

// convResize converts a tcell resize event to the termdash format.
func convResize(event *tcell.EventResize) terminalapi.Event {
	w, h := event.Size()
	size := image.Point{X: w, Y: h}
	if size.X < 0 || size.Y < 0 {
		return terminalapi.NewErrorf("terminal resized to negative size: %v", size)
	}
	return &terminalapi.Resize{
		Size: size,
	}
}

// toTermdashEvents converts a tcell event to the termdash event format.
// This function returns nil if the event is unsupported by termdash.
func toTermdashEvents(event tcell.Event) []terminalapi.Event {
	switch event := event.(type) {
	case *tcell.EventInterrupt:
		return []terminalapi.Event{
			terminalapi.NewError("event type EventInterrupt isn't supported"),
		}
	case *tcell.EventKey:
		return []terminalapi.Event{convKey(event)}
	case *tcell.EventMouse:
		mouseEvent := convMouse(event)
		if mouseEvent != nil {
			return []terminalapi.Event{mouseEvent}
		}
		return nil
	case *tcell.EventResize:
		return []terminalapi.Event{convResize(event)}
	case *tcell.EventError:
		return []terminalapi.Event{
			terminalapi.NewErrorf("encountered tcell error event: %v", event),
		}
	default:
		return []terminalapi.Event{
			terminalapi.NewErrorf("unknown tcell event type: %v", event),
		}
	}
}
