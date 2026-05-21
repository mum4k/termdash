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

// Package toast provides animated notification stacks.
//
// A Manager is a normal termdash widget, so it can be placed in a regular
// container, embedded in a modal window, wrapped with widgets/fx effects, or
// hosted by a container whose border is animated by widgets/borderfx.
package toast

import (
	"fmt"
	"image"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Notification is a single toast entry.
type Notification struct {
	// ID uniquely identifies the notification. A stable ID is generated when
	// empty.
	ID string
	// Title is the first, emphasized line.
	Title string
	// Message is wrapped underneath the title.
	Message string
	// Severity selects the visual style.
	Severity Severity
	// TTL controls how long the notification remains visible. Zero uses the
	// manager default; negative values make the notification sticky.
	TTL time.Duration
	// Created is the timestamp used for TTL and entrance animation. The current
	// manager clock is used when Created is zero.
	Created time.Time
	// Icon overrides the severity icon when non-zero.
	Icon rune
	// Progress is the optional progress value in the range [0, 1].
	Progress float64
	// ShowProgress controls whether Progress is drawn.
	ShowProgress bool
	// Actions are compact labels drawn at the bottom of the toast.
	Actions []Action
}

// Action is a clickable label shown in a notification footer.
type Action struct {
	// Label is drawn between square brackets.
	Label string
	// Callback runs when the label is clicked. A nil callback renders as a
	// visual-only action.
	Callback func() error
	// Dismiss removes the notification after a successful click.
	Dismiss bool
}

// NotificationOption configures a notification passed to Notify.
type NotificationOption interface {
	setNotification(*Notification)
}

type notificationOption func(*Notification)

func (no notificationOption) setNotification(n *Notification) {
	no(n)
}

// WithID sets the notification ID.
func WithID(id string) NotificationOption {
	return notificationOption(func(n *Notification) {
		n.ID = id
	})
}

// WithSeverity sets the notification severity.
func WithSeverity(sev Severity) NotificationOption {
	return notificationOption(func(n *Notification) {
		n.Severity = sev
	})
}

// WithTTL sets the notification lifetime.
func WithTTL(ttl time.Duration) NotificationOption {
	return notificationOption(func(n *Notification) {
		n.TTL = ttl
	})
}

// Sticky makes the notification stay visible until dismissed.
func Sticky() NotificationOption {
	return notificationOption(func(n *Notification) {
		n.TTL = -1
	})
}

// WithIcon sets a notification-specific icon.
func WithIcon(icon rune) NotificationOption {
	return notificationOption(func(n *Notification) {
		n.Icon = icon
	})
}

// WithProgress shows a progress bar clamped to the range [0, 1].
func WithProgress(progress float64) NotificationOption {
	return notificationOption(func(n *Notification) {
		n.Progress = progress
		n.ShowProgress = true
	})
}

// WithActions sets compact action labels for the notification footer.
func WithActions(actions ...string) NotificationOption {
	return notificationOption(func(n *Notification) {
		n.Actions = make([]Action, 0, len(actions))
		for _, label := range actions {
			n.Actions = append(n.Actions, Action{Label: label})
		}
	})
}

// WithAction appends a clickable action label to the notification footer.
func WithAction(label string, callback func() error) NotificationOption {
	return notificationOption(func(n *Notification) {
		n.Actions = append(n.Actions, Action{Label: label, Callback: callback})
	})
}

// WithActionValues replaces the notification footer with prepared actions.
func WithActionValues(actions ...Action) NotificationOption {
	return notificationOption(func(n *Notification) {
		n.Actions = cloneActions(actions)
	})
}

// WithCreated sets the notification creation time.
func WithCreated(created time.Time) NotificationOption {
	return notificationOption(func(n *Notification) {
		n.Created = created
	})
}

// Manager renders and manages a stack of toast notifications.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Manager struct {
	mu            sync.Mutex
	notifications []Notification
	nextID        int
	lastRects     map[string]image.Rectangle
	lastActions   []actionHit
	opts          *options
}

type actionHit struct {
	id       string
	rect     image.Rectangle
	callback func() error
	dismiss  bool
}

// New returns a new notification manager.
func New(opts ...Option) (*Manager, error) {
	o := newOptions(opts...)
	if err := o.validate(); err != nil {
		return nil, err
	}
	return &Manager{
		lastRects: map[string]image.Rectangle{},
		opts:      o,
	}, nil
}

