# PixelToSVG

Ultra-optimized lossless image to SVG converter.

## Progress Features

- Real-time progress bar with ETA estimation

- Multiple stages: resizing, analysis, block detection, writing

- Performance stats: compression ratio and processing time

- Quiet mode for scripts and batch processing

## Usage

```bash
# Absolute Minimum
pixel2svg input.jpg
```
## With options
```bash
# Auto-generate output name
pixel2svg photo.jpg
```

### Explicit output name
```bash
pixel2svg input.png output.svg
```

### Resize with auto-generated name
```bash
pixel2svg -i large.jpg -w 1920
```

### Multiple files (in scripts)
```bash
pixel2svg -q image1.jpg
pixel2svg -q image2.png
```

## Options

```
-i, --input     Input image file (required)
-o, --output    Output SVG file (required)  
-w, --width     Max width (default: original size)
-h, --height    Max height (default: original size)
-q, --quiet     Quiet mode - disable progress bar
```