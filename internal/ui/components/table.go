package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/models"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type PageChangeMsg struct {
	Page int
}

type RowSelectMsg struct {
	Row table.Row
}

type DataTable struct {
	table      table.Model
	columns    []table.Column
	Pagination models.Pagination
	Height     int
	Width      int
}

func buildStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.ColorGridPurple).
		BorderBottom(true).
		Bold(true).
		Foreground(theme.ColorNeonCyan)

	s.Selected = s.Selected.
		Foreground(theme.ColorDeepNavy).
		Background(theme.ColorNeonCyan).
		Bold(true)

	// NOTE: Do NOT set Foreground on s.Cell. The cell's ANSI reset codes
	// would break the Selected row's background color. Unselected rows
	// use the terminal's default foreground (typically white).

	return s
}

// NewDataTable creates a table that fills the given width and height.
// Column widths are treated as minimum/proportional hints -- extra space
// is distributed to wider columns first (typically the "name/title" column).
func NewDataTable(columns []table.Column, rows []table.Row, pagination models.Pagination, height, width int) DataTable {
	// Distribute available width across columns
	cols := distributeWidth(columns, width)

	s := buildStyles()

	tableHeight := height - 4 // Room for pagination line
	if tableHeight < 5 {
		tableHeight = 5
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
		table.WithStyles(s),
		table.WithWidth(width),
	)

	return DataTable{
		table:      t,
		columns:    columns,
		Pagination: pagination,
		Height:     height,
		Width:      width,
	}
}

// distributeWidth expands column widths to fill the available width.
// Extra space goes to the widest column (typically name/title).
func distributeWidth(columns []table.Column, totalWidth int) []table.Column {
	if len(columns) == 0 || totalWidth <= 0 {
		return columns
	}

	result := make([]table.Column, len(columns))
	copy(result, columns)

	minTotal := 0
	widestIdx := 0
	for i, c := range result {
		minTotal += c.Width
		if c.Width > result[widestIdx].Width {
			widestIdx = i
		}
	}

	extra := totalWidth - minTotal - (len(columns) * 2)
	if extra <= 0 {
		return result
	}

	// Give all extra space to the widest column
	result[widestIdx].Width += extra

	return result
}

// Resize updates the table dimensions (called on WindowSizeMsg).
func (dt DataTable) Resize(width, height int) DataTable {
	dt.Width = width
	dt.Height = height

	cols := distributeWidth(dt.columns, width)
	dt.table.SetColumns(cols)
	dt.table.SetWidth(width)

	dt.table.SetStyles(buildStyles())

	tableHeight := height - 4
	if tableHeight < 5 {
		tableHeight = 5
	}
	dt.table.SetHeight(tableHeight)

	return dt
}

func (dt DataTable) Update(msg tea.Msg) (DataTable, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "right", "l":
			if dt.Pagination.HasNext() {
				return dt, func() tea.Msg {
					return PageChangeMsg{Page: dt.Pagination.Page + 1}
				}
			}
		case "left", "h":
			if dt.Pagination.HasPrev() {
				return dt, func() tea.Msg {
					return PageChangeMsg{Page: dt.Pagination.Page - 1}
				}
			}
		case "enter":
			selected := dt.table.SelectedRow()
			if selected != nil {
				return dt, func() tea.Msg {
					return RowSelectMsg{Row: selected}
				}
			}
		}
	}

	var cmd tea.Cmd
	dt.table, cmd = dt.table.Update(msg)
	return dt, cmd
}

func (dt DataTable) View() string {
	tableView := dt.table.View()

	pag := ""
	if dt.Pagination.TotalPages > 0 {
		pag = theme.MutedStyle.Render(
			fmt.Sprintf("  Page %d of %d (%d total)  ",
				dt.Pagination.Page, dt.Pagination.TotalPages, dt.Pagination.TotalCount,
			),
		)
		if dt.Pagination.HasPrev() {
			pag = lipgloss.NewStyle().Foreground(theme.ColorHotPink).Render("<- ") + pag
		}
		if dt.Pagination.HasNext() {
			pag += lipgloss.NewStyle().Foreground(theme.ColorHotPink).Render(" ->")
		}
	}

	return tableView + "\n" + pag
}

func (dt DataTable) SelectedRow() table.Row {
	return dt.table.SelectedRow()
}

func (dt DataTable) SetRows(rows []table.Row) DataTable {
	dt.table.SetRows(rows)
	return dt
}

func (dt DataTable) Focus() DataTable {
	dt.table.Focus()
	return dt
}

func (dt DataTable) Blur() DataTable {
	dt.table.Blur()
	return dt
}
