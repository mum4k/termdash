# Terminal Dashboard High-Level Design

## Objective

Develop infrastructure of dashboard widgets. The widgets should support both
input (mouse and keyboard) and output (display of information to the user).

Fulfill the requirements outlined in the main
[README](http://github.com/mum4k/termdash).

## Background

The terminal dashboard allows placement of configurable widgets onto the terminal.

A widget displays some information to the user, e.g. A graph, a chart, a
progress bar. A widget can receive information from the user in the form of
events, e.g. Mouse or keyboard input.

The widgets aren't placed onto the terminal directly, instead the terminal is
organized into containers. Each container can contain either a widget or
other containers.

## Overview

The terminal dashboard consists of the following layers:

- Terminal.
- Infrastructure.
- Widgets.

The **terminal layer** abstracts the terminal implementation. A real terminal
implementation is used in production when displaying data to the user. A fake
terminal implementation is used in widget unit tests and system tests. Other
implementations are possible, e.g. Image export. The terminal layer is private,
neither the users of this library nor the widgets interact with the terminal
directly.

The **infrastructure layer** is responsible for container management, tracking
of keyboard and mouse focus and handling external events like resizing of the
terminal. The infrastructure layer also decides when to flush the buffer and
refresh the screen. I.e. The widgets update content of a back buffer and the
infrastructure decides when it is synchronized to the terminal.

The **widgets layer** contains the implementations of individual widgets. Each
widget receives a canvas from the container on which it presents its content to
the user. Widgets indicate to the infrastructure layer if they support input
events, which are then forwarded from the infrastructure layer.

The user interacts with the widget API when constructing individual widgets and
with the container API when placing the widgets onto the dashboard.

<p align="center">
  <img src="hld.png" width="50%">
</p>

## Detailed design

### Terminal

The terminal provides access to the input and output.

It allows to:

- Set values and attributes of cells on a back buffer representing a 2-D
  canvas.
- Flush the content of the back buffer to the output.
- Manipulate the cursor position and visibility.
- Read input events (keyboard, mouse, terminal resize, etc...).

The terminal buffers input events until they are read by the client. The buffer
is bound, if the client isn't picking up events fast enough, new events are
dropped and a message is logged.

### Infrastructure

The infrastructure handles terminal setup, input events and manages containers.

#### Keyboard and mouse input

The raw keyboard and mouse events received from the terminal are pre-processed
by the infrastructure. The pre-processing involves recognizing keyboard
shortcuts (i.e. Key combination). The infrastructure recognizes globally
configurable keyboard shortcuts that are processed by the infrastructure. All
other keyboard and mouse events are forwarded to the currently focused widget.

#### Input focus

The infrastructure tracks focus. Only the focused widget receives keyboard and
mouse events. Focus can be changed using mouse or global keyboard shortcuts.
The focused widget is highlighted on the dashboard.

#### Containers

The container provides a way of splitting the dashboard down to smaller
elements where individual widgets can be placed. Each container can be split to
multiple sub containers. A container contains either a sub container or a
widget.

Container is responsible for coordinate translation. Each widget receives a
virtual canvas it can draw on. Each of these canvases starts at coordinates
image.Point{0, 0}. The parent container translates coordinates from the virtual
canvas to the real terminal. The container therefore enforces limits for widgets.

Containers can be styled with borders and other options.

#### Flushing the terminal

All widgets indirectly write to the back buffer of the terminal implementation. The changes
to the back buffer only become visible when the infrastructure flushes its content.

Widgets cannot force a flush, but they can indicate that a flush is desired.
The infrastructure throttles the amount of times this happens.

#### Terminal resizing

The terminal resize events are processed by the infrastructure. Each widget
indicates its desired and minimum size for its canvas when registering with its
parent container.

The parent container in turn informs the widget what is the actual size of its
canvas. The infrastructure guarantees that the actual size won't ever be
smaller than the advertised minimum and guarantees that the size will keep the
aspect ration requested by the widget.

When the size of terminal changes, the infrastructure resizes all containers
according to the rules outlined above, asks all widgets to redraw their
canvases and flushes to the back buffer to the terminal.

### Widgets

Users of the terminal dashboard construct the widgets directly. Therefore each
widget can define its own options and API for setting values (e.g. The
displayed percentage on a progress bar). The users then create the desired
container splits and place each widget into a dedicated container.

Each widget receives a canvas from the parent container, the widget can draw
anything on the canvas as long as it stays within the limits. Helper libraries
are developed that allow placement and drawing of common elements like lines or
geometrical shapes.

## APIs

### Terminal API

The Terminal API is an interface private to the terminal dashboard library. Its
primary purpose is to act as a shim layer over different terminal
implementations.

The Terminal API is defined in the
[terminalapi](http://github.com/mum4k/termdash/terminalapi/terminalapi.go)
package.

The **Event()** method returns the next input event. Different input event
types are defined in the
[event.go](http://github.com/mum4k/termdash/terminalapi/event.go)
file.

### Container API

The Container API is used to split the terminal and place the widgets. Each
container can be split to two sub containers or have a widget placed into it.
A container can be split either horizontally or vertically.

The containers further accept styling options and alignment options. The
following indicates how the Container API will be used.

The Container API is defined in the
[container](http://github.com/mum4k/termdash/container/container.go)
package.

A demonstration how this is used from the client perspective is in the
[container_test.go](http://github.com/mum4k/termdash/container/container_test.go)
file.

### Widget API

Each widget must implement the Widget API. All widget implementations must
be thread-safe since the calls that update the displayed values come in
concurrently with requests and events from the infrastructure.

The Widget API is defined in the
[widget](http://github.com/mum4k/termdash/widget/widget.go)
package.

Each widget gets a Canvas to draw on. The Canvas API is defined in the
[canvas](http://github.com/mum4k/termdash/canvas/canvas.go)
package.

## Testing plan

Unit test helpers are provided with the terminal dashboard library, these include:

- A fake implementation of the terminal API.
- Unit test comparison helpers to verify the content of the fake terminal.
- Visualization tools to display differences between the expected and the actual.

## Document history

Date        | Author | Description
------------|--------|---------------
24-Mar-2018 | mum4k  | Initial draft.