// Notify creates a notification from the provided title and message.
func (m *Manager) Notify(title, message string, opts ...NotificationOption) string {
	n := Notification{
		Title:   title,
		Message: message,
	}
	for _, opt := range opts {
		opt.setNotification(&n)
	}
	return m.Push(n)
}

// Push adds a prepared notification and returns its ID.
func (m *Manager) Push(n Notification) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextID++
	if n.ID == "" {
		n.ID = fmt.Sprintf("toast-%d", m.nextID)
	}
	if n.Created.IsZero() {
		n.Created = m.opts.clock()
	}
	if n.TTL == 0 {
		n.TTL = m.opts.defaultTTL
	}
	if n.ShowProgress {
		n.Progress = clampFloat(n.Progress, 0, 1)
	}
	n.Actions = cloneActions(n.Actions)
	m.notifications = append(m.notifications, n)
	return n.ID
}

// Dismiss removes a notification by ID.
func (m *Manager) Dismiss(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.dismissLocked(id)
}

// Clear removes all notifications.
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifications = nil
	m.lastRects = map[string]image.Rectangle{}
	m.lastActions = nil
}

// Count returns the number of notifications currently held by the manager.
func (m *Manager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.notifications)
}

// Draw implements widgetapi.Widget.
func (m *Manager) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_ = meta
	ar := cvs.Area()
	if ar.Dx() < m.opts.minimumSize.X || ar.Dy() < m.opts.minimumSize.Y {
		return draw.ResizeNeeded(cvs)
	}

	now := m.opts.clock()
	m.expireLocked(now)
	visible := m.visibleLocked()
	width := m.toastWidth(ar)
	rects := m.layout(ar, visible, width)
	m.lastRects = map[string]image.Rectangle{}
	m.lastActions = nil

	for i, n := range visible {
		progress := m.animationProgress(n, now)
		rect := m.animateRect(ar, rects[i], progress)
		if m.opts.animation == AnimationPop {
			rect = popRect(rects[i], progress)
		}
		hits, err := m.drawNotification(cvs, rect, n, progress)
		if err != nil {
			return err
		}
		m.lastRects[n.ID] = rect
		m.lastActions = append(m.lastActions, hits...)
	}
	return nil
}

// Keyboard implements widgetapi.Widget.
func (m *Manager) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	_, _ = k, meta
	return nil
}

// Mouse implements widgetapi.Widget.
func (m *Manager) Mouse(event *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	_, _ = meta, event
	if event == nil || event.Button != mouse.ButtonLeft {
		return nil
	}

	m.mu.Lock()
	var callback func() error
	for _, hit := range m.lastActions {
		if !event.Position.In(hit.rect) {
			continue
		}
		callback = hit.callback
		if hit.dismiss {
			m.dismissLocked(hit.id)
		}
		break
	}
	if callback == nil && m.opts.dismissOnClick {
		for id, rect := range m.lastRects {
			if event.Position.In(rect) {
				m.dismissLocked(id)
				break
			}
		}
	}
	m.mu.Unlock()

	if callback != nil {
		return callback()
	}
	return nil
}

// Options implements widgetapi.Widget.
func (m *Manager) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize: m.opts.minimumSize,
		WantMouse:   widgetapi.MouseScopeWidget,
	}
}

func (m *Manager) dismissLocked(id string) bool {
	for i, n := range m.notifications {
		if n.ID == id {
			m.notifications = append(m.notifications[:i], m.notifications[i+1:]...)
			delete(m.lastRects, id)
			m.removeActionHitsLocked(id)
			return true
		}
	}
	return false
}

func (m *Manager) removeActionHitsLocked(id string) {
	kept := m.lastActions[:0]
	for _, hit := range m.lastActions {
		if hit.id != id {
			kept = append(kept, hit)
		}
	}
	m.lastActions = kept
}

func (m *Manager) expireLocked(now time.Time) {
	kept := m.notifications[:0]
	for _, n := range m.notifications {
		if n.TTL > 0 && !n.Created.IsZero() && now.Sub(n.Created) >= n.TTL {
			continue
		}
		kept = append(kept, n)
	}
	m.notifications = kept
}

