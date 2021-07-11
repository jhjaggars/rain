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

var maxX float64
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

func randColor(alpha float64) pixel.RGBA {
	return pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64()).Mul(pixel.Alpha(alpha))
}

func roll(min, max int) float64 {
	return float64(rand.Intn(max-min) + min)
}

// Drop represents a single drop on screen
type Drop struct {
	height   float64
	width    float64
	x        float64
	y        float64
	top      pixel.RGBA
	bot      pixel.RGBA
	velocity float64
}

// Fall moves a Drop downards and adjusts velocity
func (drop *Drop) Fall(accel float64) {
	drop.y -= drop.velocity
	drop.velocity *= accel
}

// Draw pushes pixels to the IMDraw object
func (drop *Drop) Draw(imd *imdraw.IMDraw) *imdraw.IMDraw {
	imd.Color = drop.top
	imd.Push(pixel.V(drop.x, drop.y))
	imd.Color = drop.bot
	imd.Push(pixel.V(drop.x, drop.y+drop.height))
	imd.Line(drop.width)
	return imd
}

// NewDrop constructs a new Drop
func NewDrop(maxX, maxY float64) (drop *Drop) {
	drop = &Drop{
		height:   roll(10, 70),
		width:    roll(1, 4),
		x:        float64(rand.Intn(int(maxX))),
		top:      randColor(0.9),
		bot:      randColor(0.1),
		velocity: roll(2, 5),
	}
	drop.y = maxY - drop.height
	return
}

func rundrop(in, out chan *imdraw.IMDraw) {
	drop := NewDrop(maxX, maxY)

	wait(in, out, 500)

	for imd := range in {
		out <- drop.Draw(imd)

		if drop.y <= 0 {
			drop = NewDrop(maxX, maxY)
			wait(in, out, 500)
		} else {
			drop.Fall(accel)
		}
	}
}

func addDrop(chans map[chan *imdraw.IMDraw]chan *imdraw.IMDraw) int {
	in := make(chan *imdraw.IMDraw)
	out := make(chan *imdraw.IMDraw)
	chans[in] = out
	go rundrop(in, out)
	return 1
}

func getTimes(win *pixelgl.Window) int {
	switch {
	case win.Pressed(pixelgl.KeyLeftControl), win.Pressed(pixelgl.KeyRightControl):
		return 100
	case win.Pressed(pixelgl.KeyLeftShift), win.Pressed(pixelgl.KeyRightShift):
		return 10
	default:
		return 1
	}
}

func run() {
	monX, monY := pixelgl.PrimaryMonitor().Size()
	maxX = monX
	maxY = monY
	paused := false
	chans := make(map[chan *imdraw.IMDraw]chan *imdraw.IMDraw)
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	basicText := text.New(pixel.V(0, 0), atlas)
	numDrops := 0
	alpha := 1.0

	cfg := pixelgl.WindowConfig{
		Title:   "Rain-bow",
		Bounds:  pixel.R(0, 0, maxX, maxY),
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
			alpha = math.Max(0.0, alpha-0.004)
			basicText.Color = pixel.Alpha(alpha)
			fmt.Fprintf(basicText, "%d %.2f", numDrops, accel)
			basicText.Draw(win, pixel.IM.Scaled(basicText.Orig, 3))
			imd.Draw(win)
		}

		if win.JustPressed(pixelgl.KeyQ) {
			return
		}

		if win.JustPressed(pixelgl.KeySpace) {
			paused = !paused
		}

		if win.JustPressed(pixelgl.KeyUp) {
			for i := 0; i < getTimes(win); i++ {
				numDrops += addDrop(chans)
			}
			alpha = 1.0
		}

		if win.JustPressed(pixelgl.KeyDown) {
			for i := 0; i < getTimes(win); i++ {
				for k := range chans {
					close(k)
					delete(chans, k)
					break
				}
				numDrops = int(math.Max(0, float64(numDrops-1)))
			}
			alpha = 1.0
		}

		if win.JustPressed(pixelgl.KeyRight) {
			accel += 0.01
			alpha = 1.0
		}

		if win.JustPressed(pixelgl.KeyLeft) {
			accel -= 0.01
			alpha = 1.0
		}

		if win.JustPressed(pixelgl.KeyR) {
			for k := range chans {
				close(k)
				delete(chans, k)
			}
			numDrops = 0
			alpha = 1.0
		}

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
