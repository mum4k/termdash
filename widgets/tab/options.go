// Package tab provides configuration options for the tabbed interface.
package tab

import "github.com/mum4k/termdash/cell"

// Option represents a configuration option for the Tab.
type Option interface {
	// set applies the option to the provided Options struct.
	set(*Options)
}

// Options holds the configuration for the Tab.
type Options struct {
	Tabs                []*Tab     // List of tabs.
	LabelColor          cell.Color // Color of the tab labels.
	ActiveTextColor     cell.Color // Text color of the active tab.
	InactiveTextColor   cell.Color // Text color of inactive tabs.
	ActiveTabColor      cell.Color // Background color of the active tab.
	InactiveTabColor    cell.Color // Background color of inactive tabs.
	ActiveIcon          string     // Icon for active tabs.
	InactiveIcon        string     // Icon for inactive tabs.
	NotificationIcon    string     // Icon for tabs with notifications.
	NotificationColor   cell.Color // Foreground color of the notification icon (alarm indicator).
	AnimatedActiveTab   bool       // Enables a moving accent marker in the active tab underline row.
	SweepTextColor      cell.Color // Text color used under the active sweep.
	SweepAccentColor    cell.Color // Accent color used for the moving underline marker.
	EnableLogging       bool       // Enables logging for debugging.
	FollowNotifications bool       // Whether to follow notifications automatically.
}

// NewOptions initializes default options or applies provided options.
func NewOptions(opts ...Option) *Options {
	o := &Options{
		Tabs:              []*Tab{},
		LabelColor:        cell.ColorNumber(252),
		ActiveTextColor:   cell.ColorNumber(231),
		InactiveTextColor: cell.ColorNumber(247),
		ActiveTabColor:    cell.ColorNumber(24),
		InactiveTabColor:  cell.ColorNumber(236),
		ActiveIcon:        "◆",
		InactiveIcon:      "○",
		NotificationIcon:  "⚠",
		// Bright amber — high-contrast alarm colour visible on both light and
		// dark tab backgrounds without clashing with the active/inactive text.
		NotificationColor:   cell.ColorNumber(214),
		AnimatedActiveTab:   false,
		SweepTextColor:      cell.ColorNumber(245),
		SweepAccentColor:    cell.ColorNumber(87),
		EnableLogging:       false,
		FollowNotifications: false,
	}
	for _, opt := range opts {
		opt.set(o)
	}
	return o
}

// option is a function that modifies Options.
type option func(*Options)

// set implements Option.set.
func (o option) set(opts *Options) {
	o(opts)
}

// Tabs sets the root tabs of the Tab widget.
func Tabs(tabs ...*Tab) Option {
	return option(func(o *Options) {
		o.Tabs = tabs
	})
}

// ActiveIcon sets custom icons for the active state.
func ActiveIcon(active string) Option {
	return option(func(o *Options) {
		o.ActiveIcon = active
	})
}

// InactiveIcon sets custom icons for the inactive state.
func InactiveIcon(inactive string) Option {
	return option(func(o *Options) {
		o.InactiveIcon = inactive
	})
}

// NotificationIcon sets custom icons for notifications.
func NotificationIcon(notification string) Option {
	return option(func(o *Options) {
		o.NotificationIcon = notification
	})
}

// NotificationColor sets the foreground color of the notification/alarm icon
// shown on tabs that have an active notification. Defaults to bright amber
// (xterm colour 214) so the icon is impossible to miss.
func NotificationColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.NotificationColor = color
	})
}

// LabelColor sets the color of the tab labels.
func LabelColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.LabelColor = color
	})
}

// ActiveTabColor sets the background color of the active tab.
func ActiveTabColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.ActiveTabColor = color
	})
}

// ActiveTextColor sets the text color of the active tab.
func ActiveTextColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.ActiveTextColor = color
	})
}

// InactiveTextColor sets the text color of inactive tabs.
func InactiveTextColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.InactiveTextColor = color
	})
}

// InactiveTabColor sets the background color of inactive tabs.
func InactiveTabColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.InactiveTabColor = color
	})
}

// AnimatedActiveTab enables or disables the active-tab underline animation.
//
// When enabled, the active tab keeps its underline while a small accent marker
// moves across it. Callers that prefer a calmer header can leave this disabled
// and still get a static underline on the active tab.
func AnimatedActiveTab(enable bool) Option {
	return option(func(o *Options) {
		o.AnimatedActiveTab = enable
	})
}

// SweepTextColor sets the text color used while the active sweep passes.
func SweepTextColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.SweepTextColor = color
	})
}

// SweepAccentColor sets the accent color used by the moving underline marker.
func SweepAccentColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.SweepAccentColor = color
	})
}

// EnableLogging enables or disables logging for debugging.
func EnableLogging(enable bool) Option {
	return option(func(o *Options) {
		o.EnableLogging = enable
	})
}

// FollowNotifications sets whether the app should follow notifications automatically.
func FollowNotifications(enable bool) Option {
	return option(func(o *Options) {
		o.FollowNotifications = enable
	})
}
