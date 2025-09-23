# PixelToSVG

Ultra-optimized lossless image to SVG converter.

## Progress Features

- Real-time progress bar with ETA estimation

- Multiple stages: resizing, analysis, block detection, writing

- Performance stats: compression ratio and processing time

- Quiet mode for scripts and batch processing
## Usage

```bash
pixel2svg input.jpg output.svg
```
# With options
```bash
pixel2svg -i input.jpg -o output.svg -w 1920 -h 1080
```
## Options
-i, --input     Input image file (required)
-o, --output    Output SVG file (required)  
-w, --width     Max width (default: original size)
-h, --height    Max height (default: original size)
-q, --quiet     Quiet mode - disable progress bar