func (m *Manager) visibleLocked() []Notification {
	limit := m.opts.maxVisible
	if limit > len(m.notifications) {
		limit = len(m.notifications)
	}
	visible := make([]Notification, 0, limit)
	if m.opts.newestFirst {
		for i := len(m.notifications) - 1; i >= 0 && len(visible) < limit; i-- {
			visible = append(visible, m.notifications[i])
		}
		return visible
	}
	for i := 0; i < len(m.notifications) && len(visible) < limit; i++ {
		visible = append(visible, m.notifications[i])
	}
	return visible
}

func (m *Manager) toastWidth(ar image.Rectangle) int {
	available := ar.Dx() - 2*m.opts.margin.X
	if available < 1 {
		available = ar.Dx()
	}
	width := m.opts.width
	if width > m.opts.maxWidth {
		width = m.opts.maxWidth
	}
	if width > available {
		width = available
	}
	if width < m.opts.minWidth && available >= m.opts.minWidth {
		width = m.opts.minWidth
	}
	if width < 1 {
		width = 1
	}
	return width
}

func (m *Manager) layout(ar image.Rectangle, visible []Notification, width int) []image.Rectangle {
	rects := make([]image.Rectangle, 0, len(visible))
	offset := 0
	stack := m.resolvedStackDirection()
	totalHeight := m.totalHeight(visible, width)

	for i, n := range visible {
		size := image.Point{X: width, Y: m.toastHeight(n, width)}
		var p image.Point
		if m.opts.placement == PlacementCustom {
			p = m.opts.customPosition(ar, size, i)
		} else {
			p = m.anchorPoint(ar, size, stack, offset, totalHeight)
		}
		rects = append(rects, image.Rectangle{Min: p, Max: p.Add(size)})
		offset += size.Y + m.opts.gap
	}
	return rects
}

func (m *Manager) totalHeight(notifications []Notification, width int) int {
	if len(notifications) == 0 {
		return 0
	}
	total := 0
	for _, n := range notifications {
		total += m.toastHeight(n, width)
	}
	return total + m.opts.gap*(len(notifications)-1)
}

func (m *Manager) resolvedStackDirection() StackDirection {
	if m.opts.stackDirection != StackAuto {
		return m.opts.stackDirection
	}
	switch m.opts.placement {
	case PlacementBottomLeft, PlacementBottomRight:
		return StackUp
	default:
		return StackDown
	}
}

func (m *Manager) anchorPoint(ar image.Rectangle, size image.Point, stack StackDirection, offset, totalHeight int) image.Point {
	x := ar.Max.X - m.opts.margin.X - size.X
	switch m.opts.placement {
	case PlacementTopLeft, PlacementBottomLeft:
		x = ar.Min.X + m.opts.margin.X
	case PlacementCenter:
		x = ar.Min.X + (ar.Dx()-size.X)/2
	}

	if m.opts.placement == PlacementCenter {
		return image.Point{
			X: x,
			Y: ar.Min.Y + (ar.Dy()-totalHeight)/2 + offset,
		}
	}

	if stack == StackUp {
		return image.Point{
			X: x,
			Y: ar.Max.Y - m.opts.margin.Y - offset - size.Y,
		}
	}
	return image.Point{
		X: x,
		Y: ar.Min.Y + m.opts.margin.Y + offset,
	}
}

func (m *Manager) toastHeight(n Notification, width int) int {
	inset := m.borderInset()
	contentWidth := width - inset*2 - 2
	if contentWidth < 1 {
		contentWidth = 1
	}

	height := inset*2 + 2 // vertical padding
	height++              // title
	height += len(wrapLines(n.Message, contentWidth, m.opts.maxMessageLines))
	if n.ShowProgress {
		height++
	}
	if len(n.Actions) > 0 {
		height++
	}
	if height < 3 {
		height = 3
	}
	return height
}

func (m *Manager) borderInset() int {
	if m.opts.border == linestyle.None {
		return 0
	}
	return 1
}

func (m *Manager) animationProgress(n Notification, now time.Time) float64 {
	if m.opts.animation == AnimationNone || m.opts.animationDuration <= 0 || n.Created.IsZero() {
		return 1
	}
	elapsed := now.Sub(n.Created)
	if elapsed <= 0 {
		return 0
	}
	return clampFloat(float64(elapsed)/float64(m.opts.animationDuration), 0, 1)
}

