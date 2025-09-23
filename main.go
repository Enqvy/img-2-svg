
# main.go
```go
package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("use: pixel2svg input.jpg output.svg")
	}

	imgFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("err open file:", err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		log.Fatal("err decode:", err)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Store pixels in 2D array
	pixels := make([][]string, h)
	for y := 0; y < h; y++ {
		pixels[y] = make([]string, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pixels[y][x] = strconv.Itoa(int(r>>8)) + "," + strconv.Itoa(int(g>>8)) + "," + strconv.Itoa(int(b>>8))
		}
	}

	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal("err create svg:", err)
	}
	defer out.Close()

	out.WriteString(`<svg width="` + strconv.Itoa(w) + `" height="` + strconv.Itoa(h) + `" xmlns="http://www.w3.org/2000/svg">`)

	// Merge same-colored pixels into rectangles
	for y := 0; y < h; y++ {
		x := 0
		for x < w {
			color := pixels[y][x]
			width := 1
			
			// Find how many consecutive pixels have same color
			for x+width < w && pixels[y][x+width] == color {
				width++
			}
			
			out.WriteString(`<rect x="` + strconv.Itoa(x) + `" y="` + strconv.Itoa(y) + `" width="` + strconv.Itoa(width) + `" height="1" fill="rgb(` + color + `)"/>`)
			x += width
		}
	}

	out.WriteString(`</svg>`)
	log.Print("done: ", os.Args[2])
}