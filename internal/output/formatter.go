package output

import (
	"github.com/leo/synthwaves-cli/internal/models"
)

// Field is a key-value pair for single-item display.
type Field struct {
	Key   string
	Value string
}

// Formatter renders API data for CLI output.
type Formatter interface {
	FormatList(headers []string, rows [][]string, pagination *models.Pagination) string
	FormatItem(fields []Field) string
	FormatRaw(data []byte) string
}

// New returns a formatter for the given format name.
func New(format string) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "text":
		return &TextFormatter{}
	default:
		return &TableFormatter{}
	}
}
