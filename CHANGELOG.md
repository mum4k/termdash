# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Added `time.Duration` based `ValueFormatter` for the `LineChart` Y-axis labels.
- Added round and suffix `ValueFormatter` for the `LineChart` Y-axis labels.
- Added decimal and suffix `ValueFormatter` for the `LineChart` Y-axis labels.
- Added a `container.SplitOption` that allows fixed size container splits.
- Added `grid` functions that allow fixed size rows and columns.

### Changed

- The `LineChart` can format the labels on the Y-axis with a `ValueFormatter`.
- The `SegmentDisplay` can now display dots and colons ('.' and ':').
- The `Donut` widget now guarantees spacing between the donut and its label.
- The continuous build on Travis CI now builds with cgo explicitly disabled to
  ensure both Termdash and its dependencies use pure Go. 

### Fixed

- Lint issues found on the Go report card.
- An internal library belonging to the `Text` widget was incorrectly passing
  `math.MaxUint32` as an int argument.

## [0.9.1] - 15-May-2019

### Fixed

- Termdash could deadlock when a `Button` or a `TextInput` was configured to
  call the `Container.Update` method.

## [0.9.0] - 28-Apr-2019

### Added

- The `TextInput` widget, an input field allowing interactive text input.
- The `Donut` widget can now display an optional text label under the donut.

### Changed

- Widgets now get information whether their container is focused when Draw is
  executed.
- The SegmentDisplay widget now has a method that returns the observed character
  capacity the last time Draw was called.
- The grid.Builder API now allows users to specify options for intermediate
  containers, i.e. containers that don't have widgets, but represent rows and
  columns.
- Line chart widget now allows `math.NaN` values to represent "no value" (values
  that will not be rendered) in the values slice.

#### Breaking API changes

- The widgetapi.Widget.Draw method now accepts a second argument which provides
  widgets with additional metadata. This affects all implemented widgets.
- Termdash now requires at least Go version 1.10, which allows us to utilize
  `math.Round` instead of our own implementation and `strings.Builder` instead
  of `bytes.Buffer`.
- Terminal shortcuts like `Ctrl-A` no longer come as two separate events,
  Termdash now mirrors termbox-go and sends these as one event.

## [0.8.0] - 30-Mar-2019

### Added

- New API for building layouts, a grid.Builder. Allows defining the layout
  iteratively as repetitive Elements, Rows and Columns.
- Containers now support margin around them and padding of their content.
- Container now supports dynamic layout changes via the new Update method.

### Changed

- The Text widget now supports content wrapping on word boundaries.
- The BarChart and SparkLine widgets now have a method that returns the
  observed value capacity the last time Draw was called.
- Moving widgetapi out of the internal directory to allow external users to
  develop their own widgets.
- Event delivery to widgets now has a stable defined order and happens when the
  container is unlocked so that widgets can trigger dynamic layout changes.

### Fixed

- The termdash_test now correctly waits until all subscribers processed events,
  not just received them.
- Container focus tracker now correctly tracks focus changes in enlarged areas,
  i.e. when the terminal size increased.
- The BarChart, LineChart and SegmentDisplay widgets now protect against
  external mutation of the values passed into them by copying the data they
  receive.

## [0.7.2] - 25-Feb-2019

### Added

- Test coverage for data only packages.

### Changed

- Refactoring packages that contained a mix of public and internal identifiers.

#### Breaking API changes

The following packages were refactored, no impact is expected as the removed
identifiers shouldn't be used externally.

- Functions align.Text and align.Rectangle were moved to a new
  internal/alignfor package.
- Types cell.Cell and cell.Buffer were moved into a new internal/canvas/buffer
  package.

## [0.7.1] - 24-Feb-2019

### Fixed

- Some of the packages that were moved into internal are required externally.
  This release makes them available again.

### Changed

#### Breaking API changes

- The draw.LineStyle enum was refactored into its own package
  linestyle.LineStyle. Users will have to replace:

  - draw.LineStyleNone -> linestyle.None
  - draw.LineStyleLight -> linestyle.Light
  - draw.LineStyleDouble -> linestyle.Double
  - draw.LineStyleRound -> linestyle.Round

## [0.7.0] - 24-Feb-2019

### Added

#### New widgets

- The Button widget.

#### Improvements to documentation

- Clearly marked the public API surface by moving private packages into
  internal directory.
