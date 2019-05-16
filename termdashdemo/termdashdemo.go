// Copyright 2019 Google Inc.
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

// Binary termdashdemo demonstrates the functionality of termdash and its various widgets.
// Exits when 'q' is pressed.
package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
)

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// widgets holds the widgets used by this demo.
type widgets struct {
	segDist  *segmentdisplay.SegmentDisplay
	input    *textinput.TextInput
	rollT    *text.Text
	spGreen  *sparkline.SparkLine
	spRed    *sparkline.SparkLine
	gauge    *gauge.Gauge
	heartLC  *linechart.LineChart
	barChart *barchart.BarChart
	donut    *donut.Donut
	leftB    *button.Button
	rightB   *button.Button
	sineLC   *linechart.LineChart

	buttons *layoutButtons
}

// newWidgets creates all widgets used by this demo.
func newWidgets(ctx context.Context, c *container.Container) (*widgets, error) {
	updateText := make(chan string)
	sd, err := newSegmentDisplay(ctx, updateText)
	if err != nil {
		return nil, err
	}

	input, err := newTextInput(updateText)
	if err != nil {
		return nil, err
	}

	rollT, err := newRollText(ctx)
	if err != nil {
		return nil, err
	}
	spGreen, spRed, err := newSparkLines(ctx)
	if err != nil {
		return nil, err
	}
	g, err := newGauge(ctx)
	if err != nil {
		return nil, err
	}

	heartLC, err := newHeartbeat(ctx)
	if err != nil {
		return nil, err
	}

	bc, err := newBarChart(ctx)
	if err != nil {
		return nil, err
	}

	don, err := newDonut(ctx)
	if err != nil {
		return nil, err
	}

	leftB, rightB, sineLC, err := newSines(ctx)
	if err != nil {
		return nil, err
	}
	return &widgets{
		segDist:  sd,
		input:    input,
		rollT:    rollT,
		spGreen:  spGreen,
		spRed:    spRed,
		gauge:    g,
		heartLC:  heartLC,
		barChart: bc,
		donut:    don,
		leftB:    leftB,
		rightB:   rightB,
		sineLC:   sineLC,
	}, nil
}

// layoutType represents the possible layouts the buttons switch between.
type layoutType int

const (
	// layoutAll displays all the widgets.
	layoutAll layoutType = iota
	// layoutText focuses onto the text widget.
	layoutText
	// layoutSparkLines focuses onto the sparklines.
	layoutSparkLines
	// layoutLineChart focuses onto the linechart.
	layoutLineChart
)

