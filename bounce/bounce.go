package main

import (
	"fmt"

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

func (ball *ball) update() {
	ball.x += ball.xv * 0.005
	ball.y += ball.yv * 0.005

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

func setPixel(x, y int, c color, pixels []byte) {
	index := (y*winWidth + x) * 4
	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
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

	ball := ball{pos: pos{300, 300}, radius: 20, xv: 400, yv: 400, color: color{255, 255, 255}}

	// Gameloop
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		// Handle movements
		ball.update()

		// Draw
		clear(pixels)
		ball.draw(pixels)

		texture.Update(nil, pixels, winWidth*4)
		renderer.Copy(texture, nil, nil)
		renderer.Present()

		sdl.Delay(16)
	}
}
