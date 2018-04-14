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
	"github.com/mum4k/termdash/terminalapi"

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
		want    []keyboard.Key
		wantErr bool
	}{
		{key: tbx.KeyF1, ch: 'a', wantErr: true},
		{key: 2000, wantErr: true},
		{ch: 'a', want: []keyboard.Key{'a'}},
		{ch: 'A', want: []keyboard.Key{'A'}},
		{ch: 'z', want: []keyboard.Key{'z'}},
		{ch: 'Z', want: []keyboard.Key{'Z'}},
		{ch: '0', want: []keyboard.Key{'0'}},
		{ch: '9', want: []keyboard.Key{'9'}},
		{ch: '!', want: []keyboard.Key{'!'}},
		{ch: ')', want: []keyboard.Key{')'}},
		{key: tbx.KeySpace, want: []keyboard.Key{keyboard.KeySpace}},
		{key: tbx.KeyF1, want: []keyboard.Key{keyboard.KeyF1}},
		{key: tbx.KeyF2, want: []keyboard.Key{keyboard.KeyF2}},
		{key: tbx.KeyF3, want: []keyboard.Key{keyboard.KeyF3}},
		{key: tbx.KeyF4, want: []keyboard.Key{keyboard.KeyF4}},
		{key: tbx.KeyF5, want: []keyboard.Key{keyboard.KeyF5}},
		{key: tbx.KeyF6, want: []keyboard.Key{keyboard.KeyF6}},
		{key: tbx.KeyF7, want: []keyboard.Key{keyboard.KeyF7}},
		{key: tbx.KeyF8, want: []keyboard.Key{keyboard.KeyF8}},
		{key: tbx.KeyF9, want: []keyboard.Key{keyboard.KeyF9}},
		{key: tbx.KeyF10, want: []keyboard.Key{keyboard.KeyF10}},
		{key: tbx.KeyF11, want: []keyboard.Key{keyboard.KeyF11}},
		{key: tbx.KeyF12, want: []keyboard.Key{keyboard.KeyF12}},
		{key: tbx.KeyInsert, want: []keyboard.Key{keyboard.KeyInsert}},
		{key: tbx.KeyDelete, want: []keyboard.Key{keyboard.KeyDelete}},
		{key: tbx.KeyHome, want: []keyboard.Key{keyboard.KeyHome}},
		{key: tbx.KeyEnd, want: []keyboard.Key{keyboard.KeyEnd}},
		{key: tbx.KeyPgup, want: []keyboard.Key{keyboard.KeyPgUp}},
		{key: tbx.KeyPgdn, want: []keyboard.Key{keyboard.KeyPgDn}},
		{key: tbx.KeyArrowUp, want: []keyboard.Key{keyboard.KeyArrowUp}},
		{key: tbx.KeyArrowDown, want: []keyboard.Key{keyboard.KeyArrowDown}},
		{key: tbx.KeyArrowLeft, want: []keyboard.Key{keyboard.KeyArrowLeft}},
		{key: tbx.KeyArrowRight, want: []keyboard.Key{keyboard.KeyArrowRight}},
		{key: tbx.KeyBackspace, want: []keyboard.Key{keyboard.KeyBackspace}},
		{key: tbx.KeyCtrlH, want: []keyboard.Key{keyboard.KeyBackspace}},
		{key: tbx.KeyTab, want: []keyboard.Key{keyboard.KeyTab}},
		{key: tbx.KeyCtrlI, want: []keyboard.Key{keyboard.KeyTab}},
		{key: tbx.KeyEnter, want: []keyboard.Key{keyboard.KeyEnter}},
		{key: tbx.KeyCtrlM, want: []keyboard.Key{keyboard.KeyEnter}},
		{key: tbx.KeyEsc, want: []keyboard.Key{keyboard.KeyEsc}},
		{key: tbx.KeyCtrlLsqBracket, want: []keyboard.Key{keyboard.KeyEsc}},
		{key: tbx.KeyCtrl3, want: []keyboard.Key{keyboard.KeyEsc}},
		{key: tbx.KeyCtrl2, want: []keyboard.Key{keyboard.KeyCtrl, '2'}},
		{key: tbx.KeyCtrlTilde, want: []keyboard.Key{keyboard.KeyCtrl, '2'}},
		{key: tbx.KeyCtrlSpace, want: []keyboard.Key{keyboard.KeyCtrl, '2'}},
		{key: tbx.KeyCtrl4, want: []keyboard.Key{keyboard.KeyCtrl, '4'}},
		{key: tbx.KeyCtrlBackslash, want: []keyboard.Key{keyboard.KeyCtrl, '4'}},
		{key: tbx.KeyCtrl5, want: []keyboard.Key{keyboard.KeyCtrl, '5'}},
		{key: tbx.KeyCtrlRsqBracket, want: []keyboard.Key{keyboard.KeyCtrl, '5'}},
		{key: tbx.KeyCtrl6, want: []keyboard.Key{keyboard.KeyCtrl, '6'}},
		{key: tbx.KeyCtrl7, want: []keyboard.Key{keyboard.KeyCtrl, '7'}},
		{key: tbx.KeyCtrlSlash, want: []keyboard.Key{keyboard.KeyCtrl, '7'}},
		{key: tbx.KeyCtrlUnderscore, want: []keyboard.Key{keyboard.KeyCtrl, '7'}},
		{key: tbx.KeyCtrl8, want: []keyboard.Key{keyboard.KeyCtrl, '8'}},
		{key: tbx.KeyCtrlA, want: []keyboard.Key{keyboard.KeyCtrl, 'a'}},
		{key: tbx.KeyCtrlB, want: []keyboard.Key{keyboard.KeyCtrl, 'b'}},
		{key: tbx.KeyCtrlC, want: []keyboard.Key{keyboard.KeyCtrl, 'c'}},
		{key: tbx.KeyCtrlD, want: []keyboard.Key{keyboard.KeyCtrl, 'd'}},
		{key: tbx.KeyCtrlE, want: []keyboard.Key{keyboard.KeyCtrl, 'e'}},
		{key: tbx.KeyCtrlF, want: []keyboard.Key{keyboard.KeyCtrl, 'f'}},
		{key: tbx.KeyCtrlG, want: []keyboard.Key{keyboard.KeyCtrl, 'g'}},
		{key: tbx.KeyCtrlJ, want: []keyboard.Key{keyboard.KeyCtrl, 'j'}},
		{key: tbx.KeyCtrlK, want: []keyboard.Key{keyboard.KeyCtrl, 'k'}},
		{key: tbx.KeyCtrlL, want: []keyboard.Key{keyboard.KeyCtrl, 'l'}},
		{key: tbx.KeyCtrlN, want: []keyboard.Key{keyboard.KeyCtrl, 'n'}},
		{key: tbx.KeyCtrlO, want: []keyboard.Key{keyboard.KeyCtrl, 'o'}},
		{key: tbx.KeyCtrlP, want: []keyboard.Key{keyboard.KeyCtrl, 'p'}},
		{key: tbx.KeyCtrlQ, want: []keyboard.Key{keyboard.KeyCtrl, 'q'}},
		{key: tbx.KeyCtrlR, want: []keyboard.Key{keyboard.KeyCtrl, 'r'}},
		{key: tbx.KeyCtrlS, want: []keyboard.Key{keyboard.KeyCtrl, 's'}},
		{key: tbx.KeyCtrlT, want: []keyboard.Key{keyboard.KeyCtrl, 't'}},
		{key: tbx.KeyCtrlU, want: []keyboard.Key{keyboard.KeyCtrl, 'u'}},
		{key: tbx.KeyCtrlV, want: []keyboard.Key{keyboard.KeyCtrl, 'v'}},
		{key: tbx.KeyCtrlW, want: []keyboard.Key{keyboard.KeyCtrl, 'w'}},
		{key: tbx.KeyCtrlX, want: []keyboard.Key{keyboard.KeyCtrl, 'x'}},
		{key: tbx.KeyCtrlY, want: []keyboard.Key{keyboard.KeyCtrl, 'y'}},
		{key: tbx.KeyCtrlZ, want: []keyboard.Key{keyboard.KeyCtrl, 'z'}},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("key:%v and ch:%v want:%v", tc.key, tc.ch, tc.want), func(t *testing.T) {
			evs := toTermdashEvents(tbx.Event{
				Type: tbx.EventKey,
				Key:  tc.key,
				Ch:   tc.ch,
			})

			gotCount := len(evs)
			var wantCount int
			if tc.wantErr {
				wantCount = 1
			} else {
				wantCount = len(tc.want)
			}

			if gotCount != wantCount {
				t.Fatalf("toTermdashEvents => got %d events, want %d, events were:\n%v", gotCount, wantCount, pretty.Sprint(evs))
			}

			for i, ev := range evs {
				if err, ok := ev.(*terminalapi.Error); ok != tc.wantErr {
					t.Fatalf("toTermdashEvents => unexpected error:%v, wantErr: %v", err, tc.wantErr)
				}
				if _, ok := ev.(*terminalapi.Error); ok {
					return
				}

				switch e := ev.(type) {
				case *terminalapi.Keyboard:
					if got, want := e.Key, tc.want[i]; got != want {
						t.Errorf("toTermdashEvents => got key[%d] %v, want %v", got, i, want)
					}

				default:
					t.Fatalf("toTermdashEvents => unexpected event type %T", e)
				}
			}
		})
	}
}
