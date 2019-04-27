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

import (
	"errors"
	"fmt"
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"

	tbx "github.com/nsf/termbox-go"
)

func TestToTermdashEvents(t *testing.T) {
	tests := []struct {
		desc  string
		event tbx.Event
		want  []terminalapi.Event
	}{
		{
			desc: "unknown event type",
			event: tbx.Event{
				Type: 255,
			},
			want: []terminalapi.Event{
				terminalapi.NewError("unknown termbox event type: 255"),
			},
		},
		{
			desc: "interrupts aren't supported",
			event: tbx.Event{
				Type: tbx.EventInterrupt,
			},
			want: []terminalapi.Event{
				terminalapi.NewError("event type EventInterrupt isn't supported"),
			},
		},
		{
			desc: "raw events aren't supported",
			event: tbx.Event{
				Type: tbx.EventRaw,
			},
			want: []terminalapi.Event{
				terminalapi.NewError("event type EventRaw isn't supported"),
			},
		},
		{
			desc: "none events aren't supported",
			event: tbx.Event{
				Type: tbx.EventNone,
			},
			want: []terminalapi.Event{
				terminalapi.NewError("event type EventNone isn't supported"),
			},
		},
		{
			desc: "error event",
			event: tbx.Event{
				Type: tbx.EventError,
				Err:  errors.New("error event"),
			},
			want: []terminalapi.Event{
				terminalapi.NewError("input error occurred: error event"),
			},
		},
		{
			desc: "resize event",
			event: tbx.Event{
				Type:   tbx.EventResize,
				Width:  640,
				Height: 480,
			},
			want: []terminalapi.Event{
				&terminalapi.Resize{
					Size: image.Point{640, 480},
				},
			},
		},
		{
			desc: "resize event to a negative size",
			event: tbx.Event{
				Type:   tbx.EventResize,
				Width:  -1,
				Height: -1,
			},
			want: []terminalapi.Event{
				terminalapi.NewError("terminal resized to negative size: (-1,-1)"),
			},
		},
		{
			desc: "mouse event",
			event: tbx.Event{
				Type:   tbx.EventMouse,
				Key:    tbx.MouseLeft,
				MouseX: 100,
				MouseY: 200,
			},
			want: []terminalapi.Event{
				&terminalapi.Mouse{
					Position: image.Point{100, 200},
					Button:   mouse.ButtonLeft,
				},
			},
		},
		{
			desc: "keyboard event",
			event: tbx.Event{
				Type: tbx.EventKey,
				Key:  tbx.KeyF1,
			},
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
		key     tbx.Key
		want    mouse.Button
		wantErr bool
	}{
		{wantErr: true},
		{key: tbx.KeyF1, wantErr: true},
		{key: 1, wantErr: true},
		{key: tbx.MouseLeft, want: mouse.ButtonLeft},
		{key: tbx.MouseMiddle, want: mouse.ButtonMiddle},
		{key: tbx.MouseRight, want: mouse.ButtonRight},
		{key: tbx.MouseRelease, want: mouse.ButtonRelease},
		{key: tbx.MouseWheelUp, want: mouse.ButtonWheelUp},
		{key: tbx.MouseWheelDown, want: mouse.ButtonWheelDown},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("key:%v want:%v", tc.key, tc.want), func(t *testing.T) {

			evs := toTermdashEvents(tbx.Event{Type: tbx.EventMouse, Key: tc.key})
			if got, want := len(evs), 1; got != want {
				t.Fatalf("toTermdashEvents => got %d events, want %d", got, want)
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
				if got := e.Button; got != tc.want {
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
		key     tbx.Key
		ch      rune
		want    keyboard.Key
		wantErr bool
	}{
		{key: tbx.KeyF1, ch: 'a', wantErr: true},
		{key: 2000, wantErr: true},
		{ch: 'a', want: 'a'},
		{ch: 'A', want: 'A'},
		{ch: 'z', want: 'z'},
		{ch: 'Z', want: 'Z'},
		{ch: '0', want: '0'},
		{ch: '9', want: '9'},
		{ch: '!', want: '!'},
		{ch: ')', want: ')'},
		{key: tbx.KeySpace, want: keyboard.KeySpace},
		{key: tbx.KeyF1, want: keyboard.KeyF1},
		{key: tbx.KeyF2, want: keyboard.KeyF2},
		{key: tbx.KeyF3, want: keyboard.KeyF3},
		{key: tbx.KeyF4, want: keyboard.KeyF4},
		{key: tbx.KeyF5, want: keyboard.KeyF5},
		{key: tbx.KeyF6, want: keyboard.KeyF6},
		{key: tbx.KeyF7, want: keyboard.KeyF7},
		{key: tbx.KeyF8, want: keyboard.KeyF8},
		{key: tbx.KeyF9, want: keyboard.KeyF9},
		{key: tbx.KeyF10, want: keyboard.KeyF10},
		{key: tbx.KeyF11, want: keyboard.KeyF11},
		{key: tbx.KeyF12, want: keyboard.KeyF12},
		{key: tbx.KeyInsert, want: keyboard.KeyInsert},
		{key: tbx.KeyDelete, want: keyboard.KeyDelete},
		{key: tbx.KeyHome, want: keyboard.KeyHome},
		{key: tbx.KeyEnd, want: keyboard.KeyEnd},
		{key: tbx.KeyPgup, want: keyboard.KeyPgUp},
		{key: tbx.KeyPgdn, want: keyboard.KeyPgDn},
		{key: tbx.KeyArrowUp, want: keyboard.KeyArrowUp},
		{key: tbx.KeyArrowDown, want: keyboard.KeyArrowDown},
		{key: tbx.KeyArrowLeft, want: keyboard.KeyArrowLeft},
		{key: tbx.KeyArrowRight, want: keyboard.KeyArrowRight},
		{key: tbx.KeyCtrlTilde, want: keyboard.KeyCtrlTilde},
		{key: tbx.KeyCtrlTilde, want: keyboard.KeyCtrl2},
		{key: tbx.KeyCtrlTilde, want: keyboard.KeyCtrlSpace},
		{key: tbx.KeyCtrl2, want: keyboard.KeyCtrlTilde},
		{key: tbx.KeyCtrlSpace, want: keyboard.KeyCtrlTilde},
		{key: tbx.KeyCtrlA, want: keyboard.KeyCtrlA},
		{key: tbx.KeyCtrlB, want: keyboard.KeyCtrlB},
		{key: tbx.KeyCtrlC, want: keyboard.KeyCtrlC},
		{key: tbx.KeyCtrlD, want: keyboard.KeyCtrlD},
		{key: tbx.KeyCtrlE, want: keyboard.KeyCtrlE},
		{key: tbx.KeyCtrlF, want: keyboard.KeyCtrlF},
		{key: tbx.KeyCtrlG, want: keyboard.KeyCtrlG},
		{key: tbx.KeyBackspace, want: keyboard.KeyBackspace},
		{key: tbx.KeyBackspace, want: keyboard.KeyCtrlH},
		{key: tbx.KeyCtrlH, want: keyboard.KeyBackspace},
		{key: tbx.KeyTab, want: keyboard.KeyTab},
		{key: tbx.KeyTab, want: keyboard.KeyCtrlI},
		{key: tbx.KeyCtrlI, want: keyboard.KeyTab},
		{key: tbx.KeyCtrlJ, want: keyboard.KeyCtrlJ},
		{key: tbx.KeyCtrlK, want: keyboard.KeyCtrlK},
		{key: tbx.KeyCtrlL, want: keyboard.KeyCtrlL},
		{key: tbx.KeyEnter, want: keyboard.KeyEnter},
		{key: tbx.KeyEnter, want: keyboard.KeyCtrlM},
		{key: tbx.KeyCtrlM, want: keyboard.KeyEnter},
		{key: tbx.KeyCtrlN, want: keyboard.KeyCtrlN},
		{key: tbx.KeyCtrlO, want: keyboard.KeyCtrlO},
		{key: tbx.KeyCtrlP, want: keyboard.KeyCtrlP},
		{key: tbx.KeyCtrlQ, want: keyboard.KeyCtrlQ},
		{key: tbx.KeyCtrlR, want: keyboard.KeyCtrlR},
		{key: tbx.KeyCtrlS, want: keyboard.KeyCtrlS},
		{key: tbx.KeyCtrlT, want: keyboard.KeyCtrlT},
		{key: tbx.KeyCtrlU, want: keyboard.KeyCtrlU},
		{key: tbx.KeyCtrlV, want: keyboard.KeyCtrlV},
		{key: tbx.KeyCtrlW, want: keyboard.KeyCtrlW},
		{key: tbx.KeyCtrlX, want: keyboard.KeyCtrlX},
		{key: tbx.KeyCtrlY, want: keyboard.KeyCtrlY},
		{key: tbx.KeyCtrlZ, want: keyboard.KeyCtrlZ},
		{key: tbx.KeyEsc, want: keyboard.KeyEsc},
		{key: tbx.KeyEsc, want: keyboard.KeyCtrlLsqBracket},
		{key: tbx.KeyEsc, want: keyboard.KeyCtrl3},
		{key: tbx.KeyCtrlLsqBracket, want: keyboard.KeyEsc},
		{key: tbx.KeyCtrl3, want: keyboard.KeyEsc},
		{key: tbx.KeyCtrl4, want: keyboard.KeyCtrl4},
		{key: tbx.KeyCtrl4, want: keyboard.KeyCtrlBackslash},
		{key: tbx.KeyCtrlBackslash, want: keyboard.KeyCtrl4},
		{key: tbx.KeyCtrl5, want: keyboard.KeyCtrl5},
		{key: tbx.KeyCtrl5, want: keyboard.KeyCtrlRsqBracket},
		{key: tbx.KeyCtrlRsqBracket, want: keyboard.KeyCtrl5},
		{key: tbx.KeyCtrl6, want: keyboard.KeyCtrl6},
		{key: tbx.KeyCtrl7, want: keyboard.KeyCtrl7},
		{key: tbx.KeyCtrl7, want: keyboard.KeyCtrlSlash},
		{key: tbx.KeyCtrl7, want: keyboard.KeyCtrlUnderscore},
		{key: tbx.KeyCtrlSlash, want: keyboard.KeyCtrl7},
		{key: tbx.KeyCtrlUnderscore, want: keyboard.KeyCtrl7},
		{key: tbx.KeyBackspace2, want: keyboard.KeyBackspace2},
		{key: tbx.KeyBackspace2, want: keyboard.KeyCtrl8},
		{key: tbx.KeyCtrl8, want: keyboard.KeyBackspace2},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("key:%v and ch:%v want:%v", tc.key, tc.ch, tc.want), func(t *testing.T) {
			evs := toTermdashEvents(tbx.Event{
				Type: tbx.EventKey,
				Key:  tc.key,
				Ch:   tc.ch,
			})

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
