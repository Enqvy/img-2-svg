package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"path/filepath"
	"time"
)

type Block struct {
	X, Y, Width, Height int
	R, G, B, A          uint8  // Added Alpha channel
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

	// Check if image has transparency
	if hasTransparency(img) && !quiet {
		fmt.Printf("Image has transparency, optimizing transparent areas...\n")
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

// hasTransparency checks if image contains transparent pixels
func hasTransparency(img image.Image) bool {
	bounds := img.Bounds()
	
	// Quick check: sample some pixels for transparency
	for y := bounds.Min.Y; y < bounds.Max.Y && y < bounds.Min.Y+100; y += 10 {
		for x := bounds.Min.X; x < bounds.Max.X && x < bounds.Min.X+100; x += 10 {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 65535 { // Not fully opaque
				return true
			}
		}
	}
	return false
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

// ... (rest of the existing functions remain the same until createColorGrid)

func createColorGrid(img image.Image, width, height int, progress *ProgressTracker) [][]ColorRGBA {
	grid := make([][]ColorRGBA, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]ColorRGBA, width)
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			grid[y][x] = ColorRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
			progress.Update(1)
		}
	}
	return grid
}

// ColorRGBA represents a color with alpha channel
type ColorRGBA struct {
	R, G, B, A uint8
}

// Pack RGBA into uint64 for comparison
func (c ColorRGBA) ToUint64() uint64 {
	return uint64(c.R)<<24 | uint64(c.G)<<16 | uint64(c.B)<<8 | uint64(c.A)
}

// findOptimalBlocks now uses ColorRGBA
func findOptimalBlocks(grid [][]ColorRGBA, width, height int) []Block {
	used := make([][]bool, height)
	for i := range used {
		used[i] = make([]bool, width)
	}

	var blocks []Block
	transparentBlocks := 0

	if !quiet {
		fmt.Printf("Finding optimal blocks...\n")
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if used[y][x] {
				continue
			}

			color := grid[y][x]
			
			// Skip fully transparent pixels (optimization)
			if color.A == 0 {
				used[y][x] = true
				transparentBlocks++
				continue
			}

			maxWidth := findMaxWidthRGBA(grid, x, y, color, width)
			maxHeight := findMaxHeightRGBA(grid, x, y, color, maxWidth, height)

			expandedWidth, expandedHeight := expandBlockRGBA(grid, used, x, y, maxWidth, maxHeight, color, width, height)

			blocks = append(blocks, Block{
				X:      x,
				Y:      y,
				Width:  expandedWidth,
				Height: expandedHeight,
				R:      color.R,
				G:      color.G,
				B:      color.B,
				A:      color.A,
			})

			markBlockUsed(used, x, y, expandedWidth, expandedHeight)

			x += expandedWidth - 1
		}
	}

	if !quiet && transparentBlocks > 0 {
		fmt.Printf("Optimized: skipped %d transparent blocks\n", transparentBlocks)
	}

	return blocks
}

// Updated functions for RGBA color handling
func findMaxWidthRGBA(grid [][]ColorRGBA, x, y int, color ColorRGBA, maxX int) int {
	width := 1
	for x+width < maxX && grid[y][x+width] == color {
		width++
	}
	return width
}

func findMaxHeightRGBA(grid [][]ColorRGBA, x, y int, color ColorRGBA, width, maxY int) int {
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

func expandBlockRGBA(grid [][]ColorRGBA, used [][]bool, x, y, width, height int, color ColorRGBA, maxX, maxY int) (int, int) {
	expandedWidth, expandedHeight := width, height

	for {
		expanded := false

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
