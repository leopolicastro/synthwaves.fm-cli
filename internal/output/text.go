package output

import (
	"fmt"
	"strings"

	"github.com/leo/synthwaves-cli/internal/models"
)

type TextFormatter struct{}

func (f *TextFormatter) FormatList(headers []string, rows [][]string, pagination *models.Pagination) string {
	if len(rows) == 0 {
		return "No results found."
	}
	var sb strings.Builder
	for i, row := range rows {
		if i > 0 {
			sb.WriteString("\n")
		}
		for j, cell := range row {
			if j < len(headers) {
				sb.WriteString(fmt.Sprintf("%s: %s\n", headers[j], cell))
			}
		}
	}
	if pagination != nil {
		sb.WriteString(fmt.Sprintf("\n%s\n", pagination.Summary()))
	}
	return sb.String()
}

func (f *TextFormatter) FormatItem(fields []Field) string {
	var sb strings.Builder
	for _, field := range fields {
		sb.WriteString(fmt.Sprintf("%s: %s\n", field.Key, field.Value))
	}
	return sb.String()
}

func (f *TextFormatter) FormatRaw(data []byte) string {
	return string(data)
}
