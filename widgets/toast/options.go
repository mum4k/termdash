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

package toast

import (
	"fmt"
	"image"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
)

// Placement identifies the anchor point used for toast stacks.
type Placement int

// Placement values supported by Manager.
const (
	// PlacementTopRight anchors notifications to the upper-right corner.
	PlacementTopRight Placement = iota
	// PlacementTopLeft anchors notifications to the upper-left corner.
	PlacementTopLeft
	// PlacementBottomRight anchors notifications to the lower-right corner.
	PlacementBottomRight
	// PlacementBottomLeft anchors notifications to the lower-left corner.
	PlacementBottomLeft
	// PlacementCenter anchors notifications near the center of the canvas.
	PlacementCenter
	// PlacementCustom delegates placement to a caller-provided PositionFunc.
	PlacementCustom
)

// Animation identifies the motion used when a toast appears.
type Animation int

// Animation values supported by Manager.
const (
	// AnimationNone renders notifications at their final position immediately.
	AnimationNone Animation = iota
	// AnimationSlide slides notifications in from the configured direction.
	AnimationSlide
	// AnimationFade renders notifications with dimmed cells until settled.
	AnimationFade
	// AnimationPop grows notifications from a compact centered rectangle.
	AnimationPop
)

// Direction identifies the edge used by slide animations.
type Direction int

// Direction values supported by slide animations.
const (
	// DirectionRight slides notifications in from the right edge.
	DirectionRight Direction = iota
	// DirectionLeft slides notifications in from the left edge.
	DirectionLeft
	// DirectionTop slides notifications in from the top edge.
	DirectionTop
	// DirectionBottom slides notifications in from the bottom edge.
	DirectionBottom
)

// StackDirection identifies how multiple notifications are stacked.
type StackDirection int

// StackDirection values supported by Manager.
const (
	// StackAuto chooses a natural stack direction for the configured Placement.
	StackAuto StackDirection = iota
	// StackDown stacks later notifications downward from the anchor.
	StackDown
	// StackUp stacks later notifications upward from the anchor.
	StackUp
)

// Severity identifies the semantic style of a notification.
type Severity int

// Severity values supported by Manager.
const (
	// SeverityInfo is the default informational notification style.
	SeverityInfo Severity = iota
	// SeveritySuccess is used for successful operations.
	SeveritySuccess
	// SeverityWarning is used for warnings and recoverable problems.
	SeverityWarning
	// SeverityError is used for failures and destructive outcomes.
	SeverityError
	// SeverityNeutral is used for low-emphasis notifications.
	SeverityNeutral
)

// PositionFunc returns the top-left corner for a toast.
//
// The canvas rectangle is the full drawing area, size is the toast's requested
// size, and index is the toast's index in the visible stack.
type PositionFunc func(canvas image.Rectangle, size image.Point, index int) image.Point

// Style defines visual details for one notification severity.
type Style struct {
	// Icon is drawn before the notification title when non-zero.
	Icon rune
	// Accent is the default accent color used when a more specific cell option
	// is not supplied.
	Accent cell.Color
	// BorderCellOpts styles the toast border.
	BorderCellOpts []cell.Option
	// FillCellOpts styles the toast interior fill cells.
	FillCellOpts []cell.Option
	// TitleCellOpts styles the notification title.
	TitleCellOpts []cell.Option
	// MessageCellOpts styles the notification message.
	MessageCellOpts []cell.Option
	// IconCellOpts styles the severity icon.
	IconCellOpts []cell.Option
	// ProgressCellOpts styles the optional progress bar.
	ProgressCellOpts []cell.Option
}

// Option configures a Manager.
type Option interface {
	// set applies the option to the provided options.
	set(*options)
}

type option func(*options)

func (o option) set(opts *options) {
	o(opts)
}

type options struct {
	placement         Placement
	customPosition    PositionFunc
	width             int
	minWidth          int
	maxWidth          int
	margin            image.Point
	gap               int
	maxVisible        int
	maxMessageLines   int
	defaultTTL        time.Duration
	animation         Animation
	slideDirection    Direction
	animationDuration time.Duration
	stackDirection    StackDirection
	newestFirst       bool
	border            linestyle.LineStyle
	borderCellOpts    []cell.Option
	fillCellOpts      []cell.Option
	titleCellOpts     []cell.Option
	messageCellOpts   []cell.Option
	actionCellOpts    []cell.Option
	shadow            bool
	shadowCellOpts    []cell.Option
	dismissOnClick    bool
	minimumSize       image.Point
	clock             func() time.Time
	styles            map[Severity]Style
}

