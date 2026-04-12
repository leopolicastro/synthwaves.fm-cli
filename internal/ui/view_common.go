package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leo/synthwaves-cli/internal/models"
	"github.com/leo/synthwaves-cli/internal/player"
	"github.com/leo/synthwaves-cli/internal/ui/components"
)

const sidebarWidth = 26

func contentSize(termWidth, termHeight int) (width, height int) {
	width = termWidth - sidebarWidth - 4
	if width < 40 {
		width = 40
	}
	height = termHeight - 6
	if height < 10 {
		height = 10
	}
	return
}

func friendlyTime(iso string) string {
	t, err := time.Parse(time.RFC3339Nano, iso)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z", iso)
		if err != nil {
			if len(iso) > 10 {
				return iso[:10]
			}
			return iso
		}
	}
	now := time.Now()
	diff := now.Sub(t)
	switch {
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	default:
		return t.Format("Jan 2, 2006")
	}
}

func favoriteTrackQueue(items []models.Favorite, selectedID int64) ([]player.Track, int) {
	tracks := make([]player.Track, 0, len(items))
	selectedIndex := 0
	for _, fav := range items {
		if fav.FavorableType != "Track" {
			continue
		}
		if fav.Favorable.ID == selectedID {
			selectedIndex = len(tracks)
		}
		tracks = append(tracks, player.Track{
			ID:     fav.Favorable.ID,
			Title:  fav.Favorable.DisplayName(),
			Artist: fav.Favorable.ArtistName(),
		})
	}
	return tracks, selectedIndex
}

func searchTrackQueue(items []searchItem, selectedID int64) ([]player.Track, int) {
	tracks := make([]player.Track, 0, len(items))
	selectedIndex := 0
	for _, item := range items {
		if item.kind != "Track" {
			continue
		}
		if item.id == selectedID {
			selectedIndex = len(tracks)
		}
		tracks = append(tracks, player.Track{ID: item.id, Title: item.title, Artist: item.artist, Album: item.album, Duration: item.duration})
	}
	return tracks, selectedIndex
}

type viewSize struct {
	cWidth  int
	cHeight int
}

func (s *viewSize) handleResize(msg tea.WindowSizeMsg) {
	s.cWidth, s.cHeight = contentSize(msg.Width, msg.Height)
}

func (s viewSize) perPage() int {
	h := s.tHeight() - 2
	if h < 20 {
		return 20
	}
	if h > 100 {
		return 100
	}
	return h
}

func (s viewSize) tWidth() int {
	if s.cWidth < 40 {
		return 80
	}
	return s.cWidth
}

func (s viewSize) tHeight() int {
	h := s.cHeight - 6
	if h < 5 {
		return 20
	}
	return h
}

type dataLoadedMsg[T any] struct {
	items      []T
	pagination models.Pagination
}

type detailLoadedMsg[T any] struct {
	item T
}

type errorMsg struct {
	err error
}

func handleFocusedSearchInput(msg tea.Msg, search components.SearchInput, submit func(string) tea.Cmd) (components.SearchInput, bool, tea.Cmd) {
	if !search.Focused {
		return search, false, nil
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" {
			search = search.Blur()
			return search, true, nil
		}
		if keyMsg.String() == "enter" {
			query := search.Value()
			search = search.Blur()
			return search, true, submit(query)
		}
	}
	var cmd tea.Cmd
	search, cmd = search.Update(msg)
	return search, true, cmd
}

func handleListDetailNavigation(msg tea.KeyMsg, total int, cursor *int, detailOffset *int, pageHeight int) bool {
	switch msg.String() {
	case "j", "down":
		if total > 0 {
			*cursor = moveCursor(*cursor, total, 1)
			*detailOffset = 0
		}
		return true
	case "k", "up":
		if total > 0 {
			*cursor = moveCursor(*cursor, total, -1)
			*detailOffset = 0
		}
		return true
	case "pagedown", "ctrl+d":
		*detailOffset += max(1, pageHeight/3)
		return true
	case "pageup", "ctrl+u":
		*detailOffset -= max(1, pageHeight/3)
		if *detailOffset < 0 {
			*detailOffset = 0
		}
		return true
	default:
		return false
	}
}
