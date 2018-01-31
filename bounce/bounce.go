package main

import (
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/dorant/games-with-go/noise"
	"github.com/dorant/games-with-go/vec3"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight, winDepth int = 800, 600, 600

type rgba struct {
	r, g, b byte
}

type sprite struct {
	tex  *sdl.Texture
	pos  vec3.Vector3
	dir  vec3.Vector3
	w, h int
}

func (sprite *sprite) getScaledSize() (width, height float32) {
	scale := (sprite.pos.Z/float32(winDepth) + 1) / 3
	w := float32(sprite.w) * scale
	h := float32(sprite.h) * scale
	return w, h
}

// update handles any change of the a sprite
func (sprite *sprite) update(elapsedTime float32) {
	p := vec3.Add(sprite.pos, vec3.Mult(sprite.dir, elapsedTime))

	newW, newH := sprite.getScaledSize()

	if (p.X-newW/2) < 0 || (p.X+newW/2) > float32(winWidth) {
		sprite.dir.X = -sprite.dir.X
	}

	if (p.Y-newH/2) < 0 || (p.Y+newH/2) > float32(winHeight) {
		sprite.dir.Y = -sprite.dir.Y
	}

	if p.Z < 0 || p.Z > float32(winDepth) {
		sprite.dir.Z = -sprite.dir.Z
	}

	sprite.pos = vec3.Add(sprite.pos, vec3.Mult(sprite.dir, elapsedTime))

	// Screen boundary corrections
	if sprite.pos.X-newW/2 < 0 {
		sprite.pos.X = newW / 2
	}
	if sprite.pos.X+newW/2 > float32(winWidth) {
		sprite.pos.X = float32(winWidth) - newW/2
	}
	if sprite.pos.Y-newH/2 < 0 {
		sprite.pos.Y = newH / 2
	}
	if sprite.pos.Y+newH/2 > float32(winHeight) {
		sprite.pos.Y = float32(winHeight) - newH/2
	}
}

func (sprite *sprite) draw(renderer *sdl.Renderer) {
	newW, newH := sprite.getScaledSize()
	x := int32(sprite.pos.X - newW/2)
	y := int32(sprite.pos.Y - newH/2)
	rect := &sdl.Rect{X: x, Y: y, W: int32(newW), H: int32(newH)}
	renderer.Copy(sprite.tex, nil, rect)
}

type spriteArray []*sprite

func (sprites spriteArray) Len() int {
	return len(sprites)
}

func (sprites spriteArray) Swap(i, j int) {
	sprites[i], sprites[j] = sprites[j], sprites[i]
}

func (sprites spriteArray) Less(i, j int) bool {
	diff := sprites[i].pos.Z - sprites[j].pos.Z
	return diff < -1
}

func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func loadSprites(renderer *sdl.Renderer, numSprites int) []*sprite {

	files := []string{"owl2.png"}
	spriteTextures := make([]*sdl.Texture, len(files))

	for i, file := range files {
		infile, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer infile.Close()

		img, err := png.Decode(infile)
		if err != nil {
			panic(err)
		}

		w := img.Bounds().Max.X
		h := img.Bounds().Max.Y

		pixels := make([]byte, w*h*4)
		bIndex := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				pixels[bIndex] = byte(r / 256)
				bIndex++
				pixels[bIndex] = byte(g / 256)
				bIndex++
				pixels[bIndex] = byte(b / 256)
				bIndex++
				pixels[bIndex] = byte(a / 256)
				bIndex++
			}
		}
		tex := pixelsToTexture(renderer, pixels, w, h)
		err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
		if err != nil {
			panic(err)
		}
		spriteTextures[i] = tex
	}

	sprites := make([]*sprite, numSprites)
	for i := range sprites {
		tex := spriteTextures[i%len(files)]
		pos := vec3.Vector3{rand.Float32() * float32(winWidth), rand.Float32() * float32(winHeight), rand.Float32() * float32(winDepth)}
		dir := vec3.Vector3{rand.Float32() * 0.1, rand.Float32() * 0.1, rand.Float32() * 0.05}
		_, _, w, h, err := tex.Query()
		if err != nil {
			panic(err)
		}
		sprites[i] = &sprite{tex, pos, dir, int(w), int(h)}
	}
	return sprites
}

// setPixel set the rgba for a pixel
func setPixel(x, y int, c rgba, pixels []byte) {
	index := (y*winWidth + x) * 4
	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func lerp(b1 byte, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 rgba, pct float32) rgba {
	return rgba{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func getGradient(c1, c2 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

func rescaleAndDraw(noise []float32, min, max float32, gradient []rgba, w, h int) []byte {
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

type mouseState struct {
	leftButton  bool
	rightButton bool
	x, y        int
}

func getMouseState() mouseState {
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	leftButton := mouseButtonState & sdl.ButtonLMask()
	rightButton := mouseButtonState & sdl.ButtonRMask()
	var result mouseState
	result.x = int(mouseX)
	result.y = int(mouseY)
	result.leftButton = !(leftButton == 0)
	result.rightButton = !(rightButton == 0)
	return result
}

func main() {
	// Some SDL logs
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)

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

	// Enable nice edges
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer texture.Destroy()

	sprites := loadSprites(renderer, 10)

	// Create background
	noise, min, max := noise.MakeNoise(noise.FBM, .01, 0.5, 3, 3, winWidth, winHeight)
	gradient := getGradient(rgba{50, 150, 250}, rgba{255, 255, 255})
	noisePixels := rescaleAndDraw(noise, min, max, gradient, winWidth, winHeight)
	cloudTexture := pixelsToTexture(renderer, noisePixels, winWidth, winHeight)

	// Handle mouse input
	currentMouseState := getMouseState()
	prevMouseState := currentMouseState

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
		currentMouseState = getMouseState()

		if !currentMouseState.leftButton && prevMouseState.leftButton {
			fmt.Println("Left Click!")
		}

		// Handle movements
		for _, sprite := range sprites {
			sprite.update(elapsedTime)
		}
		sort.Stable(spriteArray(sprites))

		// Draw
		renderer.Copy(cloudTexture, nil, nil)

		for _, sprite := range sprites {
			sprite.draw(renderer)
		}

		renderer.Present()

		// Make sure its about 200fps
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		//fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}

		// Update mouse record
		prevMouseState = currentMouseState
	}
}