func newOptions(opts ...Option) *options {
	o := &options{
		placement:         PlacementTopRight,
		width:             42,
		minWidth:          8,
		maxWidth:          72,
		margin:            image.Point{X: 2, Y: 1},
		gap:               1,
		maxVisible:        4,
		maxMessageLines:   4,
		defaultTTL:        5 * time.Second,
		animation:         AnimationSlide,
		slideDirection:    DirectionRight,
		animationDuration: 350 * time.Millisecond,
		stackDirection:    StackAuto,
		newestFirst:       true,
		border:            linestyle.Round,
		borderCellOpts: []cell.Option{
			cell.FgColor(cell.ColorNumber(244)),
		},
		fillCellOpts: []cell.Option{
			cell.BgColor(cell.ColorNumber(16)),
		},
		titleCellOpts: []cell.Option{
			cell.FgColor(cell.ColorNumber(231)),
			cell.Bold(),
		},
		messageCellOpts: []cell.Option{
			cell.FgColor(cell.ColorNumber(252)),
		},
		actionCellOpts: []cell.Option{
			cell.FgColor(cell.ColorNumber(159)),
			cell.Bold(),
		},
		shadow: true,
		shadowCellOpts: []cell.Option{
			cell.BgColor(cell.ColorNumber(235)),
		},
		dismissOnClick: true,
		minimumSize:    image.Point{X: 8, Y: 3},
		clock:          time.Now,
		styles:         defaultStyles(),
	}
	for _, opt := range opts {
		opt.set(o)
	}
	return o
}

func defaultStyles() map[Severity]Style {
	return map[Severity]Style{
		SeverityInfo: {
			Icon:   '●',
			Accent: cell.ColorNumber(81),
		},
		SeveritySuccess: {
			Icon:   '◆',
			Accent: cell.ColorNumber(120),
		},
		SeverityWarning: {
			Icon:   '▲',
			Accent: cell.ColorNumber(228),
		},
		SeverityError: {
			Icon:   '■',
			Accent: cell.ColorNumber(203),
		},
		SeverityNeutral: {
			Icon:   '•',
			Accent: cell.ColorNumber(250),
		},
	}
}

func (o *options) validate() error {
	if o.width < 1 {
		return fmt.Errorf("toast width must be at least one, got %d", o.width)
	}
	if o.minWidth < 1 {
		return fmt.Errorf("toast minimum width must be at least one, got %d", o.minWidth)
	}
	if o.maxWidth < o.minWidth {
		return fmt.Errorf("toast maximum width %d is smaller than minimum width %d", o.maxWidth, o.minWidth)
	}
	if o.margin.X < 0 || o.margin.Y < 0 {
		return fmt.Errorf("toast margins cannot be negative, got %v", o.margin)
	}
	if o.gap < 0 {
		return fmt.Errorf("toast gap cannot be negative, got %d", o.gap)
	}
	if o.maxVisible < 1 {
		return fmt.Errorf("toast max visible must be at least one, got %d", o.maxVisible)
	}
	if o.maxMessageLines < 1 {
		return fmt.Errorf("toast max message lines must be at least one, got %d", o.maxMessageLines)
	}
	if o.defaultTTL < 0 {
		return fmt.Errorf("toast default TTL cannot be negative, got %v", o.defaultTTL)
	}
	if o.animationDuration < 0 {
		return fmt.Errorf("toast animation duration cannot be negative, got %v", o.animationDuration)
	}
	if o.minimumSize.X < 1 || o.minimumSize.Y < 1 {
		return fmt.Errorf("toast minimum size must be positive, got %v", o.minimumSize)
	}
	switch o.placement {
	case PlacementTopRight, PlacementTopLeft, PlacementBottomRight, PlacementBottomLeft, PlacementCenter, PlacementCustom:
	default:
		return fmt.Errorf("unsupported toast placement %d", o.placement)
	}
	if o.placement == PlacementCustom && o.customPosition == nil {
		return fmt.Errorf("custom toast placement requires CustomPosition")
	}
	switch o.animation {
	case AnimationNone, AnimationSlide, AnimationFade, AnimationPop:
	default:
		return fmt.Errorf("unsupported toast animation %d", o.animation)
	}
	switch o.slideDirection {
	case DirectionRight, DirectionLeft, DirectionTop, DirectionBottom:
	default:
		return fmt.Errorf("unsupported toast slide direction %d", o.slideDirection)
	}
	switch o.stackDirection {
	case StackAuto, StackDown, StackUp:
	default:
		return fmt.Errorf("unsupported toast stack direction %d", o.stackDirection)
	}
	switch o.border {
	case linestyle.None, linestyle.Light, linestyle.Double, linestyle.Round:
	default:
		return fmt.Errorf("unsupported toast border style %v", o.border)
	}
	if o.clock == nil {
		return fmt.Errorf("toast clock cannot be nil")
	}
	return nil
}

// Anchor sets the toast stack placement.
func Anchor(p Placement) Option {
	return option(func(o *options) {
		o.placement = p
	})
}

// CustomPosition sets a custom placement function and selects PlacementCustom.
func CustomPosition(fn PositionFunc) Option {
	return option(func(o *options) {
		o.customPosition = fn
		o.placement = PlacementCustom
	})
}

// Width sets the preferred toast width in terminal cells.
func Width(width int) Option {
	return option(func(o *options) {
		o.width = width
	})
}

// MinWidth sets the smallest toast width used during responsive layout.
func MinWidth(width int) Option {
	return option(func(o *options) {
		o.minWidth = width
	})
}

