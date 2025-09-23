package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strconv"
)

type Block struct {
	x, y, w, h int
	r, g, b    uint8
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: pixel2svg input.jpg output.svg")
	}

	img, err := loadImage(os.Args[1])
	if err != nil {
		log.Fatal("err:", err)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	blocks := findOptimalBlocks(img, w, h)

	if err := writeSVG(blocks, w, h, os.Args[2]); err != nil {
		log.Fatal("err write:", err)
	}

	log.Printf("converted: %s (%d blocks from %d pixels)", os.Args[2], len(blocks), w*h)
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return image.Decode(file)
}

func findOptimalBlocks(img image.Image, w, h int) []Block {
	// Single pass with smart merging
	grid := make([][]uint32, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]uint32, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Pack RGB into single uint32 for fast comparison
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

			// Try to expand rectangle if possible
			for {
				expanded := false

				// Try expand right
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

				// Try expand down
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

			// Extract RGB from packed color
			r := uint8(color >> 16)
			g := uint8(color >> 8)
			b := uint8(color)

			blocks = append(blocks, Block{x, y, maxW, maxH, r, g, b})

			// Mark area as used
			for i := y; i < y+maxH && i < h; i++ {
				for j := x; j < x+maxW && j < w; j++ {
					used[i][j] = true
				}
			}

			// Skip ahead since we've processed this entire block
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

	file.WriteString(`<svg width="` + strconv.Itoa(w) + `" height="` + strconv.Itoa(h) + `" xmlns="http://www.w3.org/2000/svg">`)

	for _, b := range blocks {
		file.WriteString(`<rect x="` + strconv.Itoa(b.x) + `" y="` + strconv.Itoa(b.y) +
			`" width="` + strconv.Itoa(b.w) + `" height="` + strconv.Itoa(b.h) +
			`" fill="rgb(` + strconv.Itoa(int(b.r)) + `,` + strconv.Itoa(int(b.g)) + `,` + strconv.Itoa(int(b.b)) + `)"/>`)
	}

	file.WriteString(`</svg>`)
	return nil
}
