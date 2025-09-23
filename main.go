package main

import (
	"flag"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Block struct {
	x, y, w, h int
	r, g, b    uint8
}

func main() {
	var input, output string
	var width, height int

	flag.StringVar(&input, "i", "", "Input image file")
	flag.StringVar(&input, "input", "", "Input image file")
	flag.StringVar(&output, "o", "", "Output SVG file")
	flag.StringVar(&output, "output", "", "Output SVG file")
	flag.IntVar(&width, "w", 0, "Max width (0 = original)")
	flag.IntVar(&width, "width", 0, "Max width (0 = original)")
	flag.IntVar(&height, "h", 0, "Max height (0 = original)")
	flag.IntVar(&height, "height", 0, "Max height (0 = original)")
	flag.Parse()

	// Support positional arguments
	if input == "" && len(flag.Args()) >= 2 {
		input = flag.Arg(0)
		output = flag.Arg(1)
	}

	if input == "" || output == "" {
		log.Fatal("Usage: pixel2svg -i input.jpg -o output.svg\n       pixel2svg input.jpg output.svg")
	}

	if !fileExists(input) {
		log.Fatal("Input file not found:", input)
	}

	img, err := loadImage(input)
	if err != nil {
		log.Fatal("Error loading image:", err)
	}

	// Resize if needed
	if width > 0 || height > 0 {
		img = resizeImage(img, width, height)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	blocks := findOptimalBlocks(img, w, h)
	
	if err := writeSVG(blocks, w, h, output); err != nil {
		log.Fatal("Error writing SVG:", err)
	}

	log.Printf("Converted: %s -> %s (%d blocks from %d pixels)", 
		filepath.Base(input), filepath.Base(output), len(blocks), w*h)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)  // Fixed: ignore the format string
	return img, err
}

func resizeImage(img image.Image, maxW, maxH int) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	if maxW == 0 && maxH == 0 {
		return img
	}

	// Calculate new size maintaining aspect ratio
	newW, newH := calculateSize(w, h, maxW, maxH)
	if newW == w && newH == h {
		return img
	}

	resized := image.NewRGBA(image.Rect(0, 0, newW, newH))
	scaleX, scaleY := float64(w)/float64(newW), float64(h)/float64(newH)

	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX, srcY := int(float64(x)*scaleX), int(float64(y)*scaleY)
			resized.Set(x, y, img.At(srcX, srcY))
		}
	}

	log.Printf("Resized: %dx%d -> %dx%d", w, h, newW, newH)
	return resized
}

func calculateSize(w, h, maxW, maxH int) (int, int) {
	if maxW == 0 && maxH == 0 {
		return w, h
	}
	if maxW == 0 {
		maxW = w * maxH / h
	}
	if maxH == 0 {
		maxH = h * maxW / w
	}

	ratio := float64(w) / float64(h)
	newRatio := float64(maxW) / float64(maxH)

	if ratio > newRatio {
		return maxW, int(float64(maxW) / ratio)
	}
	return int(float64(maxH) * ratio), maxH
}

func findOptimalBlocks(img image.Image, w, h int) []Block {
	grid := make([][]uint32, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]uint32, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			grid[y][x] = (uint32(r>>8) << 16) | (uint32(g>>8) << 8) | uint32(b>>8)
		}
	}

	used := make([][]bool, h)
	for i := range used {
		used[i] = make([]bool, w)
	}

	var blocks []Block

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if used[y][x] {
				continue
			}

			color := grid[y][x]
			maxW := findMaxWidth(grid, x, y, color, w)
			maxH := findMaxHeight(grid, x, y, color, maxW, h)

			// Expand rectangle if possible
			for {
				expanded := false
				
				if x+maxW < w {
					canExpand := true
					for i := y; i < y+maxH; i++ {
						if used[i][x+maxW] || grid[i][x+maxW] != color {
							canExpand = false
							break
						}
					}
					if canExpand {
						maxW++
						expanded = true
					}
				}

				if y+maxH < h {
					canExpand := true
					for i := x; i < x+maxW; i++ {
						if used[y+maxH][i] || grid[y+maxH][i] != color {
							canExpand = false
							break
						}
					}
					if canExpand {
						maxH++
						expanded = true
					}
				}

				if !expanded {
					break
				}
			}

			r := uint8(color >> 16)
			g := uint8(color >> 8)
			b := uint8(color)
			
			blocks = append(blocks, Block{x, y, maxW, maxH, r, g, b})
			
			for i := y; i < y+maxH && i < h; i++ {
				for j := x; j < x+maxW && j < w; j++ {
					used[i][j] = true
				}
			}

			x += maxW - 1
		}
	}

	return blocks
}

func findMaxWidth(grid [][]uint32, x, y int, color uint32, maxX int) int {
	w := 1
	for x+w < maxX && grid[y][x+w] == color {
		w++
	}
	return w
}

func findMaxHeight(grid [][]uint32, x, y int, color uint32, width, maxY int) int {
	h := 1
	for y+h < maxY {
		for i := 0; i < width; i++ {
			if grid[y+h][x+i] != color {
				return h
			}
		}
		h++
	}
	return h
}

func writeSVG(blocks []Block, w, h int, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	file.WriteString(`<svg width="` + strconv.Itoa(w) + `" height="` + strconv.Itoa(h) + `" xmlns="http://www.w3.org/2000/svg">`)

	for _, b := range blocks {
		file.WriteString(`<rect x="` + strconv.Itoa(b.x) + `" y="` + strconv.Itoa(b.y) + 
			`" width="` + strconv.Itoa(b.w) + `" height="` + strconv.Itoa(b.h) + 
			`" fill="rgb(` + strconv.Itoa(int(b.r)) + `,` + strconv.Itoa(int(b.g)) + `,` + strconv.Itoa(int(b.b)) + `)"/>`)
	}

	file.WriteString(`</svg>`)
	return nil
}