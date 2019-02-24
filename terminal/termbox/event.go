// Copyright 2018 Google Inc.
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

package termbox

// event.go converts termbox events to the termdash format.

import (
	"image"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
	tbx "github.com/nsf/termbox-go"
)

// newKeyboard creates a new termdash keyboard events with the provided keys.
func newKeyboard(keys ...keyboard.Key) []terminalapi.Event {
	var evs []terminalapi.Event
	for _, k := range keys {
		evs = append(evs, &terminalapi.Keyboard{Key: k})
	}
	return evs
}

// convKey converts a termbox keyboard event to the termdash format.
func convKey(tbxEv tbx.Event) []terminalapi.Event {
	if tbxEv.Key != 0 && tbxEv.Ch != 0 {
		return []terminalapi.Event{
			terminalapi.NewErrorf("the key event contain both a key(%v) and a character(%v)", tbxEv.Key, tbxEv.Ch),
		}
	}

	if tbxEv.Ch != 0 {
		return []terminalapi.Event{&terminalapi.Keyboard{
			Key: keyboard.Key(tbxEv.Ch),
		}}
	}

	switch k := tbxEv.Key; k {
	case tbx.KeySpace:
		return newKeyboard(keyboard.KeySpace)
	case tbx.KeyF1:
		return newKeyboard(keyboard.KeyF1)
	case tbx.KeyF2:
		return newKeyboard(keyboard.KeyF2)
	case tbx.KeyF3:
		return newKeyboard(keyboard.KeyF3)
	case tbx.KeyF4:
		return newKeyboard(keyboard.KeyF4)
	case tbx.KeyF5:
		return newKeyboard(keyboard.KeyF5)
	case tbx.KeyF6:
		return newKeyboard(keyboard.KeyF6)
	case tbx.KeyF7:
		return newKeyboard(keyboard.KeyF7)
	case tbx.KeyF8:
		return newKeyboard(keyboard.KeyF8)
	case tbx.KeyF9:
		return newKeyboard(keyboard.KeyF9)
	case tbx.KeyF10:
		return newKeyboard(keyboard.KeyF10)
	case tbx.KeyF11:
		return newKeyboard(keyboard.KeyF11)
	case tbx.KeyF12:
		return newKeyboard(keyboard.KeyF12)
	case tbx.KeyInsert:
		return newKeyboard(keyboard.KeyInsert)
	case tbx.KeyDelete:
		return newKeyboard(keyboard.KeyDelete)
	case tbx.KeyHome:
		return newKeyboard(keyboard.KeyHome)
	case tbx.KeyEnd:
		return newKeyboard(keyboard.KeyEnd)
	case tbx.KeyPgup:
		return newKeyboard(keyboard.KeyPgUp)
	case tbx.KeyPgdn:
		return newKeyboard(keyboard.KeyPgDn)
	case tbx.KeyArrowUp:
		return newKeyboard(keyboard.KeyArrowUp)
	case tbx.KeyArrowDown:
		return newKeyboard(keyboard.KeyArrowDown)
	case tbx.KeyArrowLeft:
		return newKeyboard(keyboard.KeyArrowLeft)
	case tbx.KeyArrowRight:
		return newKeyboard(keyboard.KeyArrowRight)
	case tbx.KeyBackspace /*, tbx.KeyCtrlH */ :
		return newKeyboard(keyboard.KeyBackspace)
	case tbx.KeyTab /*, tbx.KeyCtrlI */ :
		return newKeyboard(keyboard.KeyTab)
	case tbx.KeyEnter /*, tbx.KeyCtrlM*/ :
		return newKeyboard(keyboard.KeyEnter)
	case tbx.KeyEsc /*, tbx.KeyCtrlLsqBracket, tbx.KeyCtrl3 */ :
		return newKeyboard(keyboard.KeyEsc)
	case tbx.KeyCtrl2 /*, tbx.KeyCtrlTilde, tbx.KeyCtrlSpace */ :
		return newKeyboard(keyboard.KeyCtrl, '2')
	case tbx.KeyCtrl4 /*, tbx.KeyCtrlBackslash */ :
		return newKeyboard(keyboard.KeyCtrl, '4')
	case tbx.KeyCtrl5 /*, tbx.KeyCtrlRsqBracket */ :
		return newKeyboard(keyboard.KeyCtrl, '5')
	case tbx.KeyCtrl6:
		return newKeyboard(keyboard.KeyCtrl, '6')
	case tbx.KeyCtrl7 /*, tbx.KeyCtrlSlash, tbx.KeyCtrlUnderscore */ :
		return newKeyboard(keyboard.KeyCtrl, '7')
	case tbx.KeyCtrl8:
		return newKeyboard(keyboard.KeyCtrl, '8')
	case tbx.KeyCtrlA:
		return newKeyboard(keyboard.KeyCtrl, 'a')
	case tbx.KeyCtrlB:
		return newKeyboard(keyboard.KeyCtrl, 'b')
	case tbx.KeyCtrlC:
		return newKeyboard(keyboard.KeyCtrl, 'c')
	case tbx.KeyCtrlD:
		return newKeyboard(keyboard.KeyCtrl, 'd')
	case tbx.KeyCtrlE:
		return newKeyboard(keyboard.KeyCtrl, 'e')
	case tbx.KeyCtrlF:
		return newKeyboard(keyboard.KeyCtrl, 'f')
	case tbx.KeyCtrlG:
		return newKeyboard(keyboard.KeyCtrl, 'g')
	case tbx.KeyCtrlJ:
		return newKeyboard(keyboard.KeyCtrl, 'j')
	case tbx.KeyCtrlK:
		return newKeyboard(keyboard.KeyCtrl, 'k')
	case tbx.KeyCtrlL:
		return newKeyboard(keyboard.KeyCtrl, 'l')
	case tbx.KeyCtrlN:
		return newKeyboard(keyboard.KeyCtrl, 'n')
	case tbx.KeyCtrlO:
		return newKeyboard(keyboard.KeyCtrl, 'o')
	case tbx.KeyCtrlP:
		return newKeyboard(keyboard.KeyCtrl, 'p')
	case tbx.KeyCtrlQ:
		return newKeyboard(keyboard.KeyCtrl, 'q')
	case tbx.KeyCtrlR:
		return newKeyboard(keyboard.KeyCtrl, 'r')
	case tbx.KeyCtrlS:
		return newKeyboard(keyboard.KeyCtrl, 's')
	case tbx.KeyCtrlT:
		return newKeyboard(keyboard.KeyCtrl, 't')
	case tbx.KeyCtrlU:
		return newKeyboard(keyboard.KeyCtrl, 'u')
	case tbx.KeyCtrlV:
		return newKeyboard(keyboard.KeyCtrl, 'v')
	case tbx.KeyCtrlW:
		return newKeyboard(keyboard.KeyCtrl, 'w')
	case tbx.KeyCtrlX:
		return newKeyboard(keyboard.KeyCtrl, 'x')
	case tbx.KeyCtrlY:
		return newKeyboard(keyboard.KeyCtrl, 'y')
	case tbx.KeyCtrlZ:
		return newKeyboard(keyboard.KeyCtrl, 'z')
	default:
		return []terminalapi.Event{
			terminalapi.NewErrorf("unknown keyboard key %v in a keyboard event", k),
		}
	}
}