// MaxWidth sets the largest toast width used during responsive layout.
func MaxWidth(width int) Option {
	return option(func(o *options) {
		o.maxWidth = width
	})
}

// Margin sets the horizontal and vertical distance from the canvas edge.
func Margin(x, y int) Option {
	return option(func(o *options) {
		o.margin = image.Point{X: x, Y: y}
	})
}

// Gap sets the number of cells between stacked notifications.
func Gap(gap int) Option {
	return option(func(o *options) {
		o.gap = gap
	})
}

// MaxVisible sets the maximum number of notifications drawn at once.
func MaxVisible(count int) Option {
	return option(func(o *options) {
		o.maxVisible = count
	})
}

// MaxMessageLines sets the maximum message lines shown in each toast.
func MaxMessageLines(count int) Option {
	return option(func(o *options) {
		o.maxMessageLines = count
	})
}

// DefaultTTL sets the lifetime used by notifications that do not set a TTL.
func DefaultTTL(ttl time.Duration) Option {
	return option(func(o *options) {
		o.defaultTTL = ttl
	})
}

// AnimationMode sets the notification entrance animation.
func AnimationMode(a Animation) Option {
	return option(func(o *options) {
		o.animation = a
	})
}

// SlideDirection sets the edge used by AnimationSlide.
func SlideDirection(d Direction) Option {
	return option(func(o *options) {
		o.slideDirection = d
	})
}

// AnimationDuration sets how long entrance animations run.
func AnimationDuration(d time.Duration) Option {
	return option(func(o *options) {
		o.animationDuration = d
	})
}

// Stack sets how multiple visible notifications are stacked.
func Stack(direction StackDirection) Option {
	return option(func(o *options) {
		o.stackDirection = direction
	})
}

// NewestFirst controls whether the newest notifications are closest to the
// configured anchor.
func NewestFirst(enabled bool) Option {
	return option(func(o *options) {
		o.newestFirst = enabled
	})
}

// Border sets the toast border style and optional base cell options.
func Border(ls linestyle.LineStyle, opts ...cell.Option) Option {
	return option(func(o *options) {
		o.border = ls
		o.borderCellOpts = append([]cell.Option(nil), opts...)
	})
}

// Borderless disables the toast border.
func Borderless() Option {
	return option(func(o *options) {
		o.border = linestyle.None
	})
}

// FillCellOpts sets the base styling for toast interior cells.
func FillCellOpts(opts ...cell.Option) Option {
	return option(func(o *options) {
		o.fillCellOpts = append([]cell.Option(nil), opts...)
	})
}

// TitleCellOpts sets the base styling for toast titles.
func TitleCellOpts(opts ...cell.Option) Option {
	return option(func(o *options) {
		o.titleCellOpts = append([]cell.Option(nil), opts...)
	})
}

// MessageCellOpts sets the base styling for toast message text.
func MessageCellOpts(opts ...cell.Option) Option {
	return option(func(o *options) {
		o.messageCellOpts = append([]cell.Option(nil), opts...)
	})
}

// ActionCellOpts sets the base styling for action labels.
func ActionCellOpts(opts ...cell.Option) Option {
	return option(func(o *options) {
		o.actionCellOpts = append([]cell.Option(nil), opts...)
	})
}

// Shadow enables or disables the one-cell toast drop shadow.
func Shadow(enabled bool, opts ...cell.Option) Option {
	return option(func(o *options) {
		o.shadow = enabled
		if len(opts) > 0 {
			o.shadowCellOpts = append([]cell.Option(nil), opts...)
		}
	})
}

// DismissOnClick controls whether clicking a visible toast removes it.
func DismissOnClick(enabled bool) Option {
	return option(func(o *options) {
		o.dismissOnClick = enabled
	})
}

// MinimumSize sets the smallest canvas requested by the widget.
func MinimumSize(size image.Point) Option {
	return option(func(o *options) {
		o.minimumSize = size
	})
}

// Clock sets the time source used for TTL and animation calculations.
func Clock(clock func() time.Time) Option {
	return option(func(o *options) {
		o.clock = clock
	})
}

// SeverityStyle replaces the visual style for a severity.
func SeverityStyle(sev Severity, style Style) Option {
	return option(func(o *options) {
		o.styles[sev] = cloneStyle(style)
	})
}

func cloneStyle(style Style) Style {
	return Style{
		Icon:             style.Icon,
		Accent:           style.Accent,
		BorderCellOpts:   append([]cell.Option(nil), style.BorderCellOpts...),
		FillCellOpts:     append([]cell.Option(nil), style.FillCellOpts...),
		TitleCellOpts:    append([]cell.Option(nil), style.TitleCellOpts...),
		MessageCellOpts:  append([]cell.Option(nil), style.MessageCellOpts...),
		IconCellOpts:     append([]cell.Option(nil), style.IconCellOpts...),
		ProgressCellOpts: append([]cell.Option(nil), style.ProgressCellOpts...),
	}
}
