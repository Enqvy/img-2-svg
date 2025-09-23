package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strconv"
)

type Rect struct {
	x, y, w, h int
	color      string
}

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

	// Create color grid
	grid := make([][]string, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]string, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			grid[y][x] = strconv.Itoa(int(r>>8)) + "," + strconv.Itoa(int(g>>8)) + "," + strconv.Itoa(int(b>>8))
		}
	}

	// Track processed pixels
	used := make([][]bool, h)
	for i := range used {
		used[i] = make([]bool, w)
	}

	var rects []Rect

	// Find optimal rectangles
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if used[y][x] {
				continue
			}

			color := grid[y][x]
			maxW := findWidth(grid, x, y, color, w)
			maxH := findHeight(grid, x, y, color, maxW, h)

			rects = append(rects, Rect{x, y, maxW, maxH, color})

			// Mark as used
			for i := y; i < y+maxH && i < h; i++ {
				for j := x; j < x+maxW && j < w; j++ {
					used[i][j] = true
				}
			}
		}
	}

	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal("err create svg:", err)
	}
	defer out.Close()

	out.WriteString(`<svg width="` + strconv.Itoa(w) + `" height="` + strconv.Itoa(h) + `" xmlns="http://www.w3.org/2000/svg">`)

	for _, r := range rects {
		out.WriteString(`<rect x="` + strconv.Itoa(r.x) + `" y="` + strconv.Itoa(r.y) +
			`" width="` + strconv.Itoa(r.w) + `" height="` + strconv.Itoa(r.h) +
			`" fill="rgb(` + r.color + `)"/>`)
	}

	out.WriteString(`</svg>`)
	log.Printf("done: %s (%d rects, was %d pixels)", os.Args[2], len(rects), w*h)
}

func findWidth(grid [][]string, x, y int, color string, maxX int) int {
	w := 1
	for x+w < maxX && grid[y][x+w] == color {
		w++
	}
	return w
}

func findHeight(grid [][]string, x, y int, color string, width, maxY int) int {
	h := 1
	for y+h < maxY {
		// Check if next row has same color for entire width
		for i := 0; i < width; i++ {
			if grid[y+h][x+i] != color {
				return h
			}
		}
		h++
	}
	return h
}