// convMouse converts a termbox mouse event to the termdash format.
func convMouse(tbxEv tbx.Event) terminalapi.Event {
	var button mouse.Button

	switch k := tbxEv.Key; k {
	case tbx.MouseLeft:
		button = mouse.ButtonLeft
	case tbx.MouseMiddle:
		button = mouse.ButtonMiddle
	case tbx.MouseRight:
		button = mouse.ButtonRight
	case tbx.MouseRelease:
		button = mouse.ButtonRelease
	case tbx.MouseWheelUp:
		button = mouse.ButtonWheelUp
	case tbx.MouseWheelDown:
		button = mouse.ButtonWheelDown
	default:
		return terminalapi.NewErrorf("unknown mouse key %v in a mouse event", k)
	}

	return &terminalapi.Mouse{
		Position: image.Point{tbxEv.MouseX, tbxEv.MouseY},
		Button:   button,
	}
}

// convResize converts a termbox resize event to the termdash format.
func convResize(tbxEv tbx.Event) terminalapi.Event {
	size := image.Point{tbxEv.Width, tbxEv.Height}
	if size.X < 0 || size.Y < 0 {
		return terminalapi.NewErrorf("terminal resized to negative size: %v", size)
	}
	return &terminalapi.Resize{
		Size: size,
	}
}

// toTermdashEvents converts a termbox event to the termdash event format.
func toTermdashEvents(tbxEv tbx.Event) []terminalapi.Event {
	switch t := tbxEv.Type; t {
	case tbx.EventInterrupt:
		return []terminalapi.Event{
			terminalapi.NewError("event type EventInterrupt isn't supported"),
		}
	case tbx.EventRaw:
		return []terminalapi.Event{
			terminalapi.NewError("event type EventRaw isn't supported"),
		}
	case tbx.EventNone:
		return []terminalapi.Event{
			terminalapi.NewError("event type EventNone isn't supported"),
		}
	case tbx.EventError:
		return []terminalapi.Event{
			terminalapi.NewErrorf("input error occurred: %v", tbxEv.Err),
		}
	case tbx.EventResize:
		return []terminalapi.Event{convResize(tbxEv)}
	case tbx.EventMouse:
		return []terminalapi.Event{convMouse(tbxEv)}
	case tbx.EventKey:
		return convKey(tbxEv)
	default:
		return []terminalapi.Event{
			terminalapi.NewErrorf("unknown termbox event type: %v", t),
		}
	}
}
