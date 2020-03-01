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
	"errors"
	"fmt"
	"image"
	"testing"
	"time"

	"github.com/gdamore/tcell"
	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

type mockUnknownEvent struct {
}

func (m *mockUnknownEvent) When() time.Time {
	return time.Now()
}

func TestToTermdashEvents(t *testing.T) {
	tests := []struct {
		desc  string
		event tcell.Event
		want  []terminalapi.Event
	}{
		{
			desc:  "unknown event type",
			event: &mockUnknownEvent{},
			want: []terminalapi.Event{
				terminalapi.NewError("unknown tcell event type: &{}"),
			},
		},
		{
			desc:  "interrupts aren't supported",
			event: tcell.NewEventInterrupt(nil),
			want: []terminalapi.Event{
				terminalapi.NewError("event type EventInterrupt isn't supported"),
			},
		},
		{
			desc:  "error event",
			event: tcell.NewEventError(errors.New("error event")),
			want: []terminalapi.Event{
				terminalapi.NewError("encountered tcell error event: error event"),
			},
		},
		{
			desc:  "resize event",
			event: tcell.NewEventResize(640, 480),
			want: []terminalapi.Event{
				&terminalapi.Resize{
					Size: image.Point{X: 640, Y: 480},
				},
			},
		},
		{
			desc:  "resize event to a negative size",
			event: tcell.NewEventResize(-1, -1),
			want: []terminalapi.Event{
				terminalapi.NewError("terminal resized to negative size: (-1,-1)"),
			},
		},
		{
			desc:  "mouse event",
			event: tcell.NewEventMouse(100, 200, tcell.Button1, tcell.ModNone),
			want: []terminalapi.Event{
				&terminalapi.Mouse{
					Position: image.Point{X: 100, Y: 200},
					Button:   mouse.ButtonLeft,
				},
			},
		},
		{
			desc:  "keyboard event",
			event: tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone),
			want: []terminalapi.Event{
				&terminalapi.Keyboard{
					Key: keyboard.KeyF1,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := toTermdashEvents(tc.event)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("toTermdashEvents => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestMouseButtons(t *testing.T) {
	tests := []struct {
		btnMask tcell.ButtonMask
		want    []mouse.Button
		wantErr bool
	}{
		{btnMask: -1, want: []mouse.Button{mouse.Button(-1)}, wantErr: true},
		{btnMask: tcell.Button1, want: []mouse.Button{mouse.ButtonLeft}},
		{btnMask: tcell.Button3, want: []mouse.Button{mouse.ButtonMiddle}},
		{btnMask: tcell.Button2, want: []mouse.Button{mouse.ButtonRight}},
		{btnMask: tcell.ButtonNone, want: []mouse.Button{mouse.ButtonRelease}},
		{btnMask: tcell.WheelUp, want: []mouse.Button{mouse.ButtonWheelUp}},
		{btnMask: tcell.WheelDown, want: []mouse.Button{mouse.ButtonWheelDown}},
		{btnMask: tcell.Button1 | tcell.Button2, want: nil},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("key:%v want:%v", tc.btnMask, tc.want), func(t *testing.T) {

			evs := toTermdashEvents(tcell.NewEventMouse(0, 0, tc.btnMask, tcell.ModNone))
			if got, want := len(evs), len(tc.want); got != want {
				t.Fatalf("toTermdashEvents => got %d events, want %d", got, want)
			}

			// Events that may exist for the terminal implementation but are not valid for termdash will be nil
			if len(tc.want) == 0 {
				return
			}

			ev := evs[0]
			if err, ok := ev.(*terminalapi.Error); ok != tc.wantErr {
				t.Fatalf("toTermdashEvents => unexpected error:%v, wantErr: %v", err, tc.wantErr)
			}
			if _, ok := ev.(*terminalapi.Error); ok {
				return
			}

			switch e := ev.(type) {
			case *terminalapi.Mouse:
				if got := e.Button; got != tc.want[0] {
					t.Errorf("toTermdashEvents => got %v, want %v", got, tc.want)
				}

			default:
				t.Fatalf("toTermdashEvents => unexpected event type %T", e)
			}
		})
	}
}

func TestKeyboardKeys(t *testing.T) {
	tests := []struct {
		key     tcell.Key
		ch      rune
		want    keyboard.Key
		wantErr bool
	}{
		{key: 2000, wantErr: true},
		{key: tcell.KeyRune, ch: 'a', want: 'a'},
		{key: tcell.KeyRune, ch: 'A', want: 'A'},
		{key: tcell.KeyRune, ch: 'z', want: 'z'},
		{key: tcell.KeyRune, ch: 'Z', want: 'Z'},
		{key: tcell.KeyRune, ch: '0', want: '0'},
		{key: tcell.KeyRune, ch: '9', want: '9'},
		{key: tcell.KeyRune, ch: '!', want: '!'},
		{key: tcell.KeyRune, ch: ')', want: ')'},
		{key: tcellSpaceKey, want: keyboard.KeySpace},
		{key: tcell.KeyF1, want: keyboard.KeyF1},
		{key: tcell.KeyF2, want: keyboard.KeyF2},
		{key: tcell.KeyF3, want: keyboard.KeyF3},
		{key: tcell.KeyF4, want: keyboard.KeyF4},
		{key: tcell.KeyF5, want: keyboard.KeyF5},
		{key: tcell.KeyF6, want: keyboard.KeyF6},
		{key: tcell.KeyF7, want: keyboard.KeyF7},
		{key: tcell.KeyF8, want: keyboard.KeyF8},
		{key: tcell.KeyF9, want: keyboard.KeyF9},
		{key: tcell.KeyF10, want: keyboard.KeyF10},
		{key: tcell.KeyF11, want: keyboard.KeyF11},
		{key: tcell.KeyF12, want: keyboard.KeyF12},
		{key: tcell.KeyInsert, want: keyboard.KeyInsert},
		{key: tcell.KeyDelete, want: keyboard.KeyDelete},
		{key: tcell.KeyHome, want: keyboard.KeyHome},
		{key: tcell.KeyEnd, want: keyboard.KeyEnd},
		{key: tcell.KeyPgUp, want: keyboard.KeyPgUp},
		{key: tcell.KeyPgDn, want: keyboard.KeyPgDn},
		{key: tcell.KeyUp, want: keyboard.KeyArrowUp},
		{key: tcell.KeyDown, want: keyboard.KeyArrowDown},
		{key: tcell.KeyLeft, want: keyboard.KeyArrowLeft},
		{key: tcell.KeyRight, want: keyboard.KeyArrowRight},
		{key: tcell.KeyCtrlSpace, want: keyboard.KeyCtrlTilde},
		{key: tcell.KeyCtrlA, want: keyboard.KeyCtrlA},
		{key: tcell.KeyCtrlB, want: keyboard.KeyCtrlB},
		{key: tcell.KeyCtrlC, want: keyboard.KeyCtrlC},
		{key: tcell.KeyCtrlD, want: keyboard.KeyCtrlD},
		{key: tcell.KeyCtrlE, want: keyboard.KeyCtrlE},
		{key: tcell.KeyCtrlF, want: keyboard.KeyCtrlF},
		{key: tcell.KeyCtrlG, want: keyboard.KeyCtrlG},
		{key: tcell.KeyBackspace, want: keyboard.KeyBackspace},
		{key: tcell.KeyBackspace, want: keyboard.KeyCtrlH},
		{key: tcell.KeyCtrlH, want: keyboard.KeyBackspace},
		{key: tcell.KeyTab, want: keyboard.KeyTab},
		{key: tcell.KeyTab, want: keyboard.KeyCtrlI},
		{key: tcell.KeyCtrlI, want: keyboard.KeyTab},
		{key: tcell.KeyCtrlJ, want: keyboard.KeyCtrlJ},
		{key: tcell.KeyCtrlK, want: keyboard.KeyCtrlK},
		{key: tcell.KeyCtrlL, want: keyboard.KeyCtrlL},
		{key: tcell.KeyEnter, want: keyboard.KeyEnter},
		{key: tcell.KeyEnter, want: keyboard.KeyCtrlM},
		{key: tcell.KeyCtrlM, want: keyboard.KeyEnter},
		{key: tcell.KeyCtrlN, want: keyboard.KeyCtrlN},
		{key: tcell.KeyCtrlO, want: keyboard.KeyCtrlO},
		{key: tcell.KeyCtrlP, want: keyboard.KeyCtrlP},
		{key: tcell.KeyCtrlQ, want: keyboard.KeyCtrlQ},
		{key: tcell.KeyCtrlR, want: keyboard.KeyCtrlR},
		{key: tcell.KeyCtrlS, want: keyboard.KeyCtrlS},
		{key: tcell.KeyCtrlT, want: keyboard.KeyCtrlT},
		{key: tcell.KeyCtrlU, want: keyboard.KeyCtrlU},
		{key: tcell.KeyCtrlV, want: keyboard.KeyCtrlV},
		{key: tcell.KeyCtrlW, want: keyboard.KeyCtrlW},
		{key: tcell.KeyCtrlX, want: keyboard.KeyCtrlX},
		{key: tcell.KeyCtrlY, want: keyboard.KeyCtrlY},
		{key: tcell.KeyCtrlZ, want: keyboard.KeyCtrlZ},
		{key: tcell.KeyEsc, want: keyboard.KeyEsc},
		{key: tcell.KeyEsc, want: keyboard.KeyCtrlLsqBracket},
		{key: tcell.KeyEsc, want: keyboard.KeyCtrl3},
		{key: tcell.KeyCtrlLeftSq, want: keyboard.KeyEsc},
		{key: tcell.KeyCtrlBackslash, want: keyboard.KeyCtrl4},
		{key: tcell.KeyCtrlRightSq, want: keyboard.KeyCtrl5},
		{key: tcell.KeyCtrlUnderscore, want: keyboard.KeyCtrlUnderscore},
		{key: tcell.KeyBackspace2, want: keyboard.KeyBackspace2},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("key:%v and ch:%v want:%v", tc.key, tc.ch, tc.want), func(t *testing.T) {
			evs := toTermdashEvents(tcell.NewEventKey(tc.key, tc.ch, tcell.ModNone))

			gotCount := len(evs)
			wantCount := 1
			if gotCount != wantCount {
				t.Fatalf("toTermdashEvents => got %d events, want %d, events were:\n%v", gotCount, wantCount, pretty.Sprint(evs))
			}
			ev := evs[0]

			if err, ok := ev.(*terminalapi.Error); ok != tc.wantErr {
				t.Fatalf("toTermdashEvents => unexpected error:%v, wantErr: %v", err, tc.wantErr)
			}
			if _, ok := ev.(*terminalapi.Error); ok {
				return
			}

			switch e := ev.(type) {
			case *terminalapi.Keyboard:
				if got, want := e.Key, tc.want; got != want {
					t.Errorf("toTermdashEvents => got key %v, want %v", got, want)
				}

			default:
				t.Fatalf("toTermdashEvents => unexpected event type %T", e)
			}
		})
	}
}
