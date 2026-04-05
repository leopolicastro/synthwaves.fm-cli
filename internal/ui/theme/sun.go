package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Sun gradient colors (top to bottom): yellow -> orange -> hot pink -> magenta
var sunColors = []lipgloss.Color{
	"#ffd319", // golden yellow
	"#ffac00", // orange
	"#ff6b35", // deep orange
	"#ff2d95", // hot pink
	"#e01b6a", // pink-magenta
	"#c40c5e", // magenta
	"#a80052", // deep magenta
	"#8b0046", // purple-magenta
}

// Grid color
var gridColor = lipgloss.Color("#6a2c91")

// RenderSynthwaveSun returns the iconic retrowave sunset scene.
func RenderSynthwaveSun(width int) string {
	if width < 40 {
		width = 40
	}

	sunWidth := 38
	if width < 50 {
		sunWidth = 28
	}

	var lines []string

	// Sun - semicircle with horizontal stripe gaps
	// Each row is progressively wider, with gaps every few rows
	sunRows := []struct {
		widthPct float64 // percentage of max sun width
		gap      bool    // is this a gap line?
	}{
		{0.35, false},
		{0.50, false},
		{0.60, false},
		{0.68, false},
		{0.75, true},  // gap
		{0.80, false},
		{0.85, false},
		{0.88, true},  // gap
		{0.92, false},
		{0.95, true},  // gap
		{0.97, false},
		{1.00, true},  // gap
		{1.00, false},
		{1.00, false},
	}

	for i, row := range sunRows {
		w := int(float64(sunWidth) * row.widthPct)
		if w%2 != 0 {
			w++
		}

		colorIdx := i * len(sunColors) / len(sunRows)
		if colorIdx >= len(sunColors) {
			colorIdx = len(sunColors) - 1
		}
		color := sunColors[colorIdx]

		pad := (sunWidth - w) / 2
		padding := strings.Repeat(" ", pad)

		if row.gap {
			// Gap line - empty space to create the stripe effect
			lines = append(lines, centerText("", sunWidth, width))
		} else {
			bar := strings.Repeat("\u2588", w) // full block character
			styled := lipgloss.NewStyle().Foreground(color).Render(bar)
			lines = append(lines, centerText(padding+styled, sunWidth, width))
		}
	}

	// Horizon line
	horizonLine := lipgloss.NewStyle().Foreground(ColorHotPink).
		Render(strings.Repeat("\u2500", sunWidth+4))
	lines = append(lines, centerText(horizonLine, sunWidth+4, width))

	// Perspective grid below the sun
	gridStyle := lipgloss.NewStyle().Foreground(gridColor)
	for i := 0; i < 6; i++ {
		gridW := sunWidth + 6 - (i * 2)
		if gridW < 10 {
			gridW = 10
		}

		half := gridW / 2
		leftPart := strings.Repeat("\u2500", half)
		rightPart := strings.Repeat("\u2500", half)
		center := "\u253c" // cross

		// Decrease density of horizontal lines as we go down
		if i%2 == 1 {
			leftPart = strings.Repeat(" ", half)
			rightPart = strings.Repeat(" ", half)
			center = "\u2502" // vertical bar
		}

		gridLine := gridStyle.Render(leftPart + center + rightPart)
		lines = append(lines, centerText(gridLine, gridW+1, width))
	}

	return strings.Join(lines, "\n")
}

// centerText pads text to center it within totalWidth.
func centerText(text string, textWidth, totalWidth int) string {
	if totalWidth <= textWidth {
		return text
	}
	pad := (totalWidth - textWidth) / 2
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + text
}

// RenderSunCompact returns a smaller version for tighter spaces.
func RenderSunCompact() string {
	rows := []struct {
		w     int
		color lipgloss.Color
		gap   bool
	}{
		{8, "#ffd319", false},
		{12, "#ffac00", false},
		{16, "#ff6b35", false},
		{18, "#ff2d95", true},
		{20, "#e01b6a", false},
		{22, "#c40c5e", true},
		{24, "#a80052", false},
		{24, "#8b0046", false},
	}

	var lines []string
	maxW := 24
	for _, r := range rows {
		if r.gap {
			lines = append(lines, strings.Repeat(" ", maxW/2))
			continue
		}
		pad := strings.Repeat(" ", (maxW-r.w)/2)
		bar := strings.Repeat("\u2588", r.w)
		lines = append(lines, pad+lipgloss.NewStyle().Foreground(r.color).Render(bar))
	}

	// Horizon
	lines = append(lines, lipgloss.NewStyle().Foreground(ColorHotPink).
		Render(strings.Repeat("\u2500", maxW+2)))

	return strings.Join(lines, "\n")
}
