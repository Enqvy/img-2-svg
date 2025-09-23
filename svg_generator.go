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
	header := `<?xml version="1.0" encoding="UTF-8"?>` +
		`<svg width="` + strconv.Itoa(w.width) + `" height="` + strconv.Itoa(w.height) + 
		`" xmlns="http://www.w3.org/2000/svg">`
	_, err := w.file.WriteString(header)
	return err
}

func (w *SVGWriter) writeBlock(block Block) error {
	rect := `<rect x="` + strconv.Itoa(block.X) + `" y="` + strconv.Itoa(block.Y) + 
		`" width="` + strconv.Itoa(block.Width) + `" height="` + strconv.Itoa(block.Height) + 
		`" fill="rgb(` + strconv.Itoa(int(block.R)) + `,` + strconv.Itoa(int(block.G)) + 
		`,` + strconv.Itoa(int(block.B)) + `)"/>`
	_, err := w.file.WriteString(rect)
	return err
}

func (w *SVGWriter) writeFooter() error {
	_, err := w.file.WriteString(`</svg>`)
	return err
}