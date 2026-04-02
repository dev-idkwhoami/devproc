package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func main() {
	colors := map[string]color.RGBA{
		"icons/grey.png":   {R: 128, G: 128, B: 128, A: 255},
		"icons/green.png":  {R: 34, G: 197, B: 94, A: 255},
		"icons/yellow.png": {R: 234, G: 179, B: 8, A: 255},
		"icons/red.png":    {R: 220, G: 38, B: 38, A: 255},
	}

	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	os.MkdirAll("icons", 0755)

	for path, c := range colors {
		img := image.NewRGBA(image.Rect(0, 0, 64, 64))
		cx, cy := 31.5, 31.5
		outerR := 28.0
		borderW := 3.0
		innerR := outerR - borderW

		for y := 0; y < 64; y++ {
			for x := 0; x < 64; x++ {
				dx := float64(x) - cx
				dy := float64(y) - cy
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist <= innerR {
					img.Set(x, y, c)
				} else if dist <= outerR {
					img.Set(x, y, white)
				}
			}
		}
		f, _ := os.Create(path)
		png.Encode(f, img)
		f.Close()
	}
}
