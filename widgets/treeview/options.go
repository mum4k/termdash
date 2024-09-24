package treeview

import "github.com/mum4k/termdash/cell"

// Option represents a configuration option for the TreeView.
type Option func(*options)

// options holds the configuration for the TreeView.
type options struct {
	// nodes are the root nodes of the TreeView.
	nodes []*TreeNode
	// labelColor is the color of the node labels.
	labelColor cell.Color
	// expandedIcon is the icon used for expanded nodes.
	expandedIcon string
	// collapsedIcon is the icon used for collapsed nodes.
	collapsedIcon string
	// leafIcon is the icon used for leaf nodes.
	leafIcon string
	// indentation is the number of spaces per indentation level.
	indentation int
	// waitingIcons are the icons used for the spinner.
	waitingIcons []string
	// truncate indicates whether to truncate long labels.
	truncate bool
	// enableLogging enables or disables logging for debugging.
	enableLogging bool
}

// newOptions initializes default options.
func newOptions() *options {
	return &options{
		nodes:         []*TreeNode{},
		labelColor:    cell.ColorWhite,
		expandedIcon:  "▼",
		collapsedIcon: "▶",
		leafIcon:      "→",
		waitingIcons:  []string{"◐", "◓", "◑", "◒"},
		truncate:      false,
		indentation:   2, // Default indentation
	}
}

// Nodes sets the root nodes of the TreeView.
func Nodes(nodes ...*TreeNode) Option {
	return func(o *options) {
		o.nodes = nodes
	}
}

// Indentation sets the number of spaces for each indentation level.
func Indentation(spaces int) Option {
	return func(o *options) {
		o.indentation = spaces
	}
}

// Icons sets custom icons for expanded, collapsed, and leaf nodes.
func Icons(expanded, collapsed, leaf string) Option {
	return func(o *options) {
		o.expandedIcon = expanded
		o.collapsedIcon = collapsed
		o.leafIcon = leaf
	}
}

// LabelColor sets the color of the node labels.
func LabelColor(color cell.Color) Option {
	return func(o *options) {
		o.labelColor = color
	}
}

// WaitingIcons sets the icons for the spinner.
func WaitingIcons(icons []string) Option {
	return func(o *options) {
		o.waitingIcons = icons
	}
}

// Truncate enables or disables label truncation.
func Truncate(truncate bool) Option {
	return func(o *options) {
		o.truncate = truncate
	}
}

// EnableLogging enables or disables logging for debugging.
func EnableLogging(enable bool) Option {
	return func(o *options) {
		o.enableLogging = enable
	}
}

// Label sets the widget's label.
// Note: If the widget's label is managed by the container, this can be a no-op.
func Label(label string) Option {
	return func(o *options) {
		// No action needed; label is set in container's BorderTitle.
	}
}

// CollapsedIcon sets the icon for collapsed nodes.
func CollapsedIcon(icon string) Option {
	return func(o *options) {
		o.collapsedIcon = icon
	}
}

// ExpandedIcon sets the icon for expanded nodes.
func ExpandedIcon(icon string) Option {
	return func(o *options) {
		o.expandedIcon = icon
	}
}

// LeafIcon sets the icon for leaf nodes.
func LeafIcon(icon string) Option {
	return func(o *options) {
		o.leafIcon = icon
	}
}

// IndentationPerLevel sets the indentation per level.
// Alias to Indentation for compatibility with demo code.
func IndentationPerLevel(spaces int) Option {
	return Indentation(spaces)
}
