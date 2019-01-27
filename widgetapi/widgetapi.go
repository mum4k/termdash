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

// Package widgetapi defines the API of a widget on the dashboard.
package widgetapi

import (
	"image"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/terminalapi"
)

// Options contains registration options for a widget.
// This is how the widget indicates its needs to the infrastructure.
type Options struct {
	// Ratio allows a widget to request a canvas whose size will always have
	// the specified ratio of width:height (Ratio.X:Ratio.Y).
	// The zero value i.e. image.Point{0, 0} indicates that the widget accepts
	// canvas of any ratio.
	Ratio image.Point

	// MinimumSize allows a widget to specify the smallest allowed canvas size.
	// If the terminal size and/or splits cause the assigned canvas to be
	// smaller than this, the widget will be skipped. I.e. The Draw() method
	// won't be called until a resize above the specified minimum.
	MinimumSize image.Point

	// MaximumSize allows a widget to specify the largest allowed canvas size.
	// If the terminal size and/or splits cause the assigned canvas to be larger
	// than this, the widget will only receive a canvas of this size within its
	// container. Setting any of the two coordinates to zero indicates
	// unlimited.
	MaximumSize image.Point

	// WantKeyboard allows a widget to request keyboard events.
	// If false, keyboard events won't be forwarded to the widget.
	// If true, the widget receives keyboard events if its container is
	// focused.
	WantKeyboard bool

	// WantMouse allows a widget to request mouse events.
	// If false, mouse events won't be forwarded to the widget.
	// If true, the widget receives all mouse events whose coordinates fall
	// within its canvas.
	WantMouse bool
}

// Widget is a single widget on the dashboard.
// Implementations must be thread safe.
type Widget interface {
	// When the infrastructure calls Draw(), the widget must block on the call
	// until it finishes drawing onto the provided canvas. When given the
	// canvas, the widget must first determine its size by calling
	// Canvas.Size(), then limit all its drawing to this area.
	//
	// The widget must not assume that the size of the canvas or its content
	// remains the same between calls.
	Draw(cvs *canvas.Canvas) error

	// Keyboard is called when the widget is focused on the dashboard and a key
	// shortcut the widget registered for was pressed. Only called if the widget
	// registered for keyboard events.
	Keyboard(k *terminalapi.Keyboard) error

	// Mouse is called when the widget is focused on the dashboard and a mouse
	// event happens on its canvas. Only called if the widget registered for mouse
	// events.
	Mouse(m *terminalapi.Mouse) error

	// Options returns registration options for the widget.
	// This is how the widget indicates to the infrastructure whether it is
	// interested in keyboard or mouse shortcuts, what is its minimum canvas
	// size, etc.
	//
	// Most widgets will return statically compiled options (minimum and
	// maximum size, etc.). If the returned options depend on the runtime state
	// of the widget (e.g. the user data provided to the widget), the widget
	// must protect against a case where the infrastructure calls the Draw
	// method with a canvas that doesn't meet the requested options. This is
	// because the data in the widget might change between calls to Options and
	// Draw.
	Options() Options
}
