// Package colorgen generates a sequence of colors
package colorgen

import "github.com/fatih/color"

// Default is the default color to use when no color is specified
var Default = color.RGB(255, 255, 255)

var colors = []*color.Color{
	color.RGB(255, 100, 100),
	color.RGB(100, 200, 100),
	color.RGB(100, 100, 200),
	color.RGB(200, 200, 0),
	color.RGB(200, 0, 200),
	color.RGB(0, 200, 200),
}

var currentColorIndex int = 0

// Next returns the next color to use
func Next() *color.Color {
	colorIndex := currentColorIndex

	// Next color index
	currentColorIndex++
	if currentColorIndex >= len(colors) {
		currentColorIndex = 0
	}

	return colors[colorIndex]
}
