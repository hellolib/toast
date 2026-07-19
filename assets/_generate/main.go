// Command generate renders the default "terminal" logo used as the Windows
// toast appLogoOverride image when a caller does not supply WithIcon. Run from
// the repo root:
//
//	go run assets/_generate/main.go
//
// It writes assets/toast-icon.png (a 256x256 transparent PNG: a dark rounded
// terminal body with a green ">_" prompt). The directory is underscore-prefixed
// so `go build ./...` ignores it; no build deps beyond the standard library.
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

const (
	size    = 256
	pad     = 22.0
	radius  = 46.0
	stroke  = 22.0
	outPath = "assets/toast-icon.png"
)

var (
	bodyColor  = color.RGBA{0x0F, 0x14, 0x1B, 0xFF}
	glyphColor = color.RGBA{0x4C, 0xE0, 0x8A, 0xFF}
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for i := range img.Pix {
		img.Pix[i] = 0 // transparent
	}

	x0, y0, x1, y1 := pad, pad, float64(size)-pad, float64(size)-pad
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if insideRoundRect(float64(x)+0.5, float64(y)+0.5, x0, y0, x1, y1, radius) {
				img.SetRGBA(x, y, bodyColor)
			}
		}
	}

	const leftX, apexX = 80.0, 128.0
	const topY, midY, botY = 92.0, 134.0, 176.0
	drawThickLine(img, leftX, topY, apexX, midY, stroke, glyphColor)
	drawThickLine(img, apexX, midY, leftX, botY, stroke, glyphColor)
	drawThickLine(img, 150.0, 180.0, 196.0, 180.0, stroke, glyphColor)

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		panic(err)
	}
	f, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func insideRoundRect(px, py, x0, y0, x1, y1, r float64) bool {
	rx0, ry0, rx1, ry1 := x0+r, y0+r, x1-r, y1-r
	dx := math.Max(math.Max(rx0-px, px-rx1), 0)
	dy := math.Max(math.Max(ry0-py, py-ry1), 0)
	return dx*dx+dy*dy <= r*r
}

func drawThickLine(img *image.RGBA, x0, y0, x1, y1, thick float64, c color.RGBA) {
	half := thick / 2
	minX := int(math.Floor(math.Min(x0, x1) - half))
	maxX := int(math.Ceil(math.Max(x0, x1) + half))
	minY := int(math.Floor(math.Min(y0, y1) - half))
	maxY := int(math.Ceil(math.Max(y0, y1) + half))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if x < 0 || y < 0 || x >= size || y >= size {
				continue
			}
			if distToSegment(float64(x)+0.5, float64(y)+0.5, x0, y0, x1, y1) <= half {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func distToSegment(px, py, x0, y0, x1, y1 float64) float64 {
	vx, vy := x1-x0, y1-y0
	wx, wy := px-x0, py-y0
	c1 := vx*wx + vy*wy
	if c1 <= 0 {
		return math.Hypot(px-x0, py-y0)
	}
	c2 := vx*vx + vy*vy
	if c2 <= c1 {
		return math.Hypot(px-x1, py-y1)
	}
	t := c1 / c2
	return math.Hypot(px-(x0+t*vx), py-(y0+t*vy))
}
