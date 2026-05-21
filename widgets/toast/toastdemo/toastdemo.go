// Copyright 2026 Google Inc.
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

// Binary toastdemo shows animated toast notifications in normal containers and
// modal windows.
package main

import (
	"context"
	"fmt"
	"image"
	"log"
	"sync"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/borderfx"
	"github.com/mum4k/termdash/widgets/fx"
	"github.com/mum4k/termdash/widgets/modal"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/toast"
)

type toastSurface struct {
	managers []*toast.Manager
}

func (ts *toastSurface) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	for _, manager := range ts.managers {
		if err := manager.Draw(cvs, meta); err != nil {
			return err
		}
	}
	return nil
}

func (ts *toastSurface) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	_, _ = k, meta
	return nil
}

func (ts *toastSurface) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	for i := len(ts.managers) - 1; i >= 0; i-- {
		if err := ts.managers[i].Mouse(m, meta); err != nil {
			return err
		}
	}
	return nil
}

func (ts *toastSurface) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize: image.Point{X: 40, Y: 14},
		WantMouse:   widgetapi.MouseScopeWidget,
	}
}

type demoState struct {
	mu         sync.Mutex
	topRight   *toast.Manager
	topLeft    *toast.Manager
	bottomLeft *toast.Manager
	center     *toast.Manager
	custom     *toast.Manager
	modal      *toast.Manager
	log        *text.Text
	cancel     context.CancelFunc
	seq        int
}

func main() {
	term, err := tcell.New()
	if err != nil {
		log.Fatalf("failed to create terminal: %v", err)
	}
	defer term.Close()
	term.EnableMouseMotion()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, surface, err := newDemoState(cancel)
	if err != nil {
		log.Fatalf("failed to create demo widgets: %v", err)
	}

	stage, err := framedStage(surface)
	if err != nil {
		log.Fatalf("failed to create fx stage: %v", err)
	}

	controls, err := controlsWidget()
	if err != nil {
		log.Fatalf("failed to create controls widget: %v", err)
	}

	modalWidget := modalWidget(state.modal)
	root, err := container.New(
		term,
		container.ID("root"),
		container.Border(linestyle.Round),
		container.BorderTitle(" Toast Notification Showcase "),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.ID("stage"),
						container.Border(linestyle.Round),
						container.BorderTitle(" fx + toast surface "),
						container.PlaceWidget(stage),
					),
					container.Right(
						container.SplitHorizontal(
							container.Top(
								container.ID("controls"),
								container.Border(linestyle.Round),
								container.BorderTitle(" controls "),
								container.PlaceWidget(controls),
							),
							container.Bottom(
								container.ID("modal-host"),
								container.Border(linestyle.Round),
								container.BorderTitle(" modal host "),
								container.PlaceWidget(modalWidget),
							),
							container.SplitPercent(42),
						),
					),
					container.SplitPercent(66),
				),
			),
			container.Bottom(
				container.ID("event-log"),
				container.Border(linestyle.Round),
				container.BorderTitle(" event log "),
				container.PlaceWidget(state.log),
			),
			container.SplitPercent(78),
		),
	)
	if err != nil {
		log.Fatalf("failed to create root container: %v", err)
	}

	animator := borderfx.NewAnimator(root)
	animator.ApplyProfile(borderfx.Profiles.FuturisticSweep, "root", "stage", "controls", "modal-host", "event-log")
	animator.SetAlwaysActive(true)
	go func() {
		if err := animator.Run(ctx); err != nil && err != context.Canceled {
			log.Printf("borderfx animator stopped: %v", err)
		}
	}()

	state.seed()
	go state.heartbeat(ctx)

	if err := termdash.Run(
		ctx,
		term,
		root,
		termdash.KeyboardSubscriber(state.handleKeyboard),
		termdash.RedrawInterval(50*time.Millisecond),
	); err != nil && err != context.Canceled {
		log.Fatalf("termdash run failed: %v", err)
	}
}

