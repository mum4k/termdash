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

package treeview

import (
	"io"

	"github.com/mum4k/termdash/cell"
)

// Option configures a TreeView.
type Option interface {
	set(*options)
}

// optionFunc adapts a plain function to the Option interface.
type optionFunc func(*options)

func (f optionFunc) set(o *options) { f(o) }

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
	// logWriter receives debug log output; nil means io.Discard.
	logWriter io.Writer
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
		indentation:   2,
	}
}

// Nodes sets the root nodes of the TreeView.
func Nodes(nodes ...*TreeNode) Option {
	return optionFunc(func(o *options) {
		o.nodes = nodes
	})
}

// Indentation sets the number of spaces for each indentation level.
func Indentation(spaces int) Option {
	return optionFunc(func(o *options) {
		o.indentation = spaces
	})
}

// Icons sets custom icons for expanded, collapsed, and leaf nodes.
func Icons(expanded, collapsed, leaf string) Option {
	return optionFunc(func(o *options) {
		o.expandedIcon = expanded
		o.collapsedIcon = collapsed
		o.leafIcon = leaf
	})
}

// LabelColor sets the color of the node labels.
func LabelColor(color cell.Color) Option {
	return optionFunc(func(o *options) {
		o.labelColor = color
	})
}

// WaitingIcons sets the icons for the spinner.
func WaitingIcons(icons []string) Option {
	return optionFunc(func(o *options) {
		o.waitingIcons = icons
	})
}

// Truncate enables or disables label truncation.
func Truncate(truncate bool) Option {
	return optionFunc(func(o *options) {
		o.truncate = truncate
	})
}

// LogWriter directs debug log output to w.  Pass io.Discard (or omit the
// option) to silence all output.  New() never opens files on your behalf.
func LogWriter(w io.Writer) Option {
	return optionFunc(func(o *options) {
		o.logWriter = w
	})
}

// EnableLogging is a no-op kept for backward compatibility.
// Use LogWriter(w) to capture debug output instead.
//
// Deprecated: use LogWriter.
func EnableLogging(_ bool) Option {
	return optionFunc(func(*options) {})
}

// Label sets the widget's label.
// Note: If the widget's label is managed by the container, this can be a no-op.
func Label(_ string) Option {
	return optionFunc(func(*options) {
		// No action needed; label is set in container's BorderTitle.
	})
}

// CollapsedIcon sets the icon for collapsed nodes.
func CollapsedIcon(icon string) Option {
	return optionFunc(func(o *options) {
		o.collapsedIcon = icon
	})
}

// ExpandedIcon sets the icon for expanded nodes.
func ExpandedIcon(icon string) Option {
	return optionFunc(func(o *options) {
		o.expandedIcon = icon
	})
}

// LeafIcon sets the icon for leaf nodes.
func LeafIcon(icon string) Option {
	return optionFunc(func(o *options) {
		o.leafIcon = icon
	})
}

// IndentationPerLevel sets the indentation per level.
// Alias to Indentation for compatibility with demo code.
func IndentationPerLevel(spaces int) Option {
	return Indentation(spaces)
}