func (m *Manager) animateRect(ar image.Rectangle, rect image.Rectangle, progress float64) image.Rectangle {
	if m.opts.animation != AnimationSlide || progress >= 1 {
		return rect
	}
	eased := easeOutCubic(progress)
	switch m.opts.slideDirection {
	case DirectionLeft:
		offset := int(math.Round(float64(ar.Min.X-rect.Min.X-rect.Dx()) * (1 - eased)))
		return rect.Add(image.Point{X: offset})
	case DirectionTop:
		offset := int(math.Round(float64(ar.Min.Y-rect.Min.Y-rect.Dy()) * (1 - eased)))
		return rect.Add(image.Point{Y: offset})
	case DirectionBottom:
		offset := int(math.Round(float64(ar.Max.Y-rect.Min.Y) * (1 - eased)))
		return rect.Add(image.Point{Y: offset})
	default:
		offset := int(math.Round(float64(ar.Max.X-rect.Min.X) * (1 - eased)))
		return rect.Add(image.Point{X: offset})
	}
}

func popRect(rect image.Rectangle, progress float64) image.Rectangle {
	if progress >= 1 {
		return rect
	}
	eased := easeOutCubic(progress)
	w := int(math.Round(float64(rect.Dx()) * (0.55 + 0.45*eased)))
	h := int(math.Round(float64(rect.Dy()) * (0.55 + 0.45*eased)))
	if w < 3 {
		w = 3
	}
	if h < 3 {
		h = 3
	}
	cx := rect.Min.X + rect.Dx()/2
	cy := rect.Min.Y + rect.Dy()/2
	return image.Rect(cx-w/2, cy-h/2, cx-w/2+w, cy-h/2+h)
}

func (m *Manager) drawNotification(cvs *canvas.Canvas, rect image.Rectangle, n Notification, progress float64) ([]actionHit, error) {
	style := m.style(n)
	fade := m.opts.animation == AnimationFade && progress < 1
	fillOpts := mergeCellOpts(m.opts.fillCellOpts, style.FillCellOpts, fadeOpts(fade))
	borderOpts := mergeCellOpts(m.opts.borderCellOpts, []cell.Option{cell.FgColor(style.Accent)}, style.BorderCellOpts, fadeOpts(fade))
	titleOpts := mergeCellOpts(m.opts.titleCellOpts, style.TitleCellOpts, fadeOpts(fade))
	iconOpts := mergeCellOpts([]cell.Option{cell.FgColor(style.Accent), cell.Bold()}, style.IconCellOpts, fadeOpts(fade))
	messageOpts := mergeCellOpts(m.opts.messageCellOpts, style.MessageCellOpts, fadeOpts(fade))
	actionOpts := mergeCellOpts(m.opts.actionCellOpts, fadeOpts(fade))
	progressOpts := mergeCellOpts([]cell.Option{cell.FgColor(style.Accent)}, style.ProgressCellOpts, fadeOpts(fade))

	if m.opts.shadow {
		if err := fillRect(cvs, rect.Add(image.Point{X: 1, Y: 1}), ' ', m.opts.shadowCellOpts); err != nil {
			return nil, err
		}
	}
	if err := fillRect(cvs, rect, ' ', fillOpts); err != nil {
		return nil, err
	}
	if m.opts.border != linestyle.None {
		if err := drawBorder(cvs, rect, m.opts.border, borderOpts); err != nil {
			return nil, err
		}
	}

	inset := m.borderInset()
	x := rect.Min.X + inset + 1
	maxX := rect.Max.X - inset - 1
	y := rect.Min.Y + inset + 1
	if maxX <= x {
		return nil, nil
	}

	title := n.Title
	if title == "" {
		title = "Notification"
	}
	icon := n.Icon
	if icon == 0 {
		icon = style.Icon
	}
	if icon != 0 {
		if err := writeText(cvs, string(icon), image.Point{X: x, Y: y}, maxX, iconOpts); err != nil {
			return nil, err
		}
		x += 2
	}
	if err := writeText(cvs, title, image.Point{X: x, Y: y}, maxX, titleOpts); err != nil {
		return nil, err
	}
	y++

	contentWidth := maxX - (rect.Min.X + inset + 1)
	for _, line := range wrapLines(n.Message, contentWidth, m.opts.maxMessageLines) {
		if err := writeText(cvs, line, image.Point{X: rect.Min.X + inset + 1, Y: y}, maxX, messageOpts); err != nil {
			return nil, err
		}
		y++
	}

	if n.ShowProgress {
		if err := drawProgress(cvs, image.Point{X: rect.Min.X + inset + 1, Y: y}, maxX, n.Progress, progressOpts); err != nil {
			return nil, err
		}
		y++
	}
	if len(n.Actions) > 0 {
		return drawActions(cvs, n.ID, image.Point{X: rect.Min.X + inset + 1, Y: y}, maxX, n.Actions, actionOpts)
	}
	return nil, nil
}

