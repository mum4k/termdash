# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- A function that draws text in vertically.
- The LineChart widget can display X axis labels in vertical orientation.
- The LineChart widget allows the user to specify a custom scale for the Y
  axis.
- The Text widget now has a Write option that atomically replaces the entire
  text content.

### Changed

- Termbox is now initialized in 256 color mode by default.
- Generalized mouse button FSM for use in widgets that need to track mouse
  button clicks.

#### Breaking API changes

- The constructors of all the widgets now also return an error so that they
  can validate the options. This is a breaking change for the following
  widgets: BarChart, Gauge, LineChart, SparkLine, Text. The callers will have
  to handle the returned error.

### Fixed

- The LineChart widget now correctly determines the Y axis scale when multiple
  series are provided.

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

[Unreleased]: https://github.com/mum4k/termdash/compare/v0.6.1...devel
[0.6.1]: https://github.com/mum4k/termdash/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/mum4k/termdash/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/mum4k/termdash/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/mum4k/termdash/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/mum4k/termdash/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/mum4k/termdash/compare/v0.1.0...v0.2.0
