package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"log"
	"sync"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

//go:embed assets/Icon.svg
var iconSVG []byte

var (
	iconOnce     sync.Once
	iconPNGBytes []byte
)

func getIcon() []byte {
	iconOnce.Do(func() {
		img, err := renderSVG(iconSVG, 64, 64)
		if err != nil {
			log.Printf("[icon] SVG render error: %v, using fallback", err)
			img = generateFallbackIcon()
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			log.Printf("[icon] encode error: %v", err)
			return
		}
		iconPNGBytes = buf.Bytes()
	})
	return iconPNGBytes
}

func renderSVG(data []byte, w, h int) (image.Image, error) {
	icon, err := oksvg.ReadIconStream(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	icon.SetTarget(0, 0, float64(w), float64(h))
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	icon.Draw(raster, 1.0)
	return img, nil
}

func generateFallbackIcon() image.Image {
	size := 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	grass := color.RGBA{52, 211, 153, 255}
	dark := color.RGBA{15, 15, 25, 255}
	stem := color.RGBA{34, 197, 94, 255}
	light := color.RGBA{110, 231, 183, 255}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, dark)
		}
	}

	cx, cy := 16, 28
	drawBlade(img, cx-8, cy, cx-3, 4, grass, stem)
	drawBlade(img, cx-2, cy, cx+2, 2, light, grass)
	drawBlade(img, cx+6, cy, cx+4, 6, grass, stem)
	drawN(img, 12, 14, color.RGBA{255, 255, 255, 230})

	return img
}

func drawBlade(img *image.RGBA, x1, y1, x2, y2 int, c1, c2 color.RGBA) {
	for t := 0.0; t <= 1.0; t += 0.05 {
		x := int((1-t)*(1-t)*float64(x1) + 2*(1-t)*t*float64((x1+x2)/2) + t*t*float64(x2))
		y := int((1-t)*(1-t)*float64(y1) + 2*(1-t)*t*float64(y1-8) + t*t*float64(y2))
		if x >= 0 && x < 32 && y >= 0 && y < 32 {
			r := uint8(float64(c1.R)*(1-t) + float64(c2.R)*t)
			g := uint8(float64(c1.G)*(1-t) + float64(c2.G)*t)
			b := uint8(float64(c1.B)*(1-t) + float64(c2.B)*t)
			img.Set(x, y, color.RGBA{r, g, b, 255})
			img.Set(x+1, y, color.RGBA{r, g, b, 200})
			img.Set(x, y+1, color.RGBA{r, g, b, 200})
		}
	}
}

func drawN(img *image.RGBA, startX, startY int, c color.RGBA) {
	for y := 0; y < 8; y++ {
		img.Set(startX, startY+y, c)
		img.Set(startX+1, startY+y, c)
	}
	for i := 0; i <= 5; i++ {
		img.Set(startX+2+i, startY+i, c)
		img.Set(startX+2+i, startY+i+1, c)
	}
	dx := 7
	for y := 0; y < 8; y++ {
		img.Set(startX+dx, startY+y, c)
		img.Set(startX+dx+1, startY+y, c)
	}
}