func (m *Manager) style(n Notification) Style {
	style, ok := m.opts.styles[n.Severity]
	if !ok {
		style = m.opts.styles[SeverityInfo]
	}
	if style.Accent == cell.ColorDefault {
		style.Accent = cell.ColorNumber(81)
	}
	return style
}

func fillRect(cvs *canvas.Canvas, rect image.Rectangle, r rune, opts []cell.Option) error {
	ar := cvs.Area()
	rect = rect.Intersect(ar)
	if rect.Empty() {
		return nil
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if _, err := cvs.SetCell(image.Point{X: x, Y: y}, r, opts...); err != nil {
				return err
			}
		}
	}
	return nil
}

type borderRunes struct {
	h  rune
	v  rune
	tl rune
	tr rune
	bl rune
	br rune
}

func drawBorder(cvs *canvas.Canvas, rect image.Rectangle, lineStyle linestyle.LineStyle, opts []cell.Option) error {
	if rect.Dx() < 2 || rect.Dy() < 2 {
		return nil
	}
	br, ok := borderChars(lineStyle)
	if !ok {
		return nil
	}
	for x := rect.Min.X; x < rect.Max.X; x++ {
		if err := setCellClipped(cvs, image.Point{X: x, Y: rect.Min.Y}, br.h, opts); err != nil {
			return err
		}
		if err := setCellClipped(cvs, image.Point{X: x, Y: rect.Max.Y - 1}, br.h, opts); err != nil {
			return err
		}
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		if err := setCellClipped(cvs, image.Point{X: rect.Min.X, Y: y}, br.v, opts); err != nil {
			return err
		}
		if err := setCellClipped(cvs, image.Point{X: rect.Max.X - 1, Y: y}, br.v, opts); err != nil {
			return err
		}
	}
	for _, c := range []struct {
		p image.Point
		r rune
	}{
		{image.Point{X: rect.Min.X, Y: rect.Min.Y}, br.tl},
		{image.Point{X: rect.Max.X - 1, Y: rect.Min.Y}, br.tr},
		{image.Point{X: rect.Min.X, Y: rect.Max.Y - 1}, br.bl},
		{image.Point{X: rect.Max.X - 1, Y: rect.Max.Y - 1}, br.br},
	} {
		if err := setCellClipped(cvs, c.p, c.r, opts); err != nil {
			return err
		}
	}
	return nil
}

func borderChars(lineStyle linestyle.LineStyle) (borderRunes, bool) {
	switch lineStyle {
	case linestyle.Light:
		return borderRunes{'─', '│', '┌', '┐', '└', '┘'}, true
	case linestyle.Double:
		return borderRunes{'═', '║', '╔', '╗', '╚', '╝'}, true
	case linestyle.Round:
		return borderRunes{'─', '│', '╭', '╮', '╰', '╯'}, true
	default:
		return borderRunes{}, false
	}
}

func setCellClipped(cvs *canvas.Canvas, p image.Point, r rune, opts []cell.Option) error {
	if !p.In(cvs.Area()) {
		return nil
	}
	_, err := cvs.SetCell(p, r, opts...)
	return err
}

func writeText(cvs *canvas.Canvas, text string, start image.Point, maxX int, opts []cell.Option) error {
	if start.Y < cvs.Area().Min.Y || start.Y >= cvs.Area().Max.Y || maxX <= start.X {
		return nil
	}
	cur := start
	for _, r := range text {
		width := runewidth.RuneWidth(r)
		if width < 1 {
			continue
		}
		if cur.X+width > maxX {
			break
		}
		if cur.X >= cvs.Area().Min.X && cur.X+width <= cvs.Area().Max.X {
			if _, err := cvs.SetCell(cur, r, opts...); err != nil {
				return err
			}
		}
		cur.X += width
	}
	return nil
}

