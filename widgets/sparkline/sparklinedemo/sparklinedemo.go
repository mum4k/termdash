// Binary sparklinedemo displays a couple of SparkLine widgets.
// Exist when 'q' is pressed.
package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgets/sparkline"
)

// playSparkLine continuously adds values to the SparkLine, once every delay.
// Exits when the context expires.
func playSparkLine(ctx context.Context, sl *sparkline.SparkLine, delay time.Duration) {
	const max = 100

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			v := int(rand.Int31n(max + 1))
			if err := sl.Add(v); err != nil {
				panic(err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func main() {
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	bc := sparkline.New()
	go playSparkLine(ctx, bc, 1*time.Second)

	c := container.New(
		t,
		container.Border(draw.LineStyleLight),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(bc),
	)

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}
