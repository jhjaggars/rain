package main

import (
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func drop(in chan *imdraw.IMDraw, out chan *imdraw.IMDraw, x float64) {
	y := 758.0

	yspeed := rand.Intn(2) + 2
	startTime := time.Now()
	initialPause := time.Duration(rand.Intn(400)) * time.Millisecond

	for time.Since(startTime) < initialPause {
		imd := <-in
		out <- imd
	}

	for {
		imd := <-in
		imd.Color = pixel.RGB(1, 0, 0)
		imd.Push(pixel.V(x, y))
		imd.Color = pixel.RGB(0, 0, 1)
		imd.Push(pixel.V(x, y+30))
		imd.Line(2)
		out <- imd

		if y <= 0 {
			y = 758
		} else {
			y -= float64(yspeed)
		}
	}
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	chans := make(map[chan *imdraw.IMDraw]chan *imdraw.IMDraw)

	for i := 0; i < 10; i++ {
		in := make(chan *imdraw.IMDraw)
		out := make(chan *imdraw.IMDraw)
		chans[in] = out
		go drop(in, out, float64(102*i))
	}

	for !win.Closed() {

		imd := imdraw.New(nil)

		for in, out := range chans {
			in <- imd
			imd = <-out
		}
		win.Clear(colornames.Black)
		imd.Draw(win)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
