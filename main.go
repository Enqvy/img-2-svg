package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Block struct {
	x, y, w, h int
	r, g, b    uint8
}

var quiet bool

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
	flag.BoolVar(&quiet, "q", false, "Quiet mode (no progress bar)")
	flag.BoolVar(&quiet, "quiet", false, "Quiet mode (no progress bar)")
	flag.Parse()

	// Support positional arguments
	if input == "" && len(flag.Args()) >= 1 {
		input = flag.Arg(0)
		if len(flag.Args()) >= 2 {
			output = flag.Arg(1)
		}
	}

	if input == "" {
		log.Fatal("Usage: pixel2svg -i input.jpg -o output.svg\n       pixel2svg input.jpg [output.svg]")
	}

	// Auto-generate output filename if not provided
	if output == "" {
		output = autoGenerateOutputName(input)
		if !quiet {
			fmt.Printf("Output file not specified, using: %s\n", filepath.Base(output))
		}
	}

	if !fileExists(input) {
		log.Fatal("Input file not found:", input)
	}

	startTime := time.Now()
	
	if !quiet {
		fmt.Printf("Converting %s...\n", filepath.Base(input))
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

	// Create progress tracker
	progress := NewProgressTracker(w*h, quiet)
	
	blocks := findOptimalBlocks(img, w, h, progress)
	
	if err := writeSVG(blocks, w, h, output, progress); err != nil {
		log.Fatal("Error writing SVG:", err)
	}

	progress.Finish()
	
	duration := time.Since(startTime)
	
	if !quiet {
		fmt.Printf("Converted: %s -> %s\n", filepath.Base(input), filepath.Base(output))
		fmt.Printf("Stats: %d blocks from %d pixels (%.1fx reduction)\n", 
			len(blocks), w*h, float64(w*h)/float64(len(blocks)))
		fmt.Printf("Time: %v\n", duration.Round(time.Millisecond))
	} else {
		log.Printf("Converted: %s -> %s (%d blocks, %v)", 
			filepath.Base(input), filepath.Base(output), len(blocks), duration)
	}
}

// Auto-generate output filename from input
func autoGenerateOutputName(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := inputPath[:len(inputPath)-len(ext)]
	return base + ".svg"
}

// ProgressTracker handles progress display
type ProgressTracker struct {
	total     int
	processed int
	quiet     bool
	startTime time.Time
	lastUpdate time.Time
}

func NewProgressTracker(total int, quiet bool) *ProgressTracker {
	return &ProgressTracker{
		total:     total,
		quiet:     quiet,
		startTime: time.Now(),
		lastUpdate: time.Now(),
	}
}

func (p *ProgressTracker) Update(increment int) {
	if p.quiet {
		return
	}
	
	p.processed += increment
	
	// Only update progress bar every 100ms to avoid flickering
	if time.Since(p.lastUpdate) < 100*time.Millisecond && p.processed < p.total {
		return
	}
	p.lastUpdate = time.Now()
	
	percent := float64(p.processed) / float64(p.total) * 100
	barWidth := 50
	completed := int(float64(barWidth) * percent / 100)
	
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < completed {
			bar += "="
		} else if i == completed {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"
	
	elapsed := time.Since(p.startTime)
	eta := time.Duration(0)
	if percent > 0 {
		totalEstimate := time.Duration(float64(elapsed) / percent * 100)
		eta = totalEstimate - elapsed
	}
	
	fmt.Printf("\r%s %.1f%% ETA: %v", bar, percent, eta.Round(time.Second))
}

func (p *ProgressTracker) Finish() {
	if p.quiet {
		return
	}
	fmt.Printf("\r[==================================================] 100.0%% ETA: 0s\n")
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
	img, _, err := image.Decode(file)
	return img, err
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

func findOptimalBlocks(img image.Image, w, h int, progress *ProgressTracker) []Block {
	if !quiet {
		fmt.Printf("Analyzing image...\n")
	}

	grid := make([][]uint32, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]uint32, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			grid[y][x] = (uint32(r>>8) << 16) | (uint32(g>>8) << 8) | uint32(b>>8)
			progress.Update(1) // Update progress for each pixel processed
		}
	}

	used := make([][]bool, h)
	for i := range used {
		used[i] = make([]bool, w)
	}

	var blocks []Block

	if !quiet {
		fmt.Printf("Finding optimal blocks...\n")
	}

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

func writeSVG(blocks []Block, w, h int, path string, progress *ProgressTracker) error {
	if !quiet {
		fmt.Printf("Writing SVG file...\n")
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	file.WriteString(`<svg width="` + strconv.Itoa(w) + `" height="` + strconv.Itoa(h) + `" xmlns="http://www.w3.org/2000/svg">`)

	// Update progress during SVG writing (for large numbers of blocks)
	for i, b := range blocks {
		file.WriteString(`<rect x="` + strconv.Itoa(b.x) + `" y="` + strconv.Itoa(b.y) + 
			`" width="` + strconv.Itoa(b.w) + `" height="` + strconv.Itoa(b.h) + 
			`" fill="rgb(` + strconv.Itoa(int(b.r)) + `,` + strconv.Itoa(int(b.g)) + `,` + strconv.Itoa(int(b.b)) + `)"/>`)
		
		// Update progress every 100 blocks to avoid slowing down too much
		if i%100 == 0 {
			progress.Update(0) // 0 means just update the display
		}
	}

	file.WriteString(`</svg>`)
	return nil
}