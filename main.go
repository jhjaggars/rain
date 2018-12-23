package main

import (
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var maxX int
var maxY float64

func wait(in chan *imdraw.IMDraw, out chan *imdraw.IMDraw, jitter int) {
	startTime := time.Now()
	initialPause := time.Duration(rand.Intn(jitter)) * time.Millisecond

	for time.Since(startTime) < initialPause {
		imd := <-in
		out <- imd
	}
}

func drop(in, out chan *imdraw.IMDraw) {
	height := float64(rand.Intn(30) + 10)
	y := maxY - height
	x := rand.Intn(maxX)
	topColor := pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64())
	bottomColor := pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64())
	width := float64(rand.Intn(3) + 1)
	yspeed := float64(rand.Intn(3) + 2)

	wait(in, out, 3000)

	for imd := range in {
		imd.Color = topColor
		imd.Push(pixel.V(float64(x), y))
		imd.Color = bottomColor
		imd.Push(pixel.V(float64(x), y+height))
		imd.Line(width)
		out <- imd

		if y <= 0 {
			x = rand.Intn(maxX)
			y = maxY - height
			yspeed = float64(rand.Intn(3) + 2)
			topColor = pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64())
			bottomColor = pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64())
			width = float64(rand.Intn(3) + 1)
			wait(in, out, 500)
		} else {
			y -= yspeed + ((yspeed + 100) / y)
		}
	}
}

func addDrop(chans map[chan *imdraw.IMDraw]chan *imdraw.IMDraw) {
	in := make(chan *imdraw.IMDraw)
	out := make(chan *imdraw.IMDraw)
	chans[in] = out
	go drop(in, out)
}

func run() {
	monX, monY := pixelgl.PrimaryMonitor().Size()
	maxX = int(monX)
	maxY = monY

	cfg := pixelgl.WindowConfig{
		Title:   "Rain-bow",
		Bounds:  pixel.R(0, 0, float64(maxX), maxY),
		VSync:   true,
		Monitor: pixelgl.PrimaryMonitor(),
	}
	paused := false
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	chans := make(map[chan *imdraw.IMDraw]chan *imdraw.IMDraw)

	addDrop(chans)

	for !win.Closed() {

		if !paused {
			imd := imdraw.New(nil)

			for in, out := range chans {
				in <- imd
				imd = <-out
			}
			win.Clear(colornames.Black)
			imd.Draw(win)
		}

		if win.JustPressed(pixelgl.KeySpace) {
			paused = !paused
		}

		if win.JustPressed(pixelgl.KeyUp) {
			addDrop(chans)
		}

		if win.JustPressed(pixelgl.KeyDown) {
			for k, _ := range chans {
				close(k)
				delete(chans, k)
				break
			}
		}

		if win.JustPressed(pixelgl.KeyR) {
			for k, _ := range chans {
				close(k)
				delete(chans, k)
			}
		}

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
