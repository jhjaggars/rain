package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var maxX int
var maxY float64
var accel = 1.03

func wait(in chan *imdraw.IMDraw, out chan *imdraw.IMDraw, jitter int) {
	startTime := time.Now()
	initialPause := time.Duration(rand.Intn(jitter)) * time.Millisecond

	for time.Since(startTime) < initialPause {
		imd := <-in
		out <- imd
	}
}

func drop(in, out chan *imdraw.IMDraw) {
	height := float64(rand.Intn(60) + 10)
	y := maxY - height
	x := rand.Intn(maxX)
	topColor := pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64())
	bottomColor := pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64())
	width := float64(rand.Intn(3) + 1)
	yspeed := float64(rand.Intn(3) + 2)

	wait(in, out, 500)

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
			y -= yspeed
			yspeed *= accel
		}
	}
}

func addDrop(chans map[chan *imdraw.IMDraw]chan *imdraw.IMDraw) int {
	in := make(chan *imdraw.IMDraw)
	out := make(chan *imdraw.IMDraw)
	chans[in] = out
	go drop(in, out)
	return 1
}

func getTimes(win *pixelgl.Window) int {
	if win.Pressed(pixelgl.KeyLeftShift) || win.Pressed(pixelgl.KeyRightShift) {
		return 10
	}
	return 1
}

func run() {
	monX, monY := pixelgl.PrimaryMonitor().Size()
	maxX = int(monX)
	maxY = monY
	paused := false
	chans := make(map[chan *imdraw.IMDraw]chan *imdraw.IMDraw)
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	basicText := text.New(pixel.V(0, 0), atlas)
	numDrops := 0

	cfg := pixelgl.WindowConfig{
		Title:   "Rain-bow",
		Bounds:  pixel.R(0, 0, float64(maxX), maxY),
		VSync:   true,
		Monitor: pixelgl.PrimaryMonitor(),
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	numDrops += addDrop(chans)

	for !win.Closed() {

		if !paused {
			imd := imdraw.New(nil)

			for in, out := range chans {
				in <- imd
				imd = <-out
			}
			win.Clear(colornames.Black)
			basicText.Clear()
			fmt.Fprintf(basicText, "%d %f", numDrops, accel)
			basicText.Draw(win, pixel.IM.Scaled(basicText.Orig, 2))
			imd.Draw(win)
		}

		if win.JustPressed(pixelgl.KeySpace) {
			paused = !paused
		}

		if win.JustPressed(pixelgl.KeyUp) {
			for i := 0; i < getTimes(win); i++ {
				numDrops += addDrop(chans)
			}
		}

		if win.JustPressed(pixelgl.KeyDown) {
			for i := 0; i < getTimes(win); i++ {
				for k, _ := range chans {
					close(k)
					delete(chans, k)
					break
				}
				numDrops = int(math.Max(0, float64(numDrops-1)))
			}
		}

		if win.JustPressed(pixelgl.KeyRight) {
			accel += 0.01
		}

		if win.JustPressed(pixelgl.KeyLeft) {
			accel -= 0.01
		}

		if win.JustPressed(pixelgl.KeyR) {
			for k, _ := range chans {
				close(k)
				delete(chans, k)
			}
			numDrops = 0
		}

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