func drawProgress(cvs *canvas.Canvas, start image.Point, maxX int, progress float64, opts []cell.Option) error {
	width := maxX - start.X
	if width <= 0 {
		return nil
	}
	progress = clampFloat(progress, 0, 1)
	filled := int(math.Round(float64(width) * progress))
	for i := 0; i < width; i++ {
		r := '─'
		if i < filled {
			r = '━'
		}
		if err := setCellClipped(cvs, image.Point{X: start.X + i, Y: start.Y}, r, opts); err != nil {
			return err
		}
	}
	return nil
}

func wrapLines(text string, width, maxLines int) []string {
	text = strings.TrimSpace(text)
	if text == "" || width <= 0 || maxLines <= 0 {
		return nil
	}

	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			continue
		}
		line := words[0]
		for _, word := range words[1:] {
			candidate := line + " " + word
			if runewidth.StringWidth(candidate) <= width {
				line = candidate
				continue
			}
			lines = appendWrappedLine(lines, line, width)
			line = word
			if len(lines) >= maxLines {
				return finishWrappedLines(lines[:maxLines], width)
			}
		}
		lines = appendWrappedLine(lines, line, width)
		if len(lines) >= maxLines {
			return finishWrappedLines(lines[:maxLines], width)
		}
	}
	return lines
}

func appendWrappedLine(lines []string, line string, width int) []string {
	for runewidth.StringWidth(line) > width {
		part, rest := splitAtWidth(line, width)
		lines = append(lines, part)
		line = strings.TrimLeft(rest, " ")
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func finishWrappedLines(lines []string, width int) []string {
	if len(lines) == 0 {
		return nil
	}
	if trimmed, err := draw.TrimText(lines[len(lines)-1], width, draw.OverrunModeThreeDot); err == nil {
		lines[len(lines)-1] = trimmed
	}
	return lines
}

func splitAtWidth(text string, width int) (string, string) {
	var b strings.Builder
	cur := 0
	for i, r := range text {
		rw := runewidth.RuneWidth(r)
		if cur+rw > width {
			return b.String(), text[i:]
		}
		b.WriteRune(r)
		cur += rw
	}
	return b.String(), ""
}

func drawActions(cvs *canvas.Canvas, id string, start image.Point, maxX int, actions []Action, opts []cell.Option) ([]actionHit, error) {
	if start.Y < cvs.Area().Min.Y || start.Y >= cvs.Area().Max.Y || maxX <= start.X {
		return nil, nil
	}
	var hits []actionHit
	cur := start
	for _, action := range actions {
		label := strings.TrimSpace(action.Label)
		if label == "" {
			continue
		}
		token := "[" + label + "]"
		width := runewidth.StringWidth(token)
		if cur.X > start.X {
			if cur.X+1 >= maxX {
				break
			}
			if err := writeText(cvs, " ", cur, maxX, opts); err != nil {
				return nil, err
			}
			cur.X++
		}
		if cur.X+width > maxX {
			break
		}
		rect := image.Rect(cur.X, cur.Y, cur.X+width, cur.Y+1)
		if err := writeText(cvs, token, cur, maxX, opts); err != nil {
			return nil, err
		}
		if action.Callback != nil {
			hits = append(hits, actionHit{
				id:       id,
				rect:     rect,
				callback: action.Callback,
				dismiss:  action.Dismiss,
			})
		}
		cur.X += width
	}
	return hits, nil
}

func actionText(actions []Action) string {
	var b strings.Builder
	for i, action := range actions {
		label := strings.TrimSpace(action.Label)
		if label == "" {
			continue
		}
		if i > 0 && b.Len() > 0 {
			b.WriteRune(' ')
		}
		b.WriteRune('[')
		b.WriteString(label)
		b.WriteRune(']')
	}
	return b.String()
}

func cloneActions(actions []Action) []Action {
	cloned := make([]Action, len(actions))
	copy(cloned, actions)
	return cloned
}

func mergeCellOpts(groups ...[]cell.Option) []cell.Option {
	var merged []cell.Option
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return merged
}

func fadeOpts(enabled bool) []cell.Option {
	if !enabled {
		return nil
	}
	return []cell.Option{cell.Dim()}
}

func easeOutCubic(v float64) float64 {
	v = clampFloat(v, 0, 1)
	return 1 - math.Pow(1-v, 3)
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
