package main

import (
	"context"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/pie"
	"time"
)

func playPie(ctx context.Context, p *pie.Pie, values [][]int, delay time.Duration) {
	idx := 0
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = p.Values(values[idx])
			idx = (idx + 1) % len(values)
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())

	pie1, _ := pie.New()
	_ = pie1.Values([]int{30, 70})
	go playPie(ctx, pie1, [][]int{{30, 70}, {50, 50}, {80, 20}}, 1*time.Second)

	pie2, _ := pie.New()
	_ = pie2.Values([]int{10, 20, 30, 40})
	go playPie(ctx, pie2, [][]int{{10, 20, 30, 40}, {40, 30, 20, 10}, {25, 25, 25, 25}}, 2*time.Second)

	c, _ := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitVertical(
			container.Left(container.PlaceWidget(pie1)),
			container.Right(container.PlaceWidget(pie2)),
		),
	)

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(1*time.Second)); err != nil {
		panic(err)
	}
}
