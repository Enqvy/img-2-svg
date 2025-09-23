package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

func validateInputs(input, output string, width, height int) error {
	if err := validateInputFile(input); err != nil {
		return err
	}

	if output != "" {
		if err := validateOutputFile(output, force); err != nil {
			return err
		}
	}

	if width > 0 || height > 0 {
		if err := validateDimensions(width, height); err != nil {
			return err
		}
	}

	return nil
}

func validateInputFile(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("cannot access file: %s - %v", path, err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("file is empty: %s", path)
	}

	if fileInfo.Size() > 500*1024*1024 {
		return fmt.Errorf("file too large (max 500MB): %s (%s)", path, formatFileSize(fileInfo.Size()))
	}

	ext := strings.ToLower(filepath.Ext(path))
	if !supportedFormats[ext] {
		supported := make([]string, 0, len(supportedFormats))
		for format := range supportedFormats {
			supported = append(supported, format)
		}
		return fmt.Errorf("unsupported format: %s (supported: %s)", ext, strings.Join(supported, ", "))
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read file (permission denied?): %s", path)
	}
	file.Close()

	return nil
}

func validateOutputFile(path string, forceOverwrite bool) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("output directory does not exist: %s", dir)
		}
	}

	if _, err := os.Stat(path); err == nil {
		if !forceOverwrite {
			return fmt.Errorf("output file already exists: %s (use --force to overwrite)", path)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot check output file: %s - %v", path, err)
	}

	testFile := filepath.Join(dir, ".pixel2svg_write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("no write permission in directory: %s", dir)
	}
	os.Remove(testFile)

	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".svg" {
		return fmt.Errorf("output file must have .svg extension: %s", path)
	}

	return nil
}

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