func newDemoState(cancel context.CancelFunc) (*demoState, *toastSurface, error) {
	topRight, err := newToastManager(
		toast.Anchor(toast.PlacementTopRight),
		toast.SlideDirection(toast.DirectionRight),
	)
	if err != nil {
		return nil, nil, err
	}
	topLeft, err := newToastManager(
		toast.Anchor(toast.PlacementTopLeft),
		toast.SlideDirection(toast.DirectionLeft),
	)
	if err != nil {
		return nil, nil, err
	}
	bottomLeft, err := newToastManager(
		toast.Anchor(toast.PlacementBottomLeft),
		toast.SlideDirection(toast.DirectionBottom),
	)
	if err != nil {
		return nil, nil, err
	}
	center, err := newToastManager(
		toast.Anchor(toast.PlacementCenter),
		toast.AnimationMode(toast.AnimationPop),
		toast.Width(34),
	)
	if err != nil {
		return nil, nil, err
	}
	custom, err := newToastManager(
		toast.CustomPosition(func(ar image.Rectangle, size image.Point, index int) image.Point {
			return image.Point{
				X: ar.Min.X + 4,
				Y: ar.Min.Y + 3 + index*(size.Y+1),
			}
		}),
		toast.AnimationMode(toast.AnimationFade),
		toast.Width(36),
		toast.SeverityStyle(toast.SeverityNeutral, toast.Style{
			Icon:   '✦',
			Accent: cell.ColorNumber(213),
			BorderCellOpts: []cell.Option{
				cell.FgColor(cell.ColorNumber(213)),
				cell.Bold(),
			},
			TitleCellOpts: []cell.Option{
				cell.FgColor(cell.ColorNumber(231)),
				cell.Bold(),
			},
		}),
	)
	if err != nil {
		return nil, nil, err
	}
	modalToasts, err := newToastManager(
		toast.Anchor(toast.PlacementBottomRight),
		toast.Width(30),
		toast.MaxVisible(3),
		toast.SlideDirection(toast.DirectionRight),
	)
	if err != nil {
		return nil, nil, err
	}

	logWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		return nil, nil, err
	}

	state := &demoState{
		topRight:   topRight,
		topLeft:    topLeft,
		bottomLeft: bottomLeft,
		center:     center,
		custom:     custom,
		modal:      modalToasts,
		log:        logWidget,
		cancel:     cancel,
	}
	surface := &toastSurface{
		managers: []*toast.Manager{
			topRight,
			topLeft,
			bottomLeft,
			center,
			custom,
		},
	}
	return state, surface, nil
}

func newToastManager(opts ...toast.Option) (*toast.Manager, error) {
	base := []toast.Option{
		toast.Width(42),
		toast.MinWidth(18),
		toast.MaxWidth(48),
		toast.MaxVisible(4),
		toast.Margin(2, 1),
		toast.Gap(1),
		toast.DefaultTTL(6 * time.Second),
		toast.AnimationDuration(420 * time.Millisecond),
		toast.Border(linestyle.Round, cell.FgColor(cell.ColorNumber(244))),
		toast.FillCellOpts(cell.BgColor(cell.ColorNumber(16))),
		toast.TitleCellOpts(cell.FgColor(cell.ColorNumber(231)), cell.Bold()),
		toast.MessageCellOpts(cell.FgColor(cell.ColorNumber(252))),
		toast.ActionCellOpts(cell.FgColor(cell.ColorNumber(159)), cell.Bold()),
		toast.Shadow(true, cell.BgColor(cell.ColorNumber(235))),
	}
	base = append(base, opts...)
	return toast.New(base...)
}

func framedStage(surface *toastSurface) (widgetapi.Widget, error) {
	framed, err := fx.FramedNew(
		surface,
		fx.FramedLineStyle(linestyle.Round),
		fx.FramedTitle(" live notification deck ", cell.FgColor(cell.ColorNumber(159)), cell.Bold()),
		fx.FramedBorderOpts(cell.FgColor(cell.ColorNumber(51))),
	)
	if err != nil {
		return nil, err
	}
	return fx.New(
		framed,
		fx.FadeIn(600*time.Millisecond),
		fx.SweepRight(450*time.Millisecond),
	)
}

func controlsWidget() (*text.Text, error) {
	controls, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, err
	}
	err = controls.Write(
		"Keys\n" +
			"1  info / top-right slide\n" +
			"2  success / action labels\n" +
			"3  warning / bottom-left\n" +
			"4  error / top-left sticky\n" +
			"5  center pop notification\n" +
			"6  progress bar toast\n" +
			"7  custom positioned toast\n" +
			"8  modal-window toast\n" +
			"9  burst all placements\n" +
			"c  clear notifications\n" +
			"q  quit\n\n" +
			"Click visible non-modal toasts to dismiss them. Drag or minimize the modal window to check the hosted toast stack.",
	)
	if err != nil {
		return nil, err
	}
	return controls, nil
}

func modalWidget(manager *toast.Manager) *modal.Modal {
	opts := modal.NewOptions(
		modal.Border(true),
		modal.MinimumSize(image.Point{X: 24, Y: 10}),
		modal.TitleBarCellOpts(cell.BgColor(cell.ColorNumber(236)), cell.FgColor(cell.ColorNumber(231))),
		modal.TitleCellOpts(cell.FgColor(cell.ColorNumber(159)), cell.Bold()),
		modal.TitleControlCellOpts(cell.FgColor(cell.ColorNumber(228)), cell.Bold()),
	)
	item := modal.NewDraggableWidget("modal-toasts", manager, 3, 2, 42, 14, opts)
	item.Title = "Modal Toast Deck"
	return modal.NewModal("toast-modal", []*modal.DraggableWidget{item}, opts)
}

