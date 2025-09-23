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
		log.Fatal("Usage: imgtosvg input.jpg output.svg")
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	file, err := os.Open(inputPath)
	if err != nil {
		log.Fatal("Error opening image:", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal("Error decoding image:", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatal("Error creating SVG file:", err)
	}
	defer outFile.Close()

	outFile.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	outFile.WriteString(`<svg width="` + strconv.Itoa(width) + `" height="` + strconv.Itoa(height) + `" viewBox="0 0 ` + strconv.Itoa(width) + ` ` + strconv.Itoa(height) + `" xmlns="http://www.w3.org/2000/svg">`)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			outFile.WriteString(`<rect x="` + strconv.Itoa(x) + `" y="` + strconv.Itoa(y) + `" width="1" height="1" fill="rgb(` + strconv.Itoa(int(r>>8)) + `,` + strconv.Itoa(int(g>>8)) + `,` + strconv.Itoa(int(b>>8)) + `)"/>`)
		}
	}

	outFile.WriteString(`</svg>`)
	log.Printf("SVG created: %s (%dx%d pixels)", outputPath, width, height)
}