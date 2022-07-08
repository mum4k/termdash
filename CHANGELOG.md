# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.17.0] - 07-Jul-2022

### Added

- Added support for `tcell`'s `Dim` text style.

### Changed

- Bump github.com/gdamore/tcell from 2.4.0 to 2.5.1.
- Bump github.com/nsf/termbox-go to v1.1.1.
- Change the Go version in `go.mod` to 1.17.
- Executed `go mod tidy`.

### Fixed

- Fixed invalid path to the `autogen_licences.sh` script in `.travis.ci`.

## [0.16.1] - 13-Jan-2022

### Changed

- Change the Go versions the CI builds with up to 1.17.
- Bump github.com/gdamore/tcell from 2.2.0 to 2.4.0.
- Bump github.com/mattn/go-runewidth from 0.0.12 to 0.0.13.

## [0.16.0] - 03-Apr-2021

### Added

- The `Text` widget has a new option `MaxTextCells` which can be used to limit
  the maximum number of cells the widget keeps in memory.

### Changed

- Bump github.com/mattn/go-runewidth from 0.0.10 to 0.0.12.

## [0.15.0] - 06-Mar-2021

### Changed

- Bump github.com/gdamore/tcell/v2 from 2.0.0 to 2.2.0.
- Bump github.com/mattn/go-runewidth from 0.0.9 to 0.0.10.
- Allowing CI to modify go.mod and go.sum when necessary.
- Executed `go mod tidy`.

### Added

- TitleColor and TitleFocusedColor options for border title which enables the
  setting of separate colors for border and title on a container.

## [0.14.0] - 30-Dec-2020

### Breaking API changes

- The `widgetapi.Widget.Keyboard` and `widgetapi.Widget.Mouse` methods now
  accepts a second argument which provides widgets with additional metadata.
  All widgets implemented outside of the `termdash` repository will need to be
  updated similarly to the `Barchart` example below. Change the original method
  signatures:
  ```go
  func (*BarChart) Keyboard(k *terminalapi.Keyboard) error { ... }

  func (*BarChart) Mouse(m *terminalapi.Mouse) error { ... }

  ```

  By adding the new `*widgetapi.EventMeta` argument as follows:
  ```go
  func (*BarChart) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error { ... }

  func (*BarChart) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error { ... }
  ```

### Fixed

- `termdash` no longer crashes when `tcell` is used and the terminal window
  downsizes while content is being drawn.

### Added

#### Text input form functionality with keyboard navigation

- added a new `formdemo` that demonstrates a text input form with keyboard
  navigation.

#### Infrastructure changes

- `container` now allows users to configure keyboard keys that move focus to
  the next or the previous container.
- containers can request to be skipped when focus is moved using keyboard keys.
- containers can register into separate focus groups and specific keyboard keys
  can be configured to move the focus within each focus group.
- widgets can now request keyboard events exclusively when focused.
- users can now set a `container` as focused using the new `container.Focused`
  option.

#### Updates to the `button` widget

- the `button` widget allows users to specify multiple trigger keys.
- the `button` widget now supports different keys for the global and focused
  scope.
- the `button` widget can now be drawn without the shadow or the press
  animation.
- the `button` widget can now be drawn without horizontal padding around its
  text.
- the `button` widget now allows specifying cell options for each cell of the
  displayed text. Separate cell options can be specified for each of button's
  main states (up, focused and up, down).
- the `button` widget allows specifying separate fill color values for each of
  its main states (up, focused and up, down).
- the `button` widget now has a method `SetCallback` that allows updating the
  callback function on an existing `button` instance.

#### Updates to the `textinput` widget

- the `textinput` widget can now be configured to request keyboard events
  exclusively when focused.
- the `textinput` widget can now be initialized with a default text in the
  input box.

## [0.13.0] - 17-Nov-2020

### Added

- the `Text` widget now allows user to specify custom scroll marker runes.

### Changed

- terminal cells now support font modifier options (bold, italic,
  underline, strike through).
- `tcell` dependency was upgraded to v2.0.0.
- upgraded versions of all other dependencies.
- aligned the definition of the first 16 colors with the definition used by
  Xterm and `tcell`. Defined two non-standard colors `ColorMagenta` and
  `ColorCyan` to make this change backward compatible for users that use
  `termbox-go`.
- made `tcell` terminal implementation the default in examples, demos and
  documentation.

### Fixed

- coveralls again triggers and reports on PRs.
- addressed some lint issues.
- improved test coverage in some modules.
- changed the Blue color in demos to a more visible shade.
- fixed a bug where segment display text in `termdashdemo` appeared to be
  jumping.

## [0.12.2] - 31-Aug-2020

### Fixed

- advanced the CI Go versions up to Go 1.15.
- fixed the build status badge to correctly point to travis-ci.com instead of
  travis-ci.org.

## [0.12.1] - 20-Jun-2020

### Fixed

- the `tcell` unit test can now pass in headless mode (when TERM="") which
  happens under bazel.
- switching coveralls integration to Github application.

## [0.12.0] - 10-Apr-2020

### Added

- Migrating to [Go modules](https://blog.golang.org/using-go-modules).
- Renamed directory `internal` to `private` so that external widget development
  is possible. Noted in
  [README.md](https://github.com/mum4k/termdash/blob/master/README.md) that packages in the
  `private` directory don't have any API stability guarantee.

## [0.11.0] - 7-Mar-2020

#### Breaking API changes

- Termdash now requires at least Go version 1.11.

### Added

- New [`tcell`](https://github.com/gdamore/tcell) based terminal implementation
  which implements the `terminalapi.Terminal` interface.
- tcell implementation supports two initialization `Option`s:
  - `ColorMode` the terminal color output mode (defaults to 256 color mode)
  - `ClearStyle` the foreground and background color style to use when clearing
     the screen (defaults to the global ColorDefault for both foreground and
     background)

### Fixed

- Improved test coverage of the `Gauge` widget.

## [0.10.0] - 5-Jun-2019

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

[unreleased]: https://github.com/mum4k/termdash/compare/v0.17.0...devel
[0.17.0]: https://github.com/mum4k/termdash/compare/v0.16.1...v0.17.0
[0.16.1]: https://github.com/mum4k/termdash/compare/v0.16.0...v0.16.1
[0.16.0]: https://github.com/mum4k/termdash/compare/v0.15.0...v0.16.0
[0.15.0]: https://github.com/mum4k/termdash/compare/v0.14.0...v0.15.0
[0.14.0]: https://github.com/mum4k/termdash/compare/v0.13.0...v0.14.0
[0.13.0]: https://github.com/mum4k/termdash/compare/v0.12.2...v0.13.0
[0.12.2]: https://github.com/mum4k/termdash/compare/v0.12.1...v0.12.2
[0.12.1]: https://github.com/mum4k/termdash/compare/v0.12.0...v0.12.1
[0.12.0]: https://github.com/mum4k/termdash/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/mum4k/termdash/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/mum4k/termdash/compare/v0.9.1...v0.10.0
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
