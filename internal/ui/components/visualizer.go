package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/player"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

// Bar characters from lowest to highest amplitude (ASCII only).
var barChars = []byte{' ', '.', ':', '|', 'I', 'H', 'M', 'W', '#'}

// Synthwave gradient colors: pink -> purple -> blue -> cyan, one per band.
var bandColors = []lipgloss.Color{
	"#ff2d95", "#f026a2", "#e11faf", "#d218bc",
	"#c311c9", "#b40ad6", "#a503e3", "#8c0af4",
	"#7310f4", "#5a16f4", "#411cf4", "#2822f4",
	"#1a31f4", "#0d5af4", "#00b0f2", "#00fff2",
}

// RenderVisualizer returns a styled string of spectrum bars from the tap's
// frequency band data. Always returns a visible string when playing.
func RenderVisualizer(tap *player.Tap) string {
	if tap == nil {
		return theme.MutedStyle.Render("[................]")
	}

	bands := tap.Bands()
	out := make([]byte, 0, player.NumBands)

	for i := 0; i < player.NumBands; i++ {
		level := int(bands[i] * float64(len(barChars)-1))
		if level < 0 {
			level = 0
		}
		if level >= len(barChars) {
			level = len(barChars) - 1
		}
		out = append(out, barChars[level])
	}

	result := theme.MutedStyle.Render("[")
	for i, ch := range out {
		color := bandColors[i%len(bandColors)]
		result += lipgloss.NewStyle().Foreground(color).Bold(true).Render(string(ch))
	}
	result += theme.MutedStyle.Render("]")
	return result
}