func (s *demoState) handleKeyboard(k *terminalapi.Keyboard) {
	switch k.Key {
	case '1':
		id := s.topRight.Notify("Deploy queued", "Top-right slide-in toast with the default informational styling.", toast.WithSeverity(toast.SeverityInfo))
		s.note("top-right info toast %s", id)
	case '2':
		id := s.topRight.Notify("Snapshot saved", "Success toast with compact action labels and a crisp border.", toast.WithSeverity(toast.SeveritySuccess), toast.WithActions("open", "copy id"))
		s.note("success toast %s with actions", id)
	case '3':
		id := s.bottomLeft.Notify("Latency rising", "Bottom-left warning anchored independently from the main stack.", toast.WithSeverity(toast.SeverityWarning))
		s.note("bottom-left warning toast %s", id)
	case '4':
		id := s.topLeft.Notify("Access denied", "Sticky top-left error. Click it or press c to clear.", toast.WithSeverity(toast.SeverityError), toast.Sticky())
		s.note("sticky error toast %s", id)
	case '5':
		id := s.center.Notify("Center stage", "Pop animation with a centered placement for focused callouts.", toast.WithSeverity(toast.SeverityNeutral), toast.WithIcon('◈'))
		s.note("center pop toast %s", id)
	case '6':
		step := float64((s.nextSeq()%8)+1) / 8
		id := s.topRight.Notify("Indexing assets", "Progress toasts use the same manager API.", toast.WithSeverity(toast.SeverityInfo), toast.WithProgress(step), toast.WithActions("pause"))
		s.note("progress toast %s at %.0f%%", id, step*100)
	case '7':
		id := s.custom.Notify("Custom anchor", "This notification uses a caller-supplied PositionFunc and a custom severity style.", toast.WithSeverity(toast.SeverityNeutral))
		s.note("custom-position toast %s", id)
	case '8':
		id := s.modal.Notify("Modal signal", "Toast manager rendered inside a draggable modal window.", toast.WithSeverity(toast.SeveritySuccess), toast.WithActions("dock", "trace"))
		s.note("modal toast %s", id)
	case '9':
		s.burst()
	case 'c', 'C':
		s.clear()
	case 'q', 'Q', keyboard.KeyEsc:
		s.cancel()
	}
}

func (s *demoState) seed() {
	s.topRight.Notify("Toast manager online", "Non-modal surface is wrapped with fx and hosted inside animated borderfx containers.", toast.WithSeverity(toast.SeverityInfo), toast.WithActions("inspect"))
	s.bottomLeft.Notify("Click-to-dismiss enabled", "Mouse events are handled by the toast widget in normal containers.", toast.WithSeverity(toast.SeveritySuccess))
	s.modal.Notify("Modal ready", "This stack lives inside a draggable modal window.", toast.WithSeverity(toast.SeverityNeutral), toast.WithIcon('◇'))
	s.note("demo booted")
}

func (s *demoState) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(7 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			id := s.topRight.Notify("Heartbeat", "Automatic toast proving redraw, TTL, and animation timing.", toast.WithSeverity(toast.SeverityInfo))
			s.note("heartbeat toast %s", id)
		}
	}
}

func (s *demoState) burst() {
	s.topRight.Notify("North-east", "Slide from the right.", toast.WithSeverity(toast.SeverityInfo))
	s.topLeft.Notify("North-west", "Slide from the left.", toast.WithSeverity(toast.SeverityError))
	s.bottomLeft.Notify("South-west", "Slide from the bottom.", toast.WithSeverity(toast.SeverityWarning))
	s.center.Notify("Center burst", "Pop animation.", toast.WithSeverity(toast.SeverityNeutral), toast.WithIcon('✹'))
	s.custom.Notify("Custom rail", "Caller-positioned stack.", toast.WithSeverity(toast.SeverityNeutral))
	s.modal.Notify("Modal burst", "Hosted in the draggable window.", toast.WithSeverity(toast.SeveritySuccess))
	s.note("burst all placements")
}

func (s *demoState) clear() {
	s.topRight.Clear()
	s.topLeft.Clear()
	s.bottomLeft.Clear()
	s.center.Clear()
	s.custom.Clear()
	s.modal.Clear()
	s.note("cleared all toast managers")
}

func (s *demoState) note(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	_ = s.log.Write(time.Now().Format("15:04:05") + "  " + line + "\n")
}

func (s *demoState) nextSeq() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	return s.seq
}