// gridLayout prepares container options that represent the desired screen layout.
// This function demonstrates the use of the grid builder.
// gridLayout() and contLayout() demonstrate the two available layout APIs and
// both produce equivalent layouts for layoutType layoutAll.
func gridLayout(w *widgets, lt layoutType) ([]container.Option, error) {
	leftRows := []grid.Element{
		grid.RowHeightPerc(25,
			grid.Widget(w.segDist,
				container.Border(linestyle.Light),
				container.BorderTitle("Press Esc to quit"),
			),
		),
		grid.RowHeightPerc(5,
			grid.Widget(w.input),
		),

		grid.RowHeightPerc(5,
			grid.ColWidthPerc(25,
				grid.Widget(w.buttons.allB),
			),
			grid.ColWidthPerc(25,
				grid.Widget(w.buttons.textB),
			),
			grid.ColWidthPerc(25,
				grid.Widget(w.buttons.spB),
			),
			grid.ColWidthPerc(25,
				grid.Widget(w.buttons.lcB),
			),
		),
	}
	switch lt {
	case layoutAll:
		leftRows = append(leftRows,
			grid.RowHeightPerc(20,
				grid.ColWidthPerc(50,
					grid.Widget(w.rollT,
						container.Border(linestyle.Light),
						container.BorderTitle("A rolling text"),
					),
				),
				grid.ColWidthPerc(50,
					grid.RowHeightPerc(50,
						grid.Widget(w.spGreen,
							container.Border(linestyle.Light),
							container.BorderTitle("Green SparkLine"),
						),
					),
					grid.RowHeightPerc(50,
						grid.Widget(w.spRed,
							container.Border(linestyle.Light),
							container.BorderTitle("Red SparkLine"),
						),
					),
				),
			),
			grid.RowHeightPerc(7,
				grid.Widget(w.gauge,
					container.Border(linestyle.Light),
					container.BorderTitle("A Gauge"),
					container.BorderColor(cell.ColorNumber(39)),
				),
			),
			grid.RowHeightPerc(38,
				grid.Widget(w.heartLC,
					container.Border(linestyle.Light),
					container.BorderTitle("A LineChart"),
				),
			),
		)
	case layoutText:
		leftRows = append(leftRows,
			grid.RowHeightPerc(65,
				grid.Widget(w.rollT,
					container.Border(linestyle.Light),
					container.BorderTitle("A rolling text"),
				),
			),
		)

	case layoutSparkLines:
		leftRows = append(leftRows,
			grid.RowHeightPerc(32,
				grid.Widget(w.spGreen,
					container.Border(linestyle.Light),
					container.BorderTitle("Green SparkLine"),
				),
			),
			grid.RowHeightPerc(33,
				grid.Widget(w.spRed,
					container.Border(linestyle.Light),
					container.BorderTitle("Red SparkLine"),
				),
			),
		)

	case layoutLineChart:
		leftRows = append(leftRows,
			grid.RowHeightPerc(65,
				grid.Widget(w.heartLC,
					container.Border(linestyle.Light),
					container.BorderTitle("A LineChart"),
				),
			),
		)
	}

	builder := grid.New()
	builder.Add(
		grid.ColWidthPerc(70, leftRows...),
	)

	builder.Add(
		grid.ColWidthPerc(30,
			grid.RowHeightPerc(30,
				grid.Widget(w.barChart,
					container.Border(linestyle.Light),
					container.BorderTitle("BarChart"),
					container.BorderTitleAlignRight(),
				),
			),
			grid.RowHeightPerc(21,
				grid.Widget(w.donut,
					container.Border(linestyle.Light),
					container.BorderTitle("A Donut"),
					container.BorderTitleAlignRight(),
				),
			),
			grid.RowHeightPerc(40,
				grid.Widget(w.sineLC,
					container.Border(linestyle.Light),
					container.BorderTitle("Multiple series"),
					container.BorderTitleAlignRight(),
				),
			),
			grid.RowHeightPerc(9,
				grid.ColWidthPerc(50,
					grid.Widget(w.leftB,
						container.AlignHorizontal(align.HorizontalRight),
						container.PaddingRight(1),
					),
				),
				grid.ColWidthPerc(50,
					grid.Widget(w.rightB,
						container.AlignHorizontal(align.HorizontalLeft),
						container.PaddingLeft(1),
					),
				),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return gridOpts, nil
}

// contLayout prepares container options that represent the desired screen layout.
// This function demonstrates the direct use of the container API.
// gridLayout() and contLayout() demonstrate the two available layout APIs and
// both produce equivalent layouts for layoutType layoutAll.
// contLayout only produces layoutAll.
func contLayout(w *widgets) ([]container.Option, error) {
	buttonRow := []container.Option{
		container.SplitVertical(
			container.Left(
				container.SplitVertical(
					container.Left(
						container.PlaceWidget(w.buttons.allB),
					),
					container.Right(
						container.PlaceWidget(w.buttons.textB),
					),
				),
			),
			container.Right(
				container.SplitVertical(
					container.Left(
						container.PlaceWidget(w.buttons.spB),
					),
					container.Right(
						container.PlaceWidget(w.buttons.lcB),
					),
				),
			),
		),
	}

	textAndSparks := []container.Option{
		container.SplitVertical(
			container.Left(
				container.Border(linestyle.Light),
				container.BorderTitle("A rolling text"),
				container.PlaceWidget(w.rollT),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Green SparkLine"),
						container.PlaceWidget(w.spGreen),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Red SparkLine"),
						container.PlaceWidget(w.spRed),
					),
				),
			),
		),
	}

	segmentTextInputSparks := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Press Esc to quit"),
				container.PlaceWidget(w.segDist),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.SplitHorizontal(
							container.Top(
								container.PlaceWidget(w.input),
							),
							container.Bottom(buttonRow...),
						),
					),
					container.Bottom(textAndSparks...),
					container.SplitPercent(40),
				),
			),
			container.SplitPercent(50),
		),
	}

	gaugeAndHeartbeat := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("A Gauge"),
				container.BorderColor(cell.ColorNumber(39)),
				container.PlaceWidget(w.gauge),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("A LineChart"),
				container.PlaceWidget(w.heartLC),
			),
			container.SplitPercent(20),
		),
	}

	leftSide := []container.Option{
		container.SplitHorizontal(
			container.Top(segmentTextInputSparks...),
			container.Bottom(gaugeAndHeartbeat...),
			container.SplitPercent(50),
		),
	}

	lcAndButtons := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Multiple series"),
				container.BorderTitleAlignRight(),
				container.PlaceWidget(w.sineLC),
			),
			container.Bottom(
				container.SplitVertical(
					container.Left(
						container.PlaceWidget(w.leftB),
						container.AlignHorizontal(align.HorizontalRight),
						container.PaddingRight(1),
					),
					container.Right(
						container.PlaceWidget(w.rightB),
						container.AlignHorizontal(align.HorizontalLeft),
						container.PaddingLeft(1),
					),
				),
			),
			container.SplitPercent(80),
		),
	}

	rightSide := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("BarChart"),
				container.PlaceWidget(w.barChart),
				container.BorderTitleAlignRight(),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("A Donut"),
						container.BorderTitleAlignRight(),
						container.PlaceWidget(w.donut),
					),
					container.Bottom(lcAndButtons...),
					container.SplitPercent(30),
				),
			),
			container.SplitPercent(30),
		),
	}

	return []container.Option{
		container.SplitVertical(
			container.Left(leftSide...),
			container.Right(rightSide...),
			container.SplitPercent(70),
		),
	}, nil
}

