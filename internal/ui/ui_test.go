package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leo/synthwaves-cli/internal/ui/components"
)

type refreshTestView struct {
	count int
}

func (v *refreshTestView) Init() tea.Cmd                      { return nil }
func (v *refreshTestView) Update(msg tea.Msg) (View, tea.Cmd) { return v, nil }
func (v *refreshTestView) View() string                       { return "" }
func (v *refreshTestView) Title() string                      { return "test" }
func (v *refreshTestView) ShortHelp() []key.Binding           { return nil }
func (v *refreshTestView) Refresh() tea.Cmd                   { v.count++; return nil }

func TestRenderWorkspaceStacksOnNarrowWidth(t *testing.T) {
	view := RenderWorkspace(WorkspaceProps{Width: 40, Height: 12, Nav: "NAV", Detail: "DETAIL", NavPreferredWidth: 16, NavMinWidth: 16, DetailMinWidth: 20, StackAt: 86})
	lines := strings.Split(view, "\n")
	sharedLine := false
	for _, line := range lines {
		if strings.Contains(line, "NAV") && strings.Contains(line, "DETAIL") {
			sharedLine = true
		}
	}
	if sharedLine {
		t.Fatalf("expected stacked layout on narrow width")
	}
	if strings.Index(view, "NAV") >= strings.Index(view, "DETAIL") {
		t.Fatalf("expected nav before detail in stacked layout")
	}
}

func TestRenderWorkspaceSplitsOnWideWidth(t *testing.T) {
	view := RenderWorkspace(WorkspaceProps{Width: 120, Height: 12, Nav: "NAV", Detail: "DETAIL", NavPreferredWidth: 30, NavMinWidth: 20, DetailMinWidth: 30, StackAt: 86})
	lines := strings.Split(view, "\n")
	sharedLine := false
	for _, line := range lines {
		if strings.Contains(line, "NAV") && strings.Contains(line, "DETAIL") {
			sharedLine = true
			break
		}
	}
	if !sharedLine {
		t.Fatalf("expected split layout on wide width")
	}
}

func TestAppPreservesRouteStateAcrossSidebarSwitches(t *testing.T) {
	app := NewApp(nil)
	model, _ := app.Update(components.NavSelectMsg{Key: string(SectionSearch)})
	app = model.(*App)

	search := app.activeView.(*SearchView)
	search.cursor = 2
	search.detailOffset = 7
	search.items = []searchItem{{title: "a"}, {title: "b"}, {title: "c"}}

	model, _ = app.Update(components.NavSelectMsg{Key: string(SectionAlbums)})
	app = model.(*App)
	model, _ = app.Update(components.NavSelectMsg{Key: string(SectionSearch)})
	app = model.(*App)

	restored, ok := app.activeView.(*SearchView)
	if !ok {
		t.Fatalf("expected search view to be restored")
	}
	if restored != search {
		t.Fatalf("expected cached search view instance to be reused")
	}
	if restored.cursor != 2 || restored.detailOffset != 7 {
		t.Fatalf("expected search route state to be preserved, got cursor=%d detailOffset=%d", restored.cursor, restored.detailOffset)
	}
}

func TestAppRefreshKeyCallsRefreshableView(t *testing.T) {
	app := NewApp(nil)
	view := &refreshTestView{}
	app.activeView = view
	app.sidebarFocused = false

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	app = model.(*App)

	if view.count != 1 {
		t.Fatalf("expected refresh key to call Refresh once, got %d", view.count)
	}
}
