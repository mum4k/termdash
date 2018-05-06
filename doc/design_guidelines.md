# Design guidelines

## Don't clutter the widget code with drawing primitives

The widget implementations should contain high level code only. Low level
drawing primitives should be in separate packages. That way the widgets remain
easy to understand, enhance and test.

E.g. the **gauge** widget contains code that calculates the size of the
rectangle that needs to be drawn. It doesn't contain code that draws the
rectangle itself as that belongs into the **draw** package.

## Provide test helpers for all functions in the draw package

To simplify unit tests of widgets, a test helper should be provided to all
functions in the **draw** package.

E.g. a function called **Rectangle()** that draws a rectangle should come with
a helper caller **MustRectangle()**. Tests of a widget that uses
**Rectangle()** can just specify the expected rectangle by calling
**MustRectangle()** on the test canvas.
