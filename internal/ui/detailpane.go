package ui

import (
	"fmt"
	"strings"

	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type DetailMeta struct {
	Offset      int
	VisibleFrom int
	VisibleTo   int
	Total       int
	HasAbove    bool
	HasBelow    bool
}

func renderScrollableDetail(body string, height, offset int) (string, DetailMeta) {
	lines := strings.Split(body, "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	if height <= 0 {
		height = len(lines)
	}
	maxOffset := len(lines) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	end := offset + height
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[offset:end]

	meta := DetailMeta{
		Offset:      offset,
		VisibleFrom: offset + 1,
		VisibleTo:   end,
		Total:       len(lines),
		HasAbove:    offset > 0,
		HasBelow:    end < len(lines),
	}

	var out []string
	if meta.HasAbove {
		out = append(out, theme.MutedStyle.Render("more above"))
	}
	out = append(out, visible...)
	if meta.HasBelow {
		out = append(out, theme.MutedStyle.Render("more below"))
	}
	if meta.Total > 0 {
		out = append(out, "", theme.MutedStyle.Render(fmt.Sprintf("lines %d-%d of %d", meta.VisibleFrom, meta.VisibleTo, meta.Total)))
	}
	return strings.Join(out, "\n"), meta
}
