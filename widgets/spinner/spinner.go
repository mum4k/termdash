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

// Package spinner exposes reusable UTF-8 spinner frame sets sourced from
// github.com/kojix2/spinner2.
//
// The package intentionally stays low-level: callers work directly with named
// frame sets and build any title, border, or overlay behavior outward from
// those primitives.
package spinner

import (
	"sort"

	"github.com/mum4k/termdash/private/runewidth"
)

// Spinner is a reusable animation sequence.
type Spinner struct {
	name     string
	interval int
	frames   []string
}

type format struct {
	interval int
	frames   []string
}

var catalog = map[string]format{
	"classic":       {interval: 10, frames: []string{"|", "/", "-", "\\"}},
	"spin":          {interval: 10, frames: []string{"◴", "◷", "◶", "◵"}},
	"spin_2":        {interval: 10, frames: []string{"◐", "◓", "◑", "◒"}},
	"spin_3":        {interval: 10, frames: []string{"◰", "◳", "◲", "◱"}},
	"spin_4":        {interval: 10, frames: []string{"╫", "╪"}},
	"pulse":         {interval: 10, frames: []string{"⎺", "⎻", "⎼", "⎽", "⎼", "⎻"}},
	"pulse_2":       {interval: 15, frames: []string{"▁", "▃", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▃"}},
	"pulse_3":       {interval: 20, frames: []string{"▉", "▊", "▋", "▌", "▍", "▎", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}},
	"dots":          {interval: 10, frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}},
	"dots_2":        {interval: 10, frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}},
	"dots_3":        {interval: 10, frames: []string{"⠋", "⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓"}},
	"dots_4":        {interval: 10, frames: []string{"⠄", "⠆", "⠇", "⠋", "⠙", "⠸", "⠰", "⠠", "⠰", "⠸", "⠙", "⠋", "⠇", "⠆"}},
	"dots_5":        {interval: 10, frames: []string{"⠋", "⠙", "⠚", "⠒", "⠂", "⠂", "⠒", "⠲", "⠴", "⠦", "⠖", "⠒", "⠐", "⠐", "⠒", "⠓", "⠋"}},
	"dots_6":        {interval: 10, frames: []string{"⠁", "⠉", "⠙", "⠚", "⠒", "⠂", "⠂", "⠒", "⠲", "⠴", "⠤", "⠄", "⠄", "⠤", "⠴", "⠲", "⠒", "⠂", "⠂", "⠒", "⠚", "⠙", "⠉", "⠁"}},
	"dots_7":        {interval: 10, frames: []string{"⠈", "⠉", "⠋", "⠓", "⠒", "⠐", "⠐", "⠒", "⠖", "⠦", "⠤", "⠠", "⠠", "⠤", "⠦", "⠖", "⠒", "⠐", "⠐", "⠒", "⠓", "⠋", "⠉", "⠈"}},
	"dots_8":        {interval: 10, frames: []string{"⠁", "⠁", "⠉", "⠙", "⠚", "⠒", "⠂", "⠂", "⠒", "⠲", "⠴", "⠤", "⠄", "⠄", "⠤", "⠠", "⠠", "⠤", "⠦", "⠖", "⠒", "⠐", "⠐", "⠒", "⠓", "⠋", "⠉", "⠈", "⠈"}},
	"dots_9":        {interval: 10, frames: []string{"⢹", "⢺", "⢼", "⣸", "⣇", "⡧", "⡗", "⡏"}},
	"dots_10":       {interval: 10, frames: []string{"⢄", "⢂", "⢁", "⡁", "⡈", "⡐", "⡠"}},
	"dots_11":       {interval: 10, frames: []string{"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"}},
	"arrow":         {interval: 10, frames: []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}},
	"arrow_pulse":   {interval: 10, frames: []string{"▹▹▹▹▹", "▸▹▹▹▹", "▹▸▹▹▹", "▹▹▸▹▹", "▹▹▹▸▹", "▹▹▹▹▸"}},
	"triangle":      {interval: 10, frames: []string{"◢", "◣", "◤", "◥"}},
	"arc":           {interval: 10, frames: []string{"◜", "◠", "◝", "◞", "◡", "◟"}},
	"pipe":          {interval: 10, frames: []string{"┤", "┘", "┴", "└", "├", "┌", "┬", "┐"}},
	"bouncing":      {interval: 10, frames: []string{"[    ]", "[   =]", "[  ==]", "[ ===]", "[====]", "[=== ]", "[==  ]", "[=   ]"}},
	"bouncing_ball": {interval: 10, frames: []string{"( ●    )", "(  ●   )", "(   ●  )", "(    ● )", "(     ●)", "(    ● )", "(   ●  )", "(  ●   )", "( ●    )", "(●     )"}},
	"bounce":        {interval: 10, frames: []string{"⠁", "⠂", "⠄", "⠂"}},
	"box_bounce":    {interval: 10, frames: []string{"▌", "▀", "▐", "▄"}},
	"box_bounce_2":  {interval: 10, frames: []string{"▖", "▘", "▝", "▗"}},
	"star":          {interval: 10, frames: []string{"✶", "✸", "✹", "✺", "✹", "✷"}},
	"toggle":        {interval: 10, frames: []string{"■", "□", "▪", "▫"}},
	"balloon":       {interval: 10, frames: []string{".", "o", "O", "@", "*"}},
	"balloon_2":     {interval: 10, frames: []string{".", "o", "O", "°", "O", "o", "."}},
	"flip":          {interval: 10, frames: []string{"-", "◡", "⊙", "-", "◠"}},
	"burger":        {interval: 6, frames: []string{"☱", "☲", "☴"}},
	"dance":         {interval: 10, frames: []string{">))'>", " >))'>", "  >))'>", "   >))'>", "    >))'>", "   <'((<", "  <'((<", " <'((<"}},
	"shark":         {interval: 10, frames: []string{"▐|\\____________▌", "▐_|\\___________▌", "▐__|\\__________▌", "▐___|\\_________▌", "▐____|\\________▌", "▐_____|\\_______▌", "▐______|\\______▌", "▐_______|\\_____▌", "▐________|\\____▌", "▐_________|\\___▌", "▐__________|\\__▌", "▐___________|\\_▌", "▐____________|\\▌", "▐____________/|▌", "▐___________/|_▌", "▐__________/|__▌", "▐_________/|___▌", "▐________/|____▌", "▐_______/|_____▌", "▐______/|______▌", "▐_____/|_______▌", "▐____/|________▌", "▐___/|_________▌", "▐__/|__________▌", "▐_/|___________▌", "▐/|____________▌"}},
	"pong":          {interval: 10, frames: []string{"▐⠂       ▌", "▐⠈       ▌", "▐ ⠂      ▌", "▐ ⠠      ▌", "▐  ⡀     ▌", "▐  ⠠     ▌", "▐   ⠂    ▌", "▐   ⠈    ▌", "▐    ⠂   ▌", "▐    ⠠   ▌", "▐     ⡀  ▌", "▐     ⠠  ▌", "▐      ⠂ ▌", "▐      ⠈ ▌", "▐       ⠂▌", "▐       ⠠▌", "▐       ⡀▌", "▐      ⠠ ▌", "▐      ⠂ ▌", "▐     ⠈  ▌", "▐     ⠂  ▌", "▐    ⠠   ▌", "▐    ⡀   ▌", "▐   ⠠    ▌", "▐   ⠂    ▌", "▐  ⠈     ▌", "▐  ⠂     ▌", "▐ ⠠      ▌", "▐ ⡀      ▌", "▐⠠       ▌"}},
}

