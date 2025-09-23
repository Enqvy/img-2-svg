package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	x, y, w, h int
	r, g, b    uint8
}

var quiet bool
var force bool

// Supported image formats
var supportedFormats = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".tiff": true,
	".tif":  true,
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
	flag.BoolVar(&quiet, "q", false, "Quiet mode (no progress bar)")
	flag.BoolVar(&quiet, "quiet", false, "Quiet mode (no progress bar)")
	flag.BoolVar(&force, "f", false, "Force overwrite existing files")
	flag.BoolVar(&force, "force", false, "Force overwrite existing files")
	flag.Parse()

	// Support positional arguments
	if input == "" && len(flag.Args()) >= 1 {
		input = flag.Arg(0)
		if len(flag.Args()) >= 2 {
			output = flag.Arg(1)
		}
	}

	if input == "" {
		printUsage()
		os.Exit(1)
	}

	// Validate input file
	if err := validateInputFile(input); err != nil {
		log.Fatal("Input validation error:", err)
	}

	// Auto-generate output filename if not provided
	if output == "" {
		output = autoGenerateOutputName(input)
		if !quiet {
			fmt.Printf("Output file not specified, using: %s\n", filepath.Base(output))
		}
	}

	// Validate output file and path
	if err := validateOutputFile(output, force); err != nil {
		log.Fatal("Output validation error:", err)
	}

	// Get input file size
	inputSize, err := getFileSize(input)
	if err != nil {
		log.Fatal("Error getting input file size:", err)
	}

	startTime := time.Now()
	
	if !quiet {
		fmt.Printf("Converting %s (%s)...\n", filepath.Base(input), formatFileSize(inputSize))
	}

	// Load and validate image
	img, err := loadAndValidateImage(input)
	if err != nil {
		log.Fatal("Image loading error:", err)
	}

	// Validate and apply resizing
	if width > 0 || height > 0 {
		if err := validateDimensions(width, height); err != nil {
			log.Fatal("Dimension validation error:", err)
		}
		img = resizeImage(img, width, height)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Validate image dimensions
	if err := validateImageDimensions(w, h); err != nil {
		log.Fatal("Image dimension error:", err)
	}

	// Create progress tracker
	progress := NewProgressTracker(w*h, quiet)
	
	blocks := findOptimalBlocks(img, w, h, progress)
	
	if err := writeSVG(blocks, w, h, output, progress); err != nil {
		log.Fatal("Error writing SVG:", err)
	}

	progress.Finish()
	
	// Get output file size
	outputSize, err := getFileSize(output)
	if err != nil {
		log.Fatal("Error getting output file size:", err)
	}

	duration := time.Since(startTime)
	
	if !quiet {
		reduction := calculateSizeReduction(inputSize, outputSize)
		printConversionSummary(input, output, inputSize, outputSize, reduction, w, h, len(blocks), duration)
	} else {
		reduction := calculateSizeReduction(inputSize, outputSize)
		log.Printf("Converted: %s (%s) -> %s (%s) - %.1f%% reduction, %d blocks, %v", 
			filepath.Base(input), formatFileSize(inputSize),
			filepath.Base(output), formatFileSize(outputSize),
			reduction, len(blocks), duration)
	}
}

// Print usage information
func printUsage() {
	fmt.Println("PixelToSVG - Lossless image to SVG converter")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  pixel2svg [flags] input.jpg [output.svg]")
	fmt.Println("  pixel2svg -i input.jpg -o output.svg")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -i, --input    Input image file (required)")
	fmt.Println("  -o, --output   Output SVG file (optional)")
	fmt.Println("  -w, --width    Max width (0 = original)")
	fmt.Println("  -h, --height   Max height (0 = original)")
	fmt.Println("  -q, --quiet    Quiet mode")
	fmt.Println("  -f, --force    Force overwrite existing files")
	fmt.Println()
	fmt.Println("Supported formats: JPG, JPEG, PNG, GIF, BMP, TIFF")
}

// Validate input file exists, is readable, and has supported format
func validateInputFile(path string) error {
	// Check if file exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("cannot access file: %s - %v", path, err)
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Check file size (minimum 1 byte, maximum 500MB)
	if fileInfo.Size() == 0 {
		return fmt.Errorf("file is empty: %s", path)
	}
	if fileInfo.Size() > 500*1024*1024 {
		return fmt.Errorf("file too large (max 500MB): %s (%s)", path, formatFileSize(fileInfo.Size()))
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	if !supportedFormats[ext] {
		supported := make([]string, 0, len(supportedFormats))
		for format := range supportedFormats {
			supported = append(supported, format)
		}
		return fmt.Errorf("unsupported format: %s (supported: %s)", ext, strings.Join(supported, ", "))
	}

	// Check read permissions
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read file (permission denied?): %s", path)
	}
	file.Close()

	return nil
}

