// Package colorgen generates a sequence of colors
package colorgen

import (
	"sync/atomic"

	"github.com/fatih/color"
)

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

var nextColorIndex atomic.Uint64

// Next returns the next color to use. Cycles through the available colors. Safe for concurrent use
func Next() *color.Color {
	index := nextColorIndex.Add(1) - 1
	return colors[index%uint64(len(colors))]
}