// Names returns the stable spinner names in sorted order.
func Names() []string {
	names := make([]string, 0, len(catalog))
	for name := range catalog {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// All returns every spinner in sorted name order.
func All() []Spinner {
	names := Names()
	all := make([]Spinner, 0, len(names))
	for _, name := range names {
		all = append(all, Must(name))
	}
	return all
}

// Lookup returns a spinner by name.
func Lookup(name string) (Spinner, bool) {
	f, ok := catalog[name]
	if !ok {
		return Spinner{}, false
	}
	frames := make([]string, len(f.frames))
	copy(frames, f.frames)
	return Spinner{name: name, interval: f.interval, frames: frames}, true
}

// Must returns a spinner by name and panics if it doesn't exist.
func Must(name string) Spinner {
	s, ok := Lookup(name)
	if !ok {
		panic("spinner: unknown format " + name)
	}
	return s
}

// Name returns the stable spinner name.
func (s Spinner) Name() string {
	return s.name
}

// Interval returns the upstream spinner interval.
func (s Spinner) Interval() int {
	return s.interval
}

// Frames returns a copy of the spinner's frame set.
func (s Spinner) Frames() []string {
	cp := make([]string, len(s.frames))
	copy(cp, s.frames)
	return cp
}

// Frame returns the string frame for the requested step.
func (s Spinner) Frame(step int) string {
	if len(s.frames) == 0 {
		return ""
	}
	return s.frames[positiveMod(step, len(s.frames))]
}

// DecorateRight appends the current frame to a label.
func (s Spinner) DecorateRight(label string, step int) string {
	frame := s.Frame(step)
	if frame == "" {
		return label
	}
	return label + frame
}

// DecorateLeft prefixes the current frame to a label.
func (s Spinner) DecorateLeft(label string, step int) string {
	frame := s.Frame(step)
	if frame == "" {
		return label
	}
	return frame + label
}

// SingleCell reports whether every frame occupies exactly one cell.
func (s Spinner) SingleCell() bool {
	_, ok := s.RuneFrames()
	return ok
}

// RuneFrames converts the spinner into single-cell runes when possible.
func (s Spinner) RuneFrames() ([]rune, bool) {
	if len(s.frames) == 0 {
		return nil, true
	}
	runes := make([]rune, len(s.frames))
	for i, frame := range s.frames {
		rs := []rune(frame)
		if len(rs) != 1 || runewidth.RuneWidth(rs[0]) != 1 {
			return nil, false
		}
		runes[i] = rs[0]
	}
	return runes, true
}

// Rune returns the rune for the requested step when the spinner is single-cell.
func (s Spinner) Rune(step int) (rune, bool) {
	frames, ok := s.RuneFrames()
	if !ok || len(frames) == 0 {
		return 0, false
	}
	return frames[positiveMod(step, len(frames))], true
}

func positiveMod(v, mod int) int {
	if mod <= 0 {
		return 0
	}
	v %= mod
	if v < 0 {
		v += mod
	}
	return v
}