// rootID is the ID assigned to the root container.
const rootID = "root"

func main() {
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}
	defer t.Close()

	c, err := container.New(t, container.ID(rootID))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	w, err := newWidgets(ctx, c)
	if err != nil {
		panic(err)
	}
	lb, err := newLayoutButtons(c, w)
	if err != nil {
		panic(err)
	}
	w.buttons = lb

	gridOpts, err := gridLayout(w, layoutAll) // equivalent to contLayout(w)
	if err != nil {
		panic(err)
	}

	if err := c.Update(rootID, gridOpts...); err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}
}

// periodic executes the provided closure periodically every interval.
// Exits when the context expires.
func periodic(ctx context.Context, interval time.Duration, fn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// textState creates a rotated state for the text we are displaying.
func textState(text string, capacity, step int) []rune {
	if capacity == 0 {
		return nil
	}

	var state []rune
	for i := 0; i < capacity; i++ {
		state = append(state, ' ')
	}
	state = append(state, []rune(text)...)
	step = step % len(state)
	return rotateRunes(state, step)
}

// newTextInput creates a new TextInput field that changes the text on the
// SegmentDisplay.
func newTextInput(updateText chan<- string) (*textinput.TextInput, error) {
	input, err := textinput.New(
		textinput.Label("Change text to: ", cell.FgColor(cell.ColorBlue)),
		textinput.MaxWidthCells(20),
		textinput.PlaceHolder("enter any text"),
		textinput.OnSubmit(func(text string) error {
			updateText <- text
			return nil
		}),
		textinput.ClearOnSubmit(),
	)
	if err != nil {
		return nil, err
	}
	return input, err
}

// newSegmentDisplay creates a new SegmentDisplay that initially shows the
// Termdash name. Shows any text that is sent over the channel.
func newSegmentDisplay(ctx context.Context, updateText <-chan string) (*segmentdisplay.SegmentDisplay, error) {
	sd, err := segmentdisplay.New()
	if err != nil {
		return nil, err
	}

	colors := []cell.Color{
		cell.ColorBlue,
		cell.ColorRed,
		cell.ColorYellow,
		cell.ColorBlue,
		cell.ColorGreen,
		cell.ColorRed,
		cell.ColorGreen,
		cell.ColorRed,
	}

	text := "Termdash"
	step := 0

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				state := textState(text, sd.Capacity(), step)
				var chunks []*segmentdisplay.TextChunk
				for i := 0; i < sd.Capacity(); i++ {
					if i >= len(state) {
						break
					}

					color := colors[i%len(colors)]
					chunks = append(chunks, segmentdisplay.NewChunk(
						string(state[i]),
						segmentdisplay.WriteCellOpts(cell.FgColor(color)),
					))
				}
				if len(chunks) == 0 {
					continue
				}
				if err := sd.Write(chunks); err != nil {
					panic(err)
				}
				step++

			case t := <-updateText:
				text = t
				sd.Reset()
				step = 0

			case <-ctx.Done():
				return
			}
		}
	}()
	return sd, nil
}

