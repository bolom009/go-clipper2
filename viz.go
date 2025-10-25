package go_clipper2

import (
	"fmt"
	"math"
	"strings"
)

const (
	// Default max SVG window size in pixels
	defaultMaxWidth  = 600
	defaultMaxHeight = 600
)

// scaleToFitWindow calculates the scale factor and new dimensions
// to fit content within maxWidth and maxHeight while preserving aspect ratio
func scaleToFitWindow(width, height, maxWidth, maxHeight int) (scale float64, newWidth, newHeight int) {
	if width <= maxWidth && height <= maxHeight {
		// Already fits
		return 1.0, width, height
	}

	scaleX := float64(maxWidth) / float64(width)
	scaleY := float64(maxHeight) / float64(height)

	scale = scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	newWidth = int(float64(width) * scale)
	newHeight = int(float64(height) * scale)

	return scale, newWidth, newHeight
}

// VisualizePath64 creates an SVG representation of a Path64
// Useful for quick debugging and visualization
func VisualizePath64(path Path64, strokeColor string, fillColor string) string {
	if len(path) == 0 {
		return ""
	}

	// Find bounds
	minX, maxX := path[0].X, path[0].X
	minY, maxY := path[0].Y, path[0].Y

	for _, pt := range path {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}

	// Add padding
	padding := 20
	width := int(maxX-minX) + 2*padding
	height := int(maxY-minY) + 2*padding

	// Scale to fit window
	scale, displayWidth, displayHeight := scaleToFitWindow(width, height, defaultMaxWidth, defaultMaxHeight)
	_ = scale // Used in coordinate transformation if needed

	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`+"\n", displayWidth, displayHeight, width, height))
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="white"/>`+"\n", width, height))

	// Draw path
	svg.WriteString(`  <polygon points="`)
	for i, pt := range path {
		x := int(pt.X-minX) + padding
		y := int(pt.Y-minY) + padding
		if i > 0 {
			svg.WriteString(" ")
		}
		svg.WriteString(fmt.Sprintf("%d,%d", x, y))
	}
	svg.WriteString(fmt.Sprintf(`" fill="%s" stroke="%s" stroke-width="1"/>`+"\n", fillColor, strokeColor))

	// Draw points
	for _, pt := range path {
		x := int(pt.X-minX) + padding
		y := int(pt.Y-minY) + padding
		svg.WriteString(fmt.Sprintf(`  <circle cx="%d" cy="%d" r="2" fill="%s"/>`+"\n", x, y, strokeColor))
	}

	svg.WriteString(`</svg>`)
	return svg.String()
}

// VisualizePaths64 creates an SVG representation of multiple Paths64
// Each path gets a different color
func VisualizePaths64(paths Paths64) string {
	if len(paths) == 0 {
		return ""
	}

	// Find bounds
	minX, maxX := int64(math.MaxInt64), int64(math.MinInt64)
	minY, maxY := int64(math.MaxInt64), int64(math.MinInt64)

	for _, path := range paths {
		for _, pt := range path {
			if pt.X < minX {
				minX = pt.X
			}
			if pt.X > maxX {
				maxX = pt.X
			}
			if pt.Y < minY {
				minY = pt.Y
			}
			if pt.Y > maxY {
				maxY = pt.Y
			}
		}
	}

	// Add padding
	padding := 20
	width := int(maxX-minX) + 2*padding
	height := int(maxY-minY) + 2*padding

	// Scale to fit window
	scale, displayWidth, displayHeight := scaleToFitWindow(width, height, defaultMaxWidth, defaultMaxHeight)
	_ = scale // Used in coordinate transformation if needed

	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`+"\n", displayWidth, displayHeight, width, height))
	svg.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="white"/>`+"\n", width, height))

	// Color palette
	colors := []string{
		"#FF6B6B", // red
		"#4ECDC4", // teal
		"#45B7D1", // blue
		"#FFA07A", // salmon
		"#98D8C8", // mint
		"#6C5CE7", // purple
		"#A29BFE", // light purple
		"#74B9FF", // light blue
		"#81ECEC", // light cyan
		"#55EFC4", // light green
	}

	// Draw paths
	for i, path := range paths {
		if len(path) == 0 {
			continue
		}

		color := colors[i%len(colors)]
		svg.WriteString(`  <polygon points="`)
		for j, pt := range path {
			x := int(pt.X-minX) + padding
			y := int(pt.Y-minY) + padding
			if j > 0 {
				svg.WriteString(" ")
			}
			svg.WriteString(fmt.Sprintf("%d,%d", x, y))
		}
		svg.WriteString(fmt.Sprintf(`" fill="%s" opacity="0.6" stroke="%s" stroke-width="1"/>`+"\n", color, color))

		// Draw points
		for _, pt := range path {
			x := int(pt.X-minX) + padding
			y := int(pt.Y-minY) + padding
			svg.WriteString(fmt.Sprintf(`  <circle cx="%d" cy="%d" r="2" fill="%s"/>`+"\n", x, y, color))
		}
	}

	svg.WriteString(`</svg>`)
	return svg.String()
}

// VisualizePaths64ToFile writes SVG visualization to a file
// Returns the SVG content as a string
func VisualizePaths64ToFile(paths Paths64) string {
	return VisualizePaths64(paths)
}

// VisualizePath64HTML creates a complete HTML file for visualization
func VisualizePath64HTML(path Path64, title string) string {
	svg := VisualizePath64(path, "#000000", "#E8F4F8")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 20px; 
            background-color: #f5f5f5;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
        }
        svg { 
            border: 1px solid #ccc; 
            background-color: white;
            max-width: 100%%;
            height: auto;
            display: block;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>%s</h1>
        %s
    </div>
</body>
</html>`, title, title, svg)

	return html
}

// VisualizePaths64HTML creates a complete HTML file for multiple paths
func VisualizePaths64HTML(paths Paths64, title string) string {
	svg := VisualizePaths64(paths)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 20px; 
            background-color: #f5f5f5;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
        }
        svg { 
            border: 1px solid #ccc; 
            background-color: white;
            max-width: 100%%;
            height: auto;
            display: block;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>%s</h1>
        <p>Paths: %d</p>
        %s
    </div>
</body>
</html>`, title, title, len(paths), svg)

	return html
}
