package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var quiet bool
var force bool

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

	// Run the conversion pipeline
	if err := runConversionPipeline(input, output, width, height); err != nil {
		log.Fatal("Conversion failed:", err)
	}
}

func runConversionPipeline(input, output string, width, height int) error {
	// Phase 1: Validation
	if err := validateInputs(input, output, width, height); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Auto-generate output filename if not provided
	if output == "" {
		output = generateOutputName(input)
		if !quiet {
			fmt.Printf("Output file not specified, using: %s\n", filepath.Base(output))
		}
	}

	// Phase 2: Load and prepare image
	img, inputSize, err := loadAndPrepareImage(input, width, height)
	if err != nil {
		return fmt.Errorf("image preparation failed: %w", err)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Phase 3: Convert to blocks
	blocks, err := convertImageToBlocks(img, w, h)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Phase 4: Generate SVG
	if err := generateSVGFile(blocks, w, h, output); err != nil {
		return fmt.Errorf("SVG generation failed: %w", err)
	}

	// Phase 5: Report results
	return reportConversionResults(input, output, inputSize, w, h, len(blocks))
}

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