// newRollText creates a new Text widget that displays rolling text.
func newRollText(ctx context.Context) (*text.Text, error) {
	t, err := text.New(text.RollContent())
	if err != nil {
		return nil, err
	}

	i := 0
	go periodic(ctx, 1*time.Second, func() error {
		if err := t.Write(fmt.Sprintf("Writing line %d.\n", i), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(142)))); err != nil {
			return err
		}
		i++
		return nil
	})
	return t, nil
}

// newSparkLines creates two new sparklines displaying random values.
func newSparkLines(ctx context.Context) (*sparkline.SparkLine, *sparkline.SparkLine, error) {
	spGreen, err := sparkline.New(
		sparkline.Color(cell.ColorGreen),
	)
	if err != nil {
		return nil, nil, err
	}

	const max = 100
	go periodic(ctx, 250*time.Millisecond, func() error {
		v := int(rand.Int31n(max + 1))
		return spGreen.Add([]int{v})
	})

	spRed, err := sparkline.New(
		sparkline.Color(cell.ColorRed),
	)
	if err != nil {
		return nil, nil, err
	}
	go periodic(ctx, 500*time.Millisecond, func() error {
		v := int(rand.Int31n(max + 1))
		return spRed.Add([]int{v})
	})
	return spGreen, spRed, nil

}

// newGauge creates a demo Gauge widget.
func newGauge(ctx context.Context) (*gauge.Gauge, error) {
	g, err := gauge.New()
	if err != nil {
		return nil, err
	}

	const start = 35
	progress := start

	go periodic(ctx, 2*time.Second, func() error {
		if err := g.Percent(progress); err != nil {
			return err
		}
		progress++
		if progress > 100 {
			progress = start
		}
		return nil
	})
	return g, nil
}

// newDonut creates a demo Donut widget.
func newDonut(ctx context.Context) (*donut.Donut, error) {
	d, err := donut.New(donut.CellOpts(
		cell.FgColor(cell.ColorNumber(33))),
	)
	if err != nil {
		return nil, err
	}

	const start = 35
	progress := start

	go periodic(ctx, 500*time.Millisecond, func() error {
		if err := d.Percent(progress); err != nil {
			return err
		}
		progress++
		if progress > 100 {
			progress = start
		}
		return nil
	})
	return d, nil
}

// newHeartbeat returns a line chart that displays a heartbeat-like progression.
func newHeartbeat(ctx context.Context) (*linechart.LineChart, error) {
	var inputs []float64
	for i := 0; i < 100; i++ {
		v := math.Pow(math.Sin(float64(i)), 63) * math.Sin(float64(i)+1.5) * 8
		inputs = append(inputs, v)
	}

	lc, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
		return nil, err
	}
	step := 0
	go periodic(ctx, redrawInterval/3, func() error {
		step = (step + 1) % len(inputs)
		return lc.Series("heartbeat", rotateFloats(inputs, step),
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
			linechart.SeriesXLabels(map[int]string{
				0: "zero",
			}),
		)
	})
	return lc, nil
}

// newBarChart returns a BarcChart that displays random values on multiple bars.
func newBarChart(ctx context.Context) (*barchart.BarChart, error) {
	bc, err := barchart.New(
		barchart.BarColors([]cell.Color{
			cell.ColorNumber(33),
			cell.ColorNumber(39),
			cell.ColorNumber(45),
			cell.ColorNumber(51),
			cell.ColorNumber(81),
			cell.ColorNumber(87),
		}),
		barchart.ValueColors([]cell.Color{
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
		}),
		barchart.ShowValues(),
	)
	if err != nil {
		return nil, err
	}

	const (
		bars = 6
		max  = 100
	)
	values := make([]int, bars)
	go periodic(ctx, 1*time.Second, func() error {
		for i := range values {
			values[i] = int(rand.Int31n(max + 1))
		}

		return bc.Values(values, max)
	})
	return bc, nil
}