// Validate output file path and permissions
func validateOutputFile(path string, forceOverwrite bool) error {
	// Check if output directory exists
	dir := filepath.Dir(path)
	if dir != "." {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("output directory does not exist: %s", dir)
		}
	}

	// Check if output file already exists
	if _, err := os.Stat(path); err == nil {
		if !forceOverwrite {
			return fmt.Errorf("output file already exists: %s (use --force to overwrite)", path)
		}
	} else if !os.IsNotExist(err) {
		// Some other error checking file existence
		return fmt.Errorf("cannot check output file: %s - %v", path, err)
	}

	// Check if we have write permissions in the directory
	testFile := filepath.Join(dir, ".pixel2svg_write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("no write permission in directory: %s", dir)
	}
	os.Remove(testFile)

	// Validate output file extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".svg" {
		return fmt.Errorf("output file must have .svg extension: %s", path)
	}

	return nil
}

// Validate resize dimensions
func validateDimensions(width, height int) error {
	if width < 0 || height < 0 {
		return fmt.Errorf("dimensions cannot be negative: width=%d, height=%d", width, height)
	}
	if width > 100000 || height > 100000 {
		return fmt.Errorf("dimensions too large (max 100000): width=%d, height=%d", width, height)
	}
	if width == 0 && height == 0 {
		return fmt.Errorf("both width and height cannot be zero")
	}
	return nil
}

// Validate image dimensions after loading
func validateImageDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid image dimensions: %dx%d", width, height)
	}
	if width > 30000 || height > 30000 {
		return fmt.Errorf("image too large (max 30000x30000): %dx%d", width, height)
	}
	if width*height > 500000000 {
		return fmt.Errorf("image has too many pixels (max 500 million): %dx%d = %d pixels", 
			width, height, width*height)
	}
	return nil
}

// Load and validate image file
func loadAndValidateImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open image file: %v", err)
	}
	defer file.Close()

	// Try to decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("cannot decode image (corrupted or unsupported format): %v", err)
	}

	if !quiet {
		fmt.Printf("Detected format: %s\n", format)
	}

	return img, nil
}

// Calculate size reduction percentage
func calculateSizeReduction(inputSize, outputSize int64) float64 {
	if inputSize == 0 {
		return 0
	}
	return float64(inputSize-outputSize) / float64(inputSize) * 100
}

// Print conversion summary
func printConversionSummary(input, output string, inputSize, outputSize int64, reduction float64, width, height, blocks int, duration time.Duration) {
	fmt.Printf("Conversion complete:\n")
	fmt.Printf("  Input:  %s (%s)\n", filepath.Base(input), formatFileSize(inputSize))
	fmt.Printf("  Output: %s (%s)\n", filepath.Base(output), formatFileSize(outputSize))
	fmt.Printf("  Size reduction: %.1f%%\n", reduction)
	fmt.Printf("  Dimensions: %dx%d pixels\n", width, height)
	fmt.Printf("  Optimization: %d blocks (%.1fx compression)\n", blocks, float64(width*height)/float64(blocks))
	fmt.Printf("  Time: %v\n", duration.Round(time.Millisecond))
}

// Get file size in bytes
func getFileSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// Format file size to human readable string
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
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
	total      int
	processed  int
	quiet      bool
	startTime  time.Time
	lastUpdate time.Time
}

func NewProgressTracker(total int, quiet bool) *ProgressTracker {
	return &ProgressTracker{
		total:      total,
		quiet:      quiet,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
}

func (p *ProgressTracker) Update(increment int) {
	if p.quiet {
		return
	}
	
	p.processed += increment
	
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
			progress.Update(1)
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

	for i, b := range blocks {
		file.WriteString(`<rect x="` + strconv.Itoa(b.x) + `" y="` + strconv.Itoa(b.y) + 
			`" width="` + strconv.Itoa(b.w) + `" height="` + strconv.Itoa(b.h) + 
			`" fill="rgb(` + strconv.Itoa(int(b.r)) + `,` + strconv.Itoa(int(b.g)) + `,` + strconv.Itoa(int(b.b)) + `)"/>`)
		
		if i%100 == 0 {
			progress.Update(0)
		}
	}

	file.WriteString(`</svg>`)
	return nil
}