- Started a GitHub wiki for Termdash.

#### Improvements to the LineChart widget

- The LineChart widget can display X axis labels in vertical orientation.
- The LineChart widget allows the user to specify a custom scale for the Y
  axis.
- The LineChart widget now has an option that disables scaling of the X axis.
  Useful for applications that want to continuously feed data and make them
  "roll" through the linechart.
- The LineChart widget now has a method that returns the observed capacity of
  the LineChart the last time Draw was called.
- The LineChart widget now supports zoom of the content triggered by mouse
  events.

#### Improvements to the Text widget

- The Text widget now has a Write option that atomically replaces the entire
  text content.

#### Improvements to the infrastructure

- A function that draws text vertically.
- A non-blocking event distribution system that can throttle repetitive events.
- Generalized mouse button FSM for use in widgets that need to track mouse
  button clicks.

### Changed

- Termbox is now initialized in 256 color mode by default.
- The infrastructure now uses the non-blocking event distribution system to
  distribute events to subscribers. Each widget is now an individual
  subscriber.
- The infrastructure now throttles event driven screen redraw rather than
  redrawing for each input event.
- Widgets can now specify the scope at which they want to receive keyboard and
  mouse events.

#### Breaking API changes

##### High impact

- The constructors of all the widgets now also return an error so that they
  can validate the options. This is a breaking change for the following
  widgets: BarChart, Gauge, LineChart, SparkLine, Text. The callers will have
  to handle the returned error.

##### Low impact

- The container package no longer exports separate methods to receive Keyboard
  and Mouse events which were replaced by a Subscribe method for the event
  distribution system. This shouldn't affect users as the removed methods
  aren't needed by container users.
- The widgetapi.Options struct now uses an enum instead of a boolean when
  widget specifies if it wants keyboard or mouse events. This only impacts
  development of new widgets.

### Fixed

- The LineChart widget now correctly determines the Y axis scale when multiple
  series are provided.
- Lint issues in the codebase, and updated Travis configuration so that golint
  is executed on every run.
- Termdash now correctly starts in locales like zh_CN.UTF-8 where some of the
  characters it uses internally can have ambiguous width.

## [0.6.1] - 12-Feb-2019

### Fixed

- The LineChart widget now correctly places custom labels.

## [0.6.0] - 07-Feb-2019

### Added

- The SegmentDisplay widget.
- A CHANGELOG.
- New line styles for borders.

### Changed

- Better recordings of the individual demos.

### Fixed

- The LineChart now has an option to change the behavior of the Y axis from
  zero anchored to adaptive.
- Lint errors reported on the Go report card.
- Widgets now correctly handle a race when new user data are supplied between
  calls to their Options() and Draw() methods.

## [0.5.0] - 21-Jan-2019

### Added

- Draw primitives for drawing circles.
- The Donut widget.

### Fixed

- Bugfixes in the braille canvas.
- Lint errors reported on the Go report card.
- Flaky behavior in termdash_test.

## [0.4.0] - 15-Jan-2019

### Added

- 256 color support.
- Variable size container splits.
- A more complete demo of the functionality.

### Changed

- Updated documentation and README.

## [0.3.0] - 13-Jan-2019

### Added

- Primitives for drawing lines.
- Implementation of a Braille canvas.
- The LineChart widget.

## [0.2.0] - 02-Jul-2018

### Added

- The SparkLine widget.
- The BarChart widget.
- Manually triggered redraw.
- Travis now checks for presence of licence headers.

### Fixed

- Fixing races in termdash_test.

## 0.1.0 - 13-Jun-2018

### Added

- Documentation of the project and its goals.
- Drawing infrastructure.
- Testing infrastructure.
- The Gauge widget.
- The Text widget.

[unreleased]: https://github.com/mum4k/termdash/compare/v0.9.1...devel
[0.9.1]: https://github.com/mum4k/termdash/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/mum4k/termdash/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/mum4k/termdash/compare/v0.7.2...v0.8.0
[0.7.2]: https://github.com/mum4k/termdash/compare/v0.7.1...v0.7.2
[0.7.1]: https://github.com/mum4k/termdash/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/mum4k/termdash/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/mum4k/termdash/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/mum4k/termdash/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/mum4k/termdash/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/mum4k/termdash/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/mum4k/termdash/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/mum4k/termdash/compare/v0.1.0...v0.2.0