// distance is a thread-safe int value used by the newSince method.
// Buttons write it and the line chart reads it.
type distance struct {
	v  int
	mu sync.Mutex
}

// add adds the provided value to the one stored.
func (d *distance) add(v int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.v += v
}

// get returns the current value.
func (d *distance) get() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.v
}

// newSines returns a line chart that displays multiple sine series and two buttons.
// The left button shifts the second series relative to the first series to
// the left and the right button shifts it to the right.
func newSines(ctx context.Context) (left, right *button.Button, lc *linechart.LineChart, err error) {
	var inputs []float64
	for i := 0; i < 200; i++ {
		v := math.Sin(float64(i) / 100 * math.Pi)
		inputs = append(inputs, v)
	}

	sineLc, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	step1 := 0
	secondDist := &distance{v: 100}
	go periodic(ctx, redrawInterval/3, func() error {
		step1 = (step1 + 1) % len(inputs)
		if err := lc.Series("first", rotateFloats(inputs, step1),
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorBlue)),
		); err != nil {
			return err
		}

		step2 := (step1 + secondDist.get()) % len(inputs)
		return lc.Series("second", rotateFloats(inputs, step2), linechart.SeriesCellOpts(cell.FgColor(cell.ColorWhite)))
	})

	// diff is the difference a single button press adds or removes to the
	// second series.
	const diff = 20
	leftB, err := button.New("(l)eft", func() error {
		secondDist.add(diff)
		return nil
	},
		button.GlobalKey('l'),
		button.WidthFor("(r)ight"),
		button.FillColor(cell.ColorNumber(220)),
	)
	if err != nil {
		return nil, nil, nil, err
	}

	rightB, err := button.New("(r)ight", func() error {
		secondDist.add(-diff)
		return nil
	},
		button.GlobalKey('r'),
		button.FillColor(cell.ColorNumber(196)),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	return leftB, rightB, sineLc, nil
}

// setLayout sets the specified layout.
func setLayout(c *container.Container, w *widgets, lt layoutType) error {
	gridOpts, err := gridLayout(w, lt)
	if err != nil {
		return err
	}
	return c.Update(rootID, gridOpts...)
}

// layoutButtons are buttons that change the layout.
type layoutButtons struct {
	allB  *button.Button
	textB *button.Button
	spB   *button.Button
	lcB   *button.Button
}

// newLayoutButtons returns buttons that dynamically switch the layouts.
func newLayoutButtons(c *container.Container, w *widgets) (*layoutButtons, error) {
	opts := []button.Option{
		button.WidthFor("sparklines"),
		button.FillColor(cell.ColorNumber(220)),
		button.Height(1),
	}

	allB, err := button.New("all", func() error {
		return setLayout(c, w, layoutAll)
	}, opts...)
	if err != nil {
		return nil, err
	}

	textB, err := button.New("text", func() error {
		return setLayout(c, w, layoutText)
	}, opts...)
	if err != nil {
		return nil, err
	}

	spB, err := button.New("sparklines", func() error {
		return setLayout(c, w, layoutSparkLines)
	}, opts...)
	if err != nil {
		return nil, err
	}

	lcB, err := button.New("linechart", func() error {
		return setLayout(c, w, layoutLineChart)
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &layoutButtons{
		allB:  allB,
		textB: textB,
		spB:   spB,
		lcB:   lcB,
	}, nil
}

// rotateFloats returns a new slice with inputs rotated by step.
// I.e. for a step of one:
//   inputs[0] -> inputs[len(inputs)-1]
//   inputs[1] -> inputs[0]
// And so on.
func rotateFloats(inputs []float64, step int) []float64 {
	return append(inputs[step:], inputs[:step]...)
}

// rotateRunes returns a new slice with inputs rotated by step.
// I.e. for a step of one:
//   inputs[0] -> inputs[len(inputs)-1]
//   inputs[1] -> inputs[0]
// And so on.
func rotateRunes(inputs []rune, step int) []rune {
	return append(inputs[step:], inputs[:step]...)
}
