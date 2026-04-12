package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type ListItem struct {
	Title    string
	Subtitle string
	Selected bool
	Meta     string
}

func moveCursor(cursor, total, delta int) int {
	if total <= 0 {
		return 0
	}
	cursor = (cursor + delta) % total
	if cursor < 0 {
		cursor += total
	}
	return cursor
}

func selectedAt[T any](items []T, cursor int) *T {
	if cursor < 0 || cursor >= len(items) {
		return nil
	}
	return &items[cursor]
}

func fittedText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}
	runes := []rune(value)
	if len(runes) > width-3 {
		runes = runes[:width-3]
	}
	return string(runes) + "..."
}

func renderListSection(title string, items []ListItem, empty string, cursor, visibleRows, width int) string {
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render(title))
	b.WriteString("\n")

	if len(items) == 0 {
		b.WriteString("\n")
		b.WriteString(theme.MutedStyle.Render(empty))
		return b.String()
	}

	// Reserve space for the section title and overflow markers so the full pane
	// never exceeds the available viewport height.
	bodyLines := visibleRows - 3
	if bodyLines < 2 {
		bodyLines = 2
	}
	start, end := listWindowByLines(cursor, items, bodyLines)
	if start > 0 {
		b.WriteString(theme.MutedStyle.Render("more above"))
		b.WriteString("\n")
	}

	bodyWidth := width - 6
	if bodyWidth < 12 {
		bodyWidth = 12
	}
	rowWidth := width - 4
	if rowWidth < 14 {
		rowWidth = 14
	}
	selectedStyle := lipgloss.NewStyle().
		Foreground(theme.ColorDeepNavy).
		Background(theme.ColorNeonCyan).
		Bold(true).
		Width(rowWidth)
	rowStyle := lipgloss.NewStyle().Width(rowWidth)
	for i := start; i < end; i++ {
		item := items[i]
		line := fittedText(item.Title, bodyWidth)
		if item.Meta != "" {
			meta := theme.MutedStyle.Render(item.Meta)
			metaWidth := lipgloss.Width(item.Meta)
			titleWidth := bodyWidth - metaWidth - 1
			if titleWidth < 8 {
				titleWidth = 8
			}
			line = fittedText(item.Title, titleWidth) + " " + meta
		}
		if item.Selected {
			b.WriteString(selectedStyle.Render("> " + line))
		} else {
			b.WriteString(rowStyle.Render("  " + theme.GhostStyle.Render(line)))
		}
		b.WriteString("\n")
		if item.Subtitle != "" {
			b.WriteString(rowStyle.Render("  " + theme.MutedStyle.Render(fittedText(item.Subtitle, bodyWidth))))
			b.WriteString("\n")
		}
	}

	if end < len(items) {
		b.WriteString(theme.MutedStyle.Render(fmt.Sprintf("more below (%d hidden)", len(items)-end)))
	}

	return strings.TrimRight(b.String(), "\n")
}

func listWindowByLines(cursor int, items []ListItem, maxLines int) (int, int) {
	total := len(items)
	if total == 0 {
		return 0, 0
	}
	if maxLines <= 0 {
		return 0, total
	}

	lineCost := func(item ListItem) int {
		if item.Subtitle != "" {
			return 2
		}
		return 1
	}

	start := cursor
	used := 0
	for start > 0 {
		next := lineCost(items[start-1])
		if used+next > maxLines/2 {
			break
		}
		start--
		used += next
	}

	end := start
	used = 0
	for end < total {
		next := lineCost(items[end])
		if used+next > maxLines {
			break
		}
		used += next
		end++
	}

	if end == start {
		end = min(total, start+1)
	}

	return start, end
}
