package main

import (
	"fmt"
	"time"

	"github.com/dorant/games-with-go/noise"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 600, 350

type pos struct {
	x, y float32
}

type color struct {
	r, g, b byte
}

type ball struct {
	pos
	radius float32
	xv     float32
	yv     float32
	color  color
}

func (ball ball) draw(pixels []byte) {
	for y := -ball.radius; y < ball.radius; y++ {
		for x := -ball.radius; x < ball.radius; x++ {
			if x*x+y*y < ball.radius*ball.radius {
				setPixel(int(ball.x+x), int(ball.y+y), ball.color, pixels)
			}
		}
	}
}

// update handles any change of the a ball
func (ball *ball) update(elapsedTime float32) {
	ball.x += ball.xv * elapsedTime
	ball.y += ball.yv * elapsedTime

	// Collision detection: Bounce
	if ball.y-ball.radius < 0 || ball.y+ball.radius > float32(winHeight) {
		ball.yv = -ball.yv

		// Corrections of post-collision position: Minimum translation vector
		if ball.y-ball.radius < 0 {
			ball.y = ball.radius
		}
		if ball.y+ball.radius > float32(winHeight) {
			ball.y = float32(winHeight) - ball.radius
		}
	}
	if ball.x-ball.radius < 0 || ball.x+ball.radius > float32(winWidth) {
		ball.xv = -ball.xv

		// Corrections of post-collision position: Minimum translation vector
		if ball.x-ball.radius < 0 {
			ball.x = ball.radius
		}
		if ball.x+ball.radius > float32(winWidth) {
			ball.x = float32(winWidth) - ball.radius
		}
	}

}

// setPixel set the color for a pixel
func setPixel(x, y int, c color, pixels []byte) {
	index := (y*winWidth + x) * 4
	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

// clear cleans the pixelbuffer
func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func lerp(b1 byte, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 color, pct float32) color {
	return color{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func getGradient(c1, c2 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

func rescaleAndDraw(noise []float32, min, max float32, gradient []color, w, h int) []byte {
	result := make([]byte, w*h*4)
	scale := 255.0 / (max - min)
	offset := min * scale
	for i := range noise {
		noise[i] = noise[i]*scale - offset
		c := gradient[clamp(0, 255, int(noise[i]))]
		p := i * 4
		result[p] = c.r
		result[p+1] = c.g
		result[p+2] = c.b
	}
	return result
}

func main() {

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Bounce", sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer texture.Destroy()

	pixels := make([]byte, winWidth*winHeight*4)

	ball := ball{pos: pos{300, 300}, radius: 30, xv: 300, yv: 300, color: color{255, 255, 0}}

	// Create background
	noise, min, max := noise.MakeNoise(noise.TURBULENCE, .02, 0.5, 2, 3, winWidth, winHeight)
	gradient := getGradient(color{255, 0, 50}, color{255, 240, 0})
	noisePixels := rescaleAndDraw(noise, min, max, gradient, winWidth, winHeight)

	// Gameloop
	var frameStart time.Time
	var elapsedTime float32
	for {
		frameStart = time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		// Handle movements
		ball.update(elapsedTime)

		// Draw
		// clear(pixels)
		for i := range noisePixels {
			pixels[i] = noisePixels[i]
		}
		ball.draw(pixels)

		texture.Update(nil, pixels, winWidth*4)
		renderer.Copy(texture, nil, nil)
		renderer.Present()

		// Make sure its about 200fps
		elapsedTime = float32(time.Since(frameStart).Seconds())
		if elapsedTime < .005 {
			sdl.Delay(5 - uint32(elapsedTime/1000.0))
			elapsedTime = float32(time.Since(frameStart).Seconds())
		}
	}
}
