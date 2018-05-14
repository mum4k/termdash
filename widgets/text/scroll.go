package text

// scroll.go contains code that tracks the current scrolling position.

import "math"

// scrollTracker tracks the current scrolling position for the Text widget.
// This is not thread safe.
type scrollTracker struct {
	// scroll stores user requests to scroll up (negative) or down (positive).
	// E.g. -1 means up by one line and 2 means down by two lines.
	scroll int

	// scrollPage stores user requests to scroll up (negative) or down
	// (positive) by a page of content. E.g. -1 means up by one page and 2
	// means down by two pages.
	scrollPage int

	// first tracks the first line that will be printed.
	first int

	// state is the state of the scrolling FSM.
	state rollState
}

// newScrollTracker returns a new scroll tracker.
func newScrollTracker(opts *options) *scrollTracker {
	if opts.rollContent {
		return &scrollTracker{state: rollToEnd}
	}
	return &scrollTracker{state: rollingDisabled}
}

// upOneLine processes a user request to scroll up by one line.
func (st *scrollTracker) upOneLine() {
	st.scroll--
}

// downOneLine processes a user request to scroll down by one line.
func (st *scrollTracker) downOneLine() {
	st.scroll++
}

// upOnePage processes a user request to scroll up by one page.
func (st *scrollTracker) upOnePage() {
	st.scrollPage--
}

// downOnePage processes a user request to scroll down by one page.
func (st *scrollTracker) downOnePage() {
	st.scrollPage++
}

// doScroll processes any outstanding scroll requests and calculates the
// resulting first line.
func (st *scrollTracker) doScroll(lines, height int) int {
	first := st.first + st.scroll + st.scrollPage*height
	st.scroll = 0
	st.scrollPage = 0
	return normalizeScroll(first, lines, height)
}

// firstLine returns the number of the first line that should be drawn on a
// canvas of the specified height if there is the provided number of lines of
// text.
func (st *scrollTracker) firstLine(lines, height int) int {
	// Execute the scrolling FSM.
	st.state = st.state(st, lines, height)
	return st.first
}

// rollState is a state in the scrolling FSM.
type rollState func(st *scrollTracker, lines, height int) rollState

// rollingDisabled is a state where content rolling was disabled by the
// configuration of the Text widget.
func rollingDisabled(st *scrollTracker, lines, height int) rollState {
	st.first = st.doScroll(lines, height)
	return rollingDisabled
}

// rollToEnd is a state in which the last line of the content is always
// visible. When new content arrives, it is rolled upwards.
func rollToEnd(st *scrollTracker, lines, height int) rollState {
	// If the user didn't scroll, just roll the content so that the last line
	// is visible.
	if st.scroll == 0 && st.scrollPage == 0 {
		st.first = normalizeScroll(math.MaxUint32, lines, height)
		return rollToEnd
	}

	st.first = st.doScroll(lines, height)
	if lastLineVisible(st.first, lines, height) {
		return rollToEnd
	}
	return rollingPaused
}

// rollingPaused is a state in which the user scrolled up and made the last
// line scroll out of the view, so the content rolling is paused.
func rollingPaused(st *scrollTracker, lines, height int) rollState {
	st.first = st.doScroll(lines, height)
	if lastLineVisible(st.first, lines, height) {
		return rollToEnd
	}
	return rollingPaused
}

// lastLineVisible returns true if the last line is visible given drawing that
// starts from the first line, the number of lines and the height of the
// canvas.
func lastLineVisible(first, lines, height int) bool {
	return lines-first <= height
}

// normalizeScroll returns normalized position of the first line that should be
// drawn when drawing the specified number of lines on a canvas with the
// provided height.
func normalizeScroll(first, lines, height int) int {
	if first < 0 || lines <= 0 || height <= 0 {
		return 0
	}

	if lines <= height {
		return 0 // Scrolling not necessary if the content fits.
	}

	max := lines - height
	if first > max {
		return max
	}
	return first
}
