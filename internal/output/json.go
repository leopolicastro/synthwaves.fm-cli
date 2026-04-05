package output

import (
	"bytes"
	"encoding/json"

	"github.com/leo/synthwaves-cli/internal/models"
)

type JSONFormatter struct{}

func (f *JSONFormatter) FormatList(headers []string, rows [][]string, pagination *models.Pagination) string {
	items := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		item := make(map[string]string)
		for i, h := range headers {
			if i < len(row) {
				item[h] = row[i]
			}
		}
		items = append(items, item)
	}
	b, _ := json.MarshalIndent(items, "", "  ")
	return string(b)
}

func (f *JSONFormatter) FormatItem(fields []Field) string {
	item := make(map[string]string)
	for _, field := range fields {
		item[field.Key] = field.Value
	}
	b, _ := json.MarshalIndent(item, "", "  ")
	return string(b)
}

func (f *JSONFormatter) FormatRaw(data []byte) string {
	var buf bytes.Buffer
	if json.Indent(&buf, data, "", "  ") == nil {
		return buf.String()
	}
	return string(data)
}
