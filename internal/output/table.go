package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/models"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00fff2")).
			PaddingRight(2)

	cellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e8e8f0")).
			PaddingRight(2)

	paginationStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c6c8a")).
			MarginTop(1)
)

type TableFormatter struct{}

func (f *TableFormatter) FormatList(headers []string, rows [][]string, pagination *models.Pagination) string {
	if len(rows) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c8a")).Render("No results found.")
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Header row
	for i, h := range headers {
		sb.WriteString(headerStyle.Width(widths[i] + 2).Render(strings.ToUpper(h)))
	}
	sb.WriteString("\n")

	// Separator
	for i, w := range widths {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#2d1b4e")).Render(strings.Repeat("-", w+2)))
		if i < len(widths)-1 {
			sb.WriteString("")
		}
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				sb.WriteString(cellStyle.Width(widths[i] + 2).Render(cell))
			}
		}
		sb.WriteString("\n")
	}

	// Pagination footer
	if pagination != nil {
		sb.WriteString(paginationStyle.Render(pagination.Summary()))
	}

	return sb.String()
}

func (f *TableFormatter) FormatItem(fields []Field) string {
	maxKey := 0
	for _, field := range fields {
		if len(field.Key) > maxKey {
			maxKey = len(field.Key)
		}
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00fff2")).
		Bold(true).
		Width(maxKey + 1).
		Align(lipgloss.Right)

	valStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e8e8f0")).
		PaddingLeft(1)

	var sb strings.Builder
	for _, field := range fields {
		sb.WriteString(keyStyle.Render(field.Key+":"))
		sb.WriteString(valStyle.Render(field.Value))
		sb.WriteString("\n")
	}
	return sb.String()
}

func (f *TableFormatter) FormatRaw(data []byte) string {
	return fmt.Sprintf("%s", data)
}
