# Developing a new widget

## The API

A widget is an object that implements the **widgetapi.Widget** interface. Apart
from implementing this interface, each widget exposes other methods that allow
the callers to change its content. E.g. the **gauge** widget enables the
callers to set the displayed percentage.

## Thread safety

All widget implementations must be thread safe, since the infrastructure calls
the widget's **Draw()** method concurrently with the user of the widget setting
the displayed values.

## Drawing the widget's content

When the widget's **Draw()** method is called, the infrastructure provides the
widget with a canvas to draw on. This canvas is always zero based (the first
point is at image.Point{0, 0}) regardless of the actual position of the widget
on the terminal.

## Scaling

Each time the widget's **Draw()** method is called, the widget must determine
the size of the received canvas and scale accordingly. The size of the terminal
might have been changed since the last call to **Draw()**.

Correctly scaling the drawn content on each call also enables the widgets to
size correctly regardless of the size and position of the container they are
placed in.

## Limits

Widget's should utilize the **widgetapi.Options** to set limits on the provided
canvas in order to handle under sized or over sized terminals gracefully.

If the current size of the terminal and the configured container splits result
in a canvas smaller than the **MinimumSize**, the infrastructure won't call the
widget's **Draw()** method. The widgets can use this to prevent impossible
scenarios where an error would have to be returned.

If the container configuration results in a canvas larger than **MaximumSize**
the canvas will be limited to the specified size. Widgets can either specify a
limit for both the maximum width and height or limit just one of them.

## Unit tests

Unit tests utilize the **faketerm** package which is a fake implementation of a
terminal. It creates an in-memory canvas where widgets can draw. The
**faketerm** package also exports the **faketerm.Diff** function which allows
the comparison of two fake terminals giving a human readable output for unit
tests.

A typical unit test creates the expected fake terminal, executes the widget to
get the actual fake terminal and compares the two:

```go
TestWidget(t *testing.T) {
  tests := []struct {
    desc    string
    canvas  image.Rectangle
    opts    []Option
    want    func(size image.Point) *faketerm.Terminal
    wantErr bool
  }{
    {
      desc: "a test case",
      // canvas determines the size of the allocated canvas in the test case.
      canvas: image.Rect(0,0,10,10),
      // want creates the expected content on the fake terminal.
      want: func(size image.Point) *faketerm.Terminal {
        ft := faketerm.MustNew(size)
        c := testcanvas.MustNew(ft.Area())

        // Utilize functions in the testdraw package to create the expected content.
        testcanvas.MustApply(c, ft)
        return ft
      },
    },
  }

  for _, tc := range tests {
    t.Run(tc.desc, func(t *testing.T) {
      c, err := canvas.New(tc.canvas)
      if err != nil {
        t.Fatalf("canvas.New => unexpected error: %v", err)
      }

      widget := New()
      err = widget.Draw(c)
      if (err != nil) != tc.wantErr {
        t.Errorf("Draw => unexpected error: %v, wantErr: %v", err, tc.wantErr)
      }
      if err != nil {
        return
      }

      got, err := faketerm.New(c.Size())
      if err != nil {
        t.Fatalf("faketerm.New => unexpected error: %v", err)
      }

      if err := c.Apply(got); err != nil {
        t.Fatalf("Apply => unexpected error: %v", err)
      }

      if diff := faketerm.Diff(tc.want(c.Size()), got); diff != "" {
        t.Errorf("Draw => %v", diff)
      }
    })
  }
}
```

## Demo and recording for the widget

Once the widget is completed, add a demo into a **demo** sub directory under
the widget's package and record a GIF of the demo. Place the recorded GIF into
the [README](http://github.com/mum4k/termdash).
