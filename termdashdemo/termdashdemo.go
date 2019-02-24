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
// Exist when 'q' is pressed.
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
)

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// layout prepares the screen layout by creating the container and placing
// widgets.
func layout(ctx context.Context, t terminalapi.Terminal) (*container.Container, error) {
	sd, err := newSegmentDisplay(ctx)
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
	segmentTextSpark := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Press Q to quit"),
				container.PlaceWidget(sd),
			),
			container.Bottom(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("A rolling text"),
						container.PlaceWidget(rollT),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("A SparkLine group"),
						container.SplitHorizontal(
							container.Top(container.PlaceWidget(spGreen)),
							container.Bottom(container.PlaceWidget(spRed)),
						),
					),
				),
			),
			container.SplitPercent(50),
		),
	}

	g, err := newGauge(ctx)
	if err != nil {
		return nil, err
	}

	heartLC, err := newHeartbeat(ctx)
	if err != nil {
		return nil, err
	}
	gaugeAndHeartbeat := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("A Gauge"),
				container.BorderColor(cell.ColorNumber(39)),
				container.PlaceWidget(g),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("A LineChart"),
				container.PlaceWidget(heartLC),
			),
			container.SplitPercent(20),
		),
	}

	leftSide := []container.Option{
		container.SplitHorizontal(
			container.Top(segmentTextSpark...),
			container.Bottom(gaugeAndHeartbeat...),
			container.SplitPercent(50),
		),
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
	lcAndButtons := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Multiple series"),
				container.BorderTitleAlignRight(),
				container.PlaceWidget(sineLC),
			),
			container.Bottom(
				container.SplitVertical(
					container.Left(
						container.PlaceWidget(leftB),
						container.AlignHorizontal(align.HorizontalRight),
					),
					container.Right(
						container.PlaceWidget(rightB),
						container.AlignHorizontal(align.HorizontalLeft),
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
				container.PlaceWidget(bc),
				container.BorderTitleAlignRight(),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("A Donut"),
						container.BorderTitleAlignRight(),
						container.PlaceWidget(don),
					),
					container.Bottom(lcAndButtons...),
					container.SplitPercent(30),
				),
			),
			container.SplitPercent(30),
		),
	}

	c, err := container.New(
		t,
		container.SplitVertical(
			container.Left(leftSide...),
			container.Right(rightSide...),
			container.SplitPercent(70),
		),
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func main() {
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	c, err := layout(ctx, t)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
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

// newSegmentDisplay creates a new SegmentDisplay that shows the Termdash name.
func newSegmentDisplay(ctx context.Context) (*segmentdisplay.SegmentDisplay, error) {
	sd, err := segmentdisplay.New()
	if err != nil {
		return nil, err
	}

	const text = "Termdash"
	colors := map[rune]cell.Color{
		'T': cell.ColorBlue,
		'e': cell.ColorRed,
		'r': cell.ColorYellow,
		'm': cell.ColorBlue,
		'd': cell.ColorGreen,
		'a': cell.ColorRed,
		's': cell.ColorGreen,
		'h': cell.ColorRed,
	}

	var state []rune
	for i := 0; i < len(text); i++ {
		state = append(state, ' ')
	}
	state = append(state, []rune(text)...)
	go periodic(ctx, 500*time.Millisecond, func() error {
		var chunks []*segmentdisplay.TextChunk
		for i := 0; i < len(text); i++ {
			chunks = append(chunks, segmentdisplay.NewChunk(
				string(state[i]),
				segmentdisplay.WriteCellOpts(cell.FgColor(colors[state[i]])),
			))
		}
		if err := sd.Write(chunks); err != nil {
			return err
		}
		state = rotateRunes(state, 1)
		return nil
	})
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
		sparkline.Label("Green SparkLine", cell.FgColor(cell.ColorBlue)),
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
		sparkline.Label("Red SparkLine", cell.FgColor(cell.ColorBlue)),
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

	rightB, err := button.New("(r)ight", func() error {
		secondDist.add(-diff)
		return nil
	},
		button.GlobalKey('r'),
		button.FillColor(cell.ColorNumber(196)),
	)
	return leftB, rightB, sineLc, nil
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
