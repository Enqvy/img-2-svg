package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"time"
)

type Block struct {
	X, Y, Width, Height int
	R, G, B             uint8
}

func loadAndPrepareImage(input string, width, height int) (image.Image, int64, error) {
	inputSize, err := getFileSize(input)
	if err != nil {
		return nil, 0, fmt.Errorf("get file size: %w", err)
	}

	if !quiet {
		fmt.Printf("Converting %s (%s)...\n", filepath.Base(input), formatFileSize(inputSize))
	}

	img, err := loadAndValidateImage(input)
	if err != nil {
		return nil, 0, fmt.Errorf("load image: %w", err)
	}

	if width > 0 || height > 0 {
		img = resizeImage(img, width, height)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	if err := validateImageDimensions(w, h); err != nil {
		return nil, 0, fmt.Errorf("validate dimensions: %w", err)
	}

	return img, inputSize, nil
}

func loadAndValidateImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	if !quiet {
		fmt.Printf("Detected format: %s\n", format)
	}

	return img, nil
}

func resizeImage(img image.Image, maxW, maxH int) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	if maxW == 0 && maxH == 0 {
		return img
	}

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

	if !quiet {
		fmt.Printf("Resized: %dx%d -> %dx%d\n", w, h, newW, newH)
	}
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

func convertImageToBlocks(img image.Image, width, height int) ([]Block, error) {
	if !quiet {
		fmt.Printf("Analyzing image...\n")
	}

	startTime := time.Now()
	progress := NewProgressTracker(width*height, quiet)

	grid := createColorGrid(img, width, height, progress)
	blocks := findOptimalBlocks(grid, width, height)

	progress.Finish()

	if !quiet {
		duration := time.Since(startTime)
		fmt.Printf("Block detection completed in %v\n", duration.Round(time.Millisecond))
	}

	return blocks, nil
}

func createColorGrid(img image.Image, width, height int, progress *ProgressTracker) [][]uint32 {
	grid := make([][]uint32, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]uint32, width)
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			grid[y][x] = (uint32(r>>8) << 16) | (uint32(g>>8) << 8) | uint32(b>>8)
			progress.Update(1)
		}
	}
	return grid
}

func findOptimalBlocks(grid [][]uint32, width, height int) []Block {
	used := make([][]bool, height)
	for i := range used {
		used[i] = make([]bool, width)
	}

	var blocks []Block

	if !quiet {
		fmt.Printf("Finding optimal blocks...\n")
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if used[y][x] {
				continue
			}

			color := grid[y][x]
			maxWidth := findMaxWidth(grid, x, y, color, width)
			maxHeight := findMaxHeight(grid, x, y, color, maxWidth, height)

			// Expand rectangle if possible
			expandedWidth, expandedHeight := expandBlock(grid, used, x, y, maxWidth, maxHeight, color, width, height)

			r := uint8(color >> 16)
			g := uint8(color >> 8)
			b := uint8(color)

			blocks = append(blocks, Block{
				X:      x,
				Y:      y,
				Width:  expandedWidth,
				Height: expandedHeight,
				R:      r,
				G:      g,
				B:      b,
			})

			markBlockUsed(used, x, y, expandedWidth, expandedHeight)

			x += expandedWidth - 1
		}
	}

	return blocks
}

func expandBlock(grid [][]uint32, used [][]bool, x, y, width, height int, color uint32, maxX, maxY int) (int, int) {
	expandedWidth, expandedHeight := width, height

	for {
		expanded := false

		// Try expand right
		if x+expandedWidth < maxX {
			canExpand := true
			for i := y; i < y+expandedHeight; i++ {
				if used[i][x+expandedWidth] || grid[i][x+expandedWidth] != color {
					canExpand = false
					break
				}
			}
			if canExpand {
				expandedWidth++
				expanded = true
			}
		}

		// Try expand down
		if y+expandedHeight < maxY {
			canExpand := true
			for i := x; i < x+expandedWidth; i++ {
				if used[y+expandedHeight][i] || grid[y+expandedHeight][i] != color {
					canExpand = false
					break
				}
			}
			if canExpand {
				expandedHeight++
				expanded = true
			}
		}

		if !expanded {
			break
		}
	}

	return expandedWidth, expandedHeight
}

func findMaxWidth(grid [][]uint32, x, y int, color uint32, maxX int) int {
	width := 1
	for x+width < maxX && grid[y][x+width] == color {
		width++
	}
	return width
}

func findMaxHeight(grid [][]uint32, x, y int, color uint32, width, maxY int) int {
	height := 1
	for y+height < maxY {
		for i := 0; i < width; i++ {
			if grid[y+height][x+i] != color {
				return height
			}
		}
		height++
	}
	return height
}

func markBlockUsed(used [][]bool, x, y, width, height int) {
	for i := y; i < y+height && i < len(used); i++ {
		for j := x; j < x+width && j < len(used[i]); j++ {
			used[i][j] = true
		}
	}
}
