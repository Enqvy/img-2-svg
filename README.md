# PixelToSVG

High-performance lossless image → SVG converter with strict validation and block-based optimization.

## Features

- Real-time progress bar with ETA  
- Multi-stage pipeline: resize → analyze → block detection → write  
- Compression stats: ratio, blocks, time, size reduction  
- Quiet mode for batch processing  
- Force overwrite option  
- Input validation: format, size, dimensions  

## Usage

### Minimal
```bash
pixel2svg input.jpg
```

### Auto-generated output name
```bash
pixel2svg photo.png
# produces photo.svg
```

### Explicit output name
```bash
pixel2svg input.png output.svg
```

### Resize on conversion
```bash
pixel2svg -i large.jpg -w 1920
```

### Multiple files in scripts
```bash
pixel2svg -q image1.jpg
pixel2svg -q image2.png
```

## Options
```
-i, --input     Input image file (required if no positional argument)
-o, --output    Output SVG file (optional, auto-generated if omitted)
-w, --width     Max width (0 = original)
-h, --height    Max height (0 = original)
-q, --quiet     Quiet mode (no progress bar)
-f, --force     Overwrite existing output
```

## Supported Formats
JPG, JPEG, PNG, GIF, BMP, TIFF