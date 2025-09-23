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

	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal("err create svg:", err)
	}
	defer out.Close()

	out.WriteString(`<svg width="` + strconv.Itoa(w) + `" height="` + strconv.Itoa(h) + `" xmlns="http://www.w3.org/2000/svg">`)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			out.WriteString(`<rect x="` + strconv.Itoa(x) + `" y="` + strconv.Itoa(y) + `" width=1 height=1 fill="rgb(` + strconv.Itoa(int(r>>8)) + `,` + strconv.Itoa(int(g>>8)) + `,` + strconv.Itoa(int(b>>8)) + `)"/>`)
		}
	}

	out.WriteString(`</svg>`)
	log.Print("done: ", os.Args[2])
}