// Package main demonstrates the 3D widget with a spinning UTF-8 symbol or
// an arbitrary image / logo supplied via the -image flag.
//
// Usage:
//
//	go run ./widgets/threed/threeddemo                       # default emoji spinner
//	go run ./widgets/threed/threeddemo -image /path/logo.png # spin a PNG logo
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/threed"
)

const symbolDemoFrame = "👌"

func main() {
	imagePath := flag.String("image", "", "path to a PNG/JPEG/GIF image or logo to spin in 3D")
	flag.Parse()

	term, err := tcell.New()
	if err != nil {
		log.Fatalf("failed to create terminal: %v", err)
	}
	defer term.Close()

	widget, err := threed.New(
		threed.RotationStep(0.08),
		threed.ZoomScale(12.0),
		threed.UprightOnly(true),
		threed.ShowAxes(false),
		threed.AmbientColor(threed.Color{R: 0.92, G: 0.92, B: 0.92}),
		threed.DiffuseColor(threed.Color{R: 1.0, G: 1.0, B: 1.0}),
		threed.SpecularColor(threed.Color{R: 0.18, G: 0.18, B: 0.18}),
		threed.Shininess(18.0),
		threed.BackfaceCulling(false),
		threed.EnableLogging(false),
	)
	if err != nil {
		log.Fatalf("failed to create ThreeD widget: %v", err)
	}

	// --- Determine which model to spin ---
	var imageModel *threed.Model
	borderTitle := "3D UTF-8 Symbol Spinner (left/right: orbit, scroll: zoom, q: quit)"

	if *imagePath != "" {
		m, err := threed.LoadImageModel(*imagePath)
		if err != nil {
			log.Fatalf("failed to load image %q: %v", *imagePath, err)
		}
		if m == nil {
			log.Fatalf("image %q produced no renderable geometry — try a PNG with a transparent background", *imagePath)
		}
		imageModel = m
		borderTitle = "3D Image Spinner — " + *imagePath + " (left/right: orbit, scroll: zoom, q: quit)"
	}

	c, err := container.New(
		term,
		container.BorderTitle(borderTitle),
		container.PlaceWidget(widget),
	)
	if err != nil {
		log.Fatalf("failed to create container: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		step := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if imageModel != nil {
					// Image model: static geometry, just rotate each frame.
					widget.SetModel(imageModel)
				} else {
					// Emoji spinner: model is regenerated per step for color animation.
					widget.SetModel(threed.NewAnimatedSymbolSpinner(symbolDemoFrame, step))
				}
				widget.Rotate(threed.Vector3D{Y: 0.015})
				step++
			}
		}
	}()

	globalKeyHandler := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyCtrlC || k.Key == keyboard.KeyEsc || k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, term, c,
		termdash.KeyboardSubscriber(globalKeyHandler),
		termdash.RedrawInterval(100*time.Millisecond),
	); err != nil {
		log.Fatalf("failed to run termdash: %v", err)
	}
}
