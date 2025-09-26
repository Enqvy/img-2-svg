package main

import (
	"fmt"
	"os"
	"strconv"
)

func generateSVGFile(blocks []Block, width, height int, outputPath string) error {
	if !quiet {
		fmt.Printf("Writing SVG file...\n")
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	writer := NewSVGWriter(file, width, height)
	return writer.WriteBlocks(blocks)
}

type SVGWriter struct {
	file   *os.File
	width  int
	height int
}

func NewSVGWriter(file *os.File, width, height int) *SVGWriter {
	return &SVGWriter{
		file:   file,
		width:  width,
		height: height,
	}
}

func (w *SVGWriter) WriteBlocks(blocks []Block) error {
	if err := w.writeHeader(); err != nil {
		return err
	}

	progress := NewProgressTracker(len(blocks), quiet)
	defer progress.Finish()

	for i, block := range blocks {
		if err := w.writeBlock(block); err != nil {
			return err
		}

		if i%100 == 0 {
			progress.Update(100)
		}
	}

	return w.writeFooter()
}

func (w *SVGWriter) writeHeader() error {
	// Ultra-compact header with minimal whitespace
	header := `<?xml version="1.0" encoding="UTF-8"?><svg width="` + 
		strconv.Itoa(w.width) + `" height="` + strconv.Itoa(w.height) + 
		`" xmlns="http://www.w3.org/2000/svg">`
	_, err := w.file.WriteString(header)
	return err
}

func (w *SVGWriter) writeBlock(block Block) error {
	// Skip fully transparent blocks (alpha = 0)
	if block.A == 0 {
		return nil
	}

	// Optimize color format
	color := w.optimizeColor(block.R, block.G, block.B)
	
	rect := `<rect x="` + strconv.Itoa(block.X) + `" y="` + strconv.Itoa(block.Y) + 
		`" width="` + strconv.Itoa(block.Width) + `" height="` + strconv.Itoa(block.Height) + 
		`" fill="` + color
	
	// Add opacity if not fully opaque
	if block.A < 255 {
		opacity := float64(block.A) / 255.0
		rect += `" fill-opacity="` + strconv.FormatFloat(opacity, 'f', 3, 64)
	}
	
	rect += `"/>`
	_, err := w.file.WriteString(rect)
	return err
}

func (w *SVGWriter) writeFooter() error {
	_, err := w.file.WriteString(`</svg>`)
	return err
}

// optimizeColor converts RGB to shortest possible hex format
func (w *SVGWriter) optimizeColor(r, g, b uint8) string {
	// Use shorthand hex if possible (#abc instead of #aabbcc)
	if r>>4 == r&0x0F && g>>4 == g&0x0F && b>>4 == b&0x0F {
		return "#" + string(hexChar(r>>4)) + string(hexChar(g>>4)) + string(hexChar(b>>4))
	}
	
	// Otherwise use full hex
	return "#" + string(hexChar(r>>4)) + string(hexChar(r&0x0F)) +
		string(hexChar(g>>4)) + string(hexChar(g&0x0F)) +
		string(hexChar(b>>4)) + string(hexChar(b&0x0F))
}

func hexChar(b byte) byte {
	if b < 10 {
		return '0' + b
	}
	return 'a' + (b - 10)
}