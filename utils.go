package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func generateOutputName(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := inputPath[:len(inputPath)-len(ext)]
	return base + ".svg"
}

func getFileSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

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

func calculateSizeReduction(inputSize, outputSize int64) float64 {
	if inputSize == 0 {
		return 0
	}
	return float64(inputSize-outputSize) / float64(inputSize) * 100
}

func reportConversionResults(input, output string, inputSize int64, width, height, blocks int) error {
	outputSize, err := getFileSize(output)
	if err != nil {
		return fmt.Errorf("get output size: %w", err)
	}

	reduction := calculateSizeReduction(inputSize, outputSize)

	if !quiet {
		printConversionSummary(input, output, inputSize, outputSize, reduction, width, height, blocks)
	} else {
		fmt.Printf("Converted: %s (%s) -> %s (%s) - %.1f%% reduction, %d blocks", 
			filepath.Base(input), formatFileSize(inputSize),
			filepath.Base(output), formatFileSize(outputSize),
			reduction, blocks)
	}

	return nil
}

func printConversionSummary(input, output string, inputSize, outputSize int64, reduction float64, width, height, blocks int) {
	fmt.Printf("Conversion complete:\n")
	fmt.Printf("  Input:  %s (%s)\n", filepath.Base(input), formatFileSize(inputSize))
	fmt.Printf("  Output: %s (%s)\n", filepath.Base(output), formatFileSize(outputSize))
	fmt.Printf("  Size reduction: %.1f%%\n", reduction)
	fmt.Printf("  Dimensions: %dx%d pixels\n", width, height)
	fmt.Printf("  Optimization: %d blocks (%.1fx compression)\n", 
		blocks, float64(width*height)/float64(blocks))
}