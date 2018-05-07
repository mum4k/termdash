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

/*
Package termdash implements a terminal based dashboard.

While running, the terminal dashboard performs the following:
  - Periodic redrawing of the canvas and all the widgets.
  - Event based redrawing of the widgets (i.e. on Keyboard or Mouse events).
  - Forwards input events to widgets and optional subscribers.
  - Handles terminal resize events.
*/
package termdash

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/terminalapi"
)

// DefaultRedrawInterval is the default for the RedrawInterval option.
const DefaultRedrawInterval = 250 * time.Millisecond

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(td *termdash)
}

// option implements Option.
type option func(td *termdash)

// set implements Option.set.
func (o option) set(td *termdash) {
	o(td)
}

// RedrawInterval sets how often termdash redraws the container and all the widgets.
// Defaults to DefaultRedrawInterval.
func RedrawInterval(t time.Duration) Option {
	return option(func(td *termdash) {
		td.redrawInterval = t
	})
}

// ErrorHandler is used to provide a function that will be called with all
// errors that occur while the dashboard is running. If not provided, any
// errors panic the application.
func ErrorHandler(f func(error)) Option {
	return option(func(td *termdash) {
		td.errorHandler = f
	})
}

// KeyboardSubscriber registers a subscriber for Keyboard events. Each
// keyboard event is forwarded to the container and the registered subscriber.
// The provided function must be non-blocking, ideally just storing the value
// and returning as termdash blocks on each subscriber.
func KeyboardSubscriber(f func(*terminalapi.Keyboard)) Option {
	return option(func(td *termdash) {
		td.keyboardSubscriber = f
	})
}

// MouseSubscriber registers a subscriber for Mouse events. Each mouse event
// is forwarded to the container and the registered subscriber.
// The provided function must be non-blocking, ideally just storing the value
// and returning as termdash blocks on each subscriber.
func MouseSubscriber(f func(*terminalapi.Mouse)) Option {
	return option(func(td *termdash) {
		td.mouseSubscriber = f
	})
}

// Run runs the terminal dashboard with the provided container on the terminal.
// Blocks until the context expires.
func Run(ctx context.Context, t terminalapi.Terminal, c *container.Container, opts ...Option) error {
	td := newTermdash(t, c, opts...)
	defer td.stop()

	return td.start(ctx)
}

// termdash is a terminal based dashboard.
// This object is thread-safe.
type termdash struct {
	// term is the terminal the dashboard runs on.
	term terminalapi.Terminal

	// container maintains terminal splits and places widgets.
	container *container.Container

	// closeCh gets closed when Stop() is called, which tells the event
	// collecting goroutine to exit.
	closeCh chan struct{}
	// exitCh gets closed when the event collecting goroutine actually exits.
	exitCh chan struct{}

	// clearNeeded indicates if the terminal needs to be cleared next time
	// we're drawing it. Terminal needs to be cleared if its sized changed.
	clearNeeded bool

	// mu protects termdash.
	mu sync.Mutex

	// Options.
	redrawInterval     time.Duration
	errorHandler       func(error)
	mouseSubscriber    func(*terminalapi.Mouse)
	keyboardSubscriber func(*terminalapi.Keyboard)
}

// newTermdash creates a new termdash.
func newTermdash(t terminalapi.Terminal, c *container.Container, opts ...Option) *termdash {
	td := &termdash{
		term:           t,
		container:      c,
		closeCh:        make(chan struct{}),
		exitCh:         make(chan struct{}),
		redrawInterval: DefaultRedrawInterval,
	}

	for _, opt := range opts {
		opt.set(td)
	}
	return td
}

// handleError forwards the error to the error handler if one was
// provided or panics.
func (td *termdash) handleError(err error) {
	if td.errorHandler != nil {
		td.errorHandler(err)
	} else {
		panic(err)
	}
}

// setClearNeeded flags that the terminal needs to be cleared next time we're
// drawing it.
func (td *termdash) setClearNeeded() {
	td.mu.Lock()
	defer td.mu.Unlock()
	td.clearNeeded = true
}

// redraw redraws the container and its widgets.
// The caller must hold td.mu.
func (td *termdash) redraw() error {
	if td.clearNeeded {
		if err := td.term.Clear(); err != nil {
			return fmt.Errorf("term.Clear => error: %v", err)
		}
		td.clearNeeded = false
	}

	if err := td.container.Draw(); err != nil {
		return fmt.Errorf("container.Draw => error: %v", err)
	}

	if err := td.term.Flush(); err != nil {
		return fmt.Errorf("term.Flush => error: %v", err)
	}
	return nil
}

// keyEvRedraw forwards the keyboard event and redraws the container and its
// widgets.
func (td *termdash) keyEvRedraw(ev *terminalapi.Keyboard) error {
	td.mu.Lock()
	defer td.mu.Unlock()

	if err := td.container.Keyboard(ev); err != nil {
		return err
	}
	if td.keyboardSubscriber != nil {
		td.keyboardSubscriber(ev)
	}
	return td.redraw()
}

// mouseEvRedraw forwards the mouse event and redraws the container and its
// widgets.
func (td *termdash) mouseEvRedraw(ev *terminalapi.Mouse) error {
	td.mu.Lock()
	defer td.mu.Unlock()

	if err := td.container.Mouse(ev); err != nil {
		return err
	}
	if td.mouseSubscriber != nil {
		td.mouseSubscriber(ev)
	}
	return td.redraw()
}

// periodicRedraw is called once each RedrawInterval.
func (td *termdash) periodicRedraw() error {
	td.mu.Lock()
	defer td.mu.Unlock()
	return td.redraw()
}

// processEvents processes terminal input events.
// This is the body of the event collecting goroutine.
func (td *termdash) processEvents(ctx context.Context) {
	defer close(td.exitCh)

	for {
		event := td.term.Event(ctx)
		switch ev := event.(type) {
		case *terminalapi.Keyboard:
			if err := td.keyEvRedraw(ev); err != nil {
				td.handleError(err)
			}

		case *terminalapi.Mouse:
			if err := td.mouseEvRedraw(ev); err != nil {
				td.handleError(err)
			}

		case *terminalapi.Resize:
			td.setClearNeeded()

		case *terminalapi.Error:
			// Don't forward the error if the context is closed.
			// It just says that the context expired.
			select {
			case <-ctx.Done():
			default:
				td.handleError(ev.Error())
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

// start starts the terminal dashboard. Blocks until the context expires or
// until stop() is called.
func (td *termdash) start(ctx context.Context) error {
	redrawTimer := time.NewTicker(td.redrawInterval)
	defer redrawTimer.Stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// stops when stop() is called or the context expires.
	go td.processEvents(ctx)

	for {
		select {
		case <-redrawTimer.C:
			if err := td.periodicRedraw(); err != nil {
				return err
			}

		case <-ctx.Done():
			return nil

		case <-td.closeCh:
			return nil
		}
	}
}

// stop tells the event collecting goroutine to stop.
// Blocks until it exits.
func (td *termdash) stop() {
	close(td.closeCh)
	<-td.exitCh
}
