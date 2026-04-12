package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/models"
	"github.com/leo/synthwaves-cli/internal/player"
	"github.com/leo/synthwaves-cli/internal/ui/components"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type TracksView struct {
	viewSize
	client       *api.Client
	search       components.SearchInput
	loading      bool
	spinner      spinner.Model
	page         int
	query        string
	items        []models.Track
	pagination   models.Pagination
	cursor       int
	detailOffset int
	err          error
}

func NewTracksView(client *api.Client) *TracksView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &TracksView{client: client, search: components.NewSearchInput("Search tracks..."), loading: true, spinner: s, page: 1}
}

type tracksLoadedMsg struct {
	items      []models.Track
	pagination models.Pagination
	err        error
}

func (v *TracksView) Init() tea.Cmd { return tea.Batch(v.spinner.Tick, v.loadTracks()) }
func (v *TracksView) loadTracks() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewTrackService(v.client)
		resp, err := svc.List(context.Background(), api.TrackListParams{Query: v.query, Page: v.page, PerPage: v.perPage()})
		if err != nil {
			return tracksLoadedMsg{err: err}
		}
		return tracksLoadedMsg{items: resp.Items, pagination: resp.Pagination}
	}
}
func (v *TracksView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case tracksLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.items = msg.items
		v.pagination = msg.pagination
		if v.cursor >= len(v.items) {
			v.cursor = len(v.items) - 1
		}
		if v.cursor < 0 {
			v.cursor = 0
		}
		return v, nil
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.loadTracks())
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.loadTracks())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	if updated, handled, cmd := handleFocusedSearchInput(msg, v.search, func(query string) tea.Cmd {
		v.query = query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return tea.Batch(v.spinner.Tick, v.loadTracks())
	}); handled {
		v.search = updated
		return v, cmd
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if handleListDetailNavigation(keyMsg, len(v.items), &v.cursor, &v.detailOffset, v.cHeight) {
			return v, nil
		}
		switch keyMsg.String() {
		case "/":
			v.search = v.search.Focus()
			return v, nil
		case "enter":
			if len(v.items) > 0 {
				tracks := make([]player.Track, len(v.items))
				for i, t := range v.items {
					tracks[i] = player.Track{ID: t.ID, Title: t.Title, Artist: t.Artist.Name, Album: t.Album.Title, Duration: t.Duration}
				}
				return v, func() tea.Msg { return PlayQueueMsg{Tracks: tracks, Index: v.cursor} }
			}
		}
	}
	return v, nil
}
func (v *TracksView) View() string {
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("TRACKS"))
	b.WriteString("\n\n")
	b.WriteString(v.search.View())
	b.WriteString("\n\n")
	if v.loading {
		b.WriteString(v.spinner.View() + theme.MutedStyle.Render(" Loading..."))
	} else if v.err != nil {
		b.WriteString(theme.ErrorStyle.Render("Error: " + v.err.Error()))
	} else {
		listWidth := v.cWidth / 3
		if listWidth < 28 {
			listWidth = 28
		}
		visibleRows := v.cHeight - 8
		if visibleRows < 6 {
			visibleRows = 6
		}
		items := make([]ListItem, 0, len(v.items))
		for i := range v.items {
			track := v.items[i]
			items = append(items, ListItem{Title: track.Title, Subtitle: track.Artist.Name + theme.MutedStyle.Render(" | ") + track.Album.Title, Meta: fmt.Sprintf("%d:%02d", int(track.Duration)/60, int(track.Duration)%60), Selected: i == v.cursor})
		}
		listPane := renderListSection("TRACKS", items, "No tracks found.", v.cursor, visibleRows, listWidth)
		detailPane := v.renderTrackDetail(visibleRows)
		b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	}
	return b.String()
}
func (v *TracksView) renderTrackDetail(height int) string {
	selected := selectedAt(v.items, v.cursor)
	if selected == nil {
		return theme.MutedStyle.Render("Select a track to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Title))
	body.WriteString("\n")
	body.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(selected.Artist.Name))
	body.WriteString(theme.MutedStyle.Render("\nAlbum: "))
	body.WriteString(theme.GhostStyle.Render(selected.Album.Title))
	body.WriteString(theme.MutedStyle.Render("\nDuration: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d:%02d", int(selected.Duration)/60, int(selected.Duration)%60)))
	body.WriteString(theme.MutedStyle.Render("\nFormat: "))
	body.WriteString(theme.GhostStyle.Render(selected.FileFormat))
	if selected.Bitrate > 0 {
		body.WriteString(theme.MutedStyle.Render("\nBitrate: "))
		body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%dk", selected.Bitrate)))
	}
	if selected.Lyrics != "" {
		body.WriteString("\n\n")
		body.WriteString(theme.SubtitleStyle.Render("Lyrics"))
		body.WriteString("\n")
		body.WriteString(selected.Lyrics)
	}
	body.WriteString("\n\nEnter plays the selected track and queues the visible page of results.")
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *TracksView) Title() string         { return "Tracks" }
func (v *TracksView) HasFocusedInput() bool { return v.search.Focused }
func (v *TracksView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.Enter, Keys.NextPage, Keys.PrevPage}
}

type PlaylistsView struct {
	viewSize
	client       *api.Client
	search       components.SearchInput
	loading      bool
	spinner      spinner.Model
	page         int
	query        string
	items        []models.Playlist
	pagination   models.Pagination
	cursor       int
	detailOffset int
	err          error
}

func NewPlaylistsView(client *api.Client) *PlaylistsView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &PlaylistsView{client: client, search: components.NewSearchInput("Search playlists..."), loading: true, spinner: s, page: 1}
}

type playlistsLoadedMsg struct {
	items      []models.Playlist
	pagination models.Pagination
	err        error
}

func (v *PlaylistsView) Init() tea.Cmd { return tea.Batch(v.spinner.Tick, v.load()) }
func (v *PlaylistsView) load() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewPlaylistService(v.client)
		resp, err := svc.List(context.Background(), api.PlaylistListParams{Query: v.query, Page: v.page, PerPage: v.perPage()})
		if err != nil {
			return playlistsLoadedMsg{err: err}
		}
		return playlistsLoadedMsg{items: resp.Items, pagination: resp.Pagination}
	}
}
func (v *PlaylistsView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case playlistsLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.items = msg.items
		v.pagination = msg.pagination
		if v.cursor >= len(v.items) {
			v.cursor = len(v.items) - 1
		}
		if v.cursor < 0 {
			v.cursor = 0
		}
		return v, nil
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.load())
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.load())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	if updated, handled, cmd := handleFocusedSearchInput(msg, v.search, func(query string) tea.Cmd {
		v.query = query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return tea.Batch(v.spinner.Tick, v.load())
	}); handled {
		v.search = updated
		return v, cmd
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if handleListDetailNavigation(keyMsg, len(v.items), &v.cursor, &v.detailOffset, v.cHeight) {
			return v, nil
		}
		switch keyMsg.String() {
		case "/":
			v.search = v.search.Focus()
			return v, nil
		case "enter":
			selected := selectedAt(v.items, v.cursor)
			if selected != nil {
				id := selected.ID
				return v, func() tea.Msg { return NavigateMsg{View: NewPlaylistDetailView(v.client, id)} }
			}
		}
	}
	return v, nil
}
func (v *PlaylistsView) View() string {
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("PLAYLISTS"))
	b.WriteString("\n\n")
	b.WriteString(v.search.View())
	b.WriteString("\n\n")
	if v.loading {
		b.WriteString(v.spinner.View() + theme.MutedStyle.Render(" Loading..."))
	} else if v.err != nil {
		b.WriteString(theme.ErrorStyle.Render("Error: " + v.err.Error()))
	} else {
		listWidth := v.cWidth / 3
		if listWidth < 28 {
			listWidth = 28
		}
		visibleRows := v.cHeight - 8
		if visibleRows < 6 {
			visibleRows = 6
		}
		items := make([]ListItem, 0, len(v.items))
		for i := range v.items {
			playlist := v.items[i]
			items = append(items, ListItem{Title: playlist.Name, Meta: fmt.Sprintf("%d tracks  %s", playlist.TracksCount, friendlyTime(playlist.UpdatedAt)), Selected: i == v.cursor})
		}
		listPane := renderListSection("PLAYLISTS", items, "No playlists found.", v.cursor, visibleRows, listWidth)
		detailPane := v.renderPlaylistSummary(visibleRows)
		b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	}
	return b.String()
}
func (v *PlaylistsView) renderPlaylistSummary(height int) string {
	selected := selectedAt(v.items, v.cursor)
	if selected == nil {
		return theme.MutedStyle.Render("Select a playlist to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Name))
	body.WriteString("\n")
	body.WriteString(theme.MutedStyle.Render(fmt.Sprintf("%d tracks", selected.TracksCount)))
	body.WriteString(theme.MutedStyle.Render("\nUpdated "))
	body.WriteString(theme.GhostStyle.Render(friendlyTime(selected.UpdatedAt)))
	body.WriteString(theme.MutedStyle.Render("\nCreated "))
	body.WriteString(theme.GhostStyle.Render(friendlyTime(selected.CreatedAt)))
	body.WriteString("\n\nEnter opens the playlist workspace and preserves your current list state here.")
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *PlaylistsView) Title() string         { return "Playlists" }
func (v *PlaylistsView) HasFocusedInput() bool { return v.search.Focused }
func (v *PlaylistsView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.Enter, Keys.NextPage, Keys.PrevPage}
}

type FavoritesView struct {
	viewSize
	client       *api.Client
	search       components.SearchInput
	loading      bool
	spinner      spinner.Model
	page         int
	query        string
	favType      string
	items        []models.Favorite
	pagination   models.Pagination
	cursor       int
	detailOffset int
	err          error
}

func NewFavoritesView(client *api.Client) *FavoritesView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &FavoritesView{client: client, search: components.NewSearchInput("Search favorites..."), loading: true, spinner: s, page: 1}
}

type favoritesLoadedMsg struct {
	items      []models.Favorite
	pagination models.Pagination
	err        error
}

func (v *FavoritesView) Init() tea.Cmd { return tea.Batch(v.spinner.Tick, v.load()) }
func (v *FavoritesView) load() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewFavoriteService(v.client)
		resp, err := svc.List(context.Background(), api.FavoriteListParams{Type: v.favType, Query: v.query, Page: v.page, PerPage: v.perPage()})
		if err != nil {
			return favoritesLoadedMsg{err: err}
		}
		return favoritesLoadedMsg{items: resp.Items, pagination: resp.Pagination}
	}
}
func (v *FavoritesView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case favoritesLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.items = msg.items
		v.pagination = msg.pagination
		if v.cursor >= len(v.items) {
			v.cursor = len(v.items) - 1
		}
		if v.cursor < 0 {
			v.cursor = 0
		}
		return v, nil
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.load())
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			v.favType = "Track"
			v.page = 1
			v.loading = true
			v.cursor = 0
			v.detailOffset = 0
			return v, tea.Batch(v.spinner.Tick, v.load())
		case "2":
			v.favType = "Album"
			v.page = 1
			v.loading = true
			v.cursor = 0
			v.detailOffset = 0
			return v, tea.Batch(v.spinner.Tick, v.load())
		case "3":
			v.favType = "Artist"
			v.page = 1
			v.loading = true
			v.cursor = 0
			v.detailOffset = 0
			return v, tea.Batch(v.spinner.Tick, v.load())
		case "0":
			v.favType = ""
			v.page = 1
			v.loading = true
			v.cursor = 0
			v.detailOffset = 0
			return v, tea.Batch(v.spinner.Tick, v.load())
		}
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.load())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	if updated, handled, cmd := handleFocusedSearchInput(msg, v.search, func(query string) tea.Cmd {
		v.query = query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return tea.Batch(v.spinner.Tick, v.load())
	}); handled {
		v.search = updated
		return v, cmd
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if handleListDetailNavigation(keyMsg, len(v.items), &v.cursor, &v.detailOffset, v.cHeight) {
			return v, nil
		}
		switch keyMsg.String() {
		case "/":
			v.search = v.search.Focus()
			return v, nil
		case "enter":
			selected := selectedAt(v.items, v.cursor)
			if selected != nil {
				switch selected.FavorableType {
				case "Track":
					return v, func() tea.Msg {
						tracks, selectedIndex := favoriteTrackQueue(v.items, selected.Favorable.ID)
						return PlayQueueMsg{Tracks: tracks, Index: selectedIndex}
					}
				case "Album":
					return v, func() tea.Msg { return NavigateMsg{View: NewAlbumDetailView(v.client, selected.Favorable.ID)} }
				case "Artist":
					return v, func() tea.Msg { return NavigateMsg{View: NewArtistDetailView(v.client, selected.Favorable.ID)} }
				}
			}
		}
	}
	return v, nil
}
func (v *FavoritesView) View() string {
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("FAVORITES"))
	b.WriteString("  ")
	tabs := []struct{ key, label, val string }{{"0", "All", ""}, {"1", "Tracks", "Track"}, {"2", "Albums", "Album"}, {"3", "Artists", "Artist"}}
	for _, t := range tabs {
		style := theme.MutedStyle
		if t.val == v.favType {
			style = lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true)
		}
		b.WriteString(style.Render("["+t.key+"] "+t.label) + "  ")
	}
	b.WriteString("\n")
	b.WriteString(v.search.View())
	b.WriteString("\n")
	if v.loading {
		b.WriteString(v.spinner.View() + theme.MutedStyle.Render(" Loading..."))
	} else if v.err != nil {
		b.WriteString(theme.ErrorStyle.Render("Error: " + v.err.Error()))
	} else {
		listWidth := v.cWidth / 3
		if listWidth < 28 {
			listWidth = 28
		}
		visibleRows := v.cHeight - 9
		if visibleRows < 6 {
			visibleRows = 6
		}
		items := make([]ListItem, 0, len(v.items))
		for i := range v.items {
			fav := v.items[i]
			artist := fav.Favorable.ArtistName()
			if artist == "" {
				artist = "-"
			}
			items = append(items, ListItem{Title: fav.Favorable.DisplayName(), Subtitle: fav.FavorableType + " | " + artist, Meta: friendlyTime(fav.CreatedAt), Selected: i == v.cursor})
		}
		listPane := renderListSection("FAVORITES", items, "No favorites found.", v.cursor, visibleRows, listWidth)
		detailPane := v.renderFavoriteDetail(visibleRows)
		b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	}
	return b.String()
}
func (v *FavoritesView) renderFavoriteDetail(height int) string {
	selected := selectedAt(v.items, v.cursor)
	if selected == nil {
		return theme.MutedStyle.Render("Select a favorite to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Favorable.DisplayName()))
	body.WriteString("\n")
	body.WriteString(theme.MutedStyle.Render(selected.FavorableType))
	artist := selected.Favorable.ArtistName()
	if artist != "" {
		body.WriteString("\n")
		body.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(artist))
	}
	body.WriteString("\n\n")
	body.WriteString(theme.MutedStyle.Render("Favorited "))
	body.WriteString(theme.GhostStyle.Render(friendlyTime(selected.CreatedAt)))
	body.WriteString("\n\n")
	switch selected.FavorableType {
	case "Track":
		body.WriteString("Enter plays the selected track and queues the track favorites in this list.")
	case "Album":
		body.WriteString("Enter opens the album detail screen.")
	case "Artist":
		body.WriteString("Enter opens the artist detail screen.")
	default:
		body.WriteString("This favorite can be inspected from here.")
	}
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *FavoritesView) Title() string         { return "Favorites" }
func (v *FavoritesView) HasFocusedInput() bool { return v.search.Focused }
func (v *FavoritesView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.Enter, Keys.NextPage, Keys.PrevPage}
}

type HistoryView struct {
	viewSize
	client       *api.Client
	search       components.SearchInput
	loading      bool
	spinner      spinner.Model
	page         int
	query        string
	items        []models.PlayHistory
	pagination   models.Pagination
	cursor       int
	detailOffset int
	err          error
}

func NewHistoryView(client *api.Client) *HistoryView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &HistoryView{client: client, search: components.NewSearchInput("Search history..."), loading: true, spinner: s, page: 1}
}

type historyLoadedMsg struct {
	items      []models.PlayHistory
	pagination models.Pagination
	err        error
}

func (v *HistoryView) Init() tea.Cmd { return tea.Batch(v.spinner.Tick, v.load()) }
func (v *HistoryView) load() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewPlayHistoryService(v.client)
		resp, err := svc.List(context.Background(), v.query, v.page, v.perPage())
		if err != nil {
			return historyLoadedMsg{err: err}
		}
		return historyLoadedMsg{items: resp.Items, pagination: resp.Pagination}
	}
}
func (v *HistoryView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case historyLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.items = msg.items
		v.pagination = msg.pagination
		if v.cursor >= len(v.items) {
			v.cursor = len(v.items) - 1
		}
		if v.cursor < 0 {
			v.cursor = 0
		}
		return v, nil
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.load())
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.load())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	if updated, handled, cmd := handleFocusedSearchInput(msg, v.search, func(query string) tea.Cmd {
		v.query = query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return tea.Batch(v.spinner.Tick, v.load())
	}); handled {
		v.search = updated
		return v, cmd
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if handleListDetailNavigation(keyMsg, len(v.items), &v.cursor, &v.detailOffset, v.cHeight) {
			return v, nil
		}
		switch keyMsg.String() {
		case "/":
			v.search = v.search.Focus()
			return v, nil
		case "pagedown", "ctrl+d":
			return v, nil
		case "pageup", "ctrl+u":
			return v, nil
		}
	}
	return v, nil
}
func (v *HistoryView) View() string {
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("PLAY HISTORY"))
	b.WriteString("\n")
	b.WriteString(v.search.View())
	b.WriteString("\n")
	if v.loading {
		b.WriteString(v.spinner.View() + theme.MutedStyle.Render(" Loading..."))
	} else if v.err != nil {
		b.WriteString(theme.ErrorStyle.Render("Error: " + v.err.Error()))
	} else {
		listWidth := v.cWidth / 3
		if listWidth < 28 {
			listWidth = 28
		}
		visibleRows := v.cHeight - 7
		if visibleRows < 6 {
			visibleRows = 6
		}
		items := make([]ListItem, 0, len(v.items))
		for i := range v.items {
			ph := v.items[i]
			items = append(items, ListItem{Title: ph.Track.Title, Subtitle: ph.Track.Artist.Name, Meta: friendlyTime(ph.PlayedAt), Selected: i == v.cursor})
		}
		listPane := renderListSection("HISTORY", items, "No play history found.", v.cursor, visibleRows, listWidth)
		detailPane := v.renderHistoryDetail(visibleRows)
		b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	}
	return b.String()
}
func (v *HistoryView) renderHistoryDetail(height int) string {
	selected := selectedAt(v.items, v.cursor)
	if selected == nil {
		return theme.MutedStyle.Render("Select a play event to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Track.Title))
	body.WriteString("\n")
	body.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(selected.Track.Artist.Name))
	body.WriteString(theme.MutedStyle.Render("\nAlbum: "))
	body.WriteString(theme.GhostStyle.Render(selected.Track.Album.Title))
	body.WriteString(theme.MutedStyle.Render("\nPlayed: "))
	body.WriteString(theme.GhostStyle.Render(friendlyTime(selected.PlayedAt)))
	body.WriteString(theme.MutedStyle.Render("\nDuration: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d:%02d", int(selected.Track.Duration)/60, int(selected.Track.Duration)%60)))
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *HistoryView) Title() string         { return "History" }
func (v *HistoryView) HasFocusedInput() bool { return v.search.Focused }
func (v *HistoryView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.NextPage, Keys.PrevPage}
}

type PlaylistDetailView struct {
	viewSize
	client       *api.Client
	id           int64
	playlist     *models.PlaylistDetail
	search       components.SearchInput
	loading      bool
	spinner      spinner.Model
	page         int
	query        string
	cursor       int
	detailOffset int
	err          error
}

func NewPlaylistDetailView(client *api.Client, id int64) *PlaylistDetailView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &PlaylistDetailView{client: client, id: id, search: components.NewSearchInput("Search playlist..."), loading: true, spinner: s, page: 1}
}

type playlistDetailMsg struct {
	playlist *models.PlaylistDetail
	err      error
}

func (v *PlaylistDetailView) Init() tea.Cmd { return tea.Batch(v.spinner.Tick, v.load()) }
func (v *PlaylistDetailView) load() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewPlaylistService(v.client)
		pl, err := svc.Get(context.Background(), v.id, v.query, v.page, v.perPage())
		if err != nil {
			return playlistDetailMsg{err: err}
		}
		return playlistDetailMsg{playlist: pl}
	}
}
func (v *PlaylistDetailView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case playlistDetailMsg:
		v.loading = false
		v.playlist = msg.playlist
		v.err = msg.err
		if v.playlist != nil && v.cursor >= len(v.playlist.Tracks) {
			v.cursor = len(v.playlist.Tracks) - 1
		}
		if v.cursor < 0 {
			v.cursor = 0
		}
		return v, nil
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.load())
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.load())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	if updated, handled, cmd := handleFocusedSearchInput(msg, v.search, func(query string) tea.Cmd {
		v.query = query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return tea.Batch(v.spinner.Tick, v.load())
	}); handled {
		v.search = updated
		return v, cmd
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		trackCount := 0
		if v.playlist != nil {
			trackCount = len(v.playlist.Tracks)
		}
		if handleListDetailNavigation(keyMsg, trackCount, &v.cursor, &v.detailOffset, v.cHeight) {
			return v, nil
		}
		switch keyMsg.String() {
		case "/":
			v.search = v.search.Focus()
			return v, nil
		case "enter":
			if v.playlist != nil && len(v.playlist.Tracks) > 0 {
				tracks := make([]player.Track, len(v.playlist.Tracks))
				for i, pt := range v.playlist.Tracks {
					tracks[i] = player.Track{ID: pt.Track.ID, Title: pt.Track.Title, Artist: pt.Track.Artist.Name, Album: pt.Track.Album.Title, Duration: pt.Track.Duration}
				}
				return v, func() tea.Msg { return PlayQueueMsg{Tracks: tracks, Index: v.cursor} }
			}
		case "d":
			selected := v.selectedPlaylistTrack()
			if selected != nil {
				plID := v.playlist.ID
				ptID := selected.PlaylistTrackID
				client := v.client
				query := v.query
				page := v.page
				perPage := v.perPage()
				return v, func() tea.Msg {
					svc := api.NewPlaylistService(client)
					err := svc.RemoveTrack(context.Background(), plID, ptID)
					if err != nil {
						return components.ToastMsg{Message: "Remove failed: " + err.Error(), IsError: true}
					}
					pl, err := svc.Get(context.Background(), plID, query, page, perPage)
					if err != nil {
						return components.ToastMsg{Message: "Track removed (refresh to see changes)"}
					}
					return playlistDetailMsg{playlist: pl}
				}
			}
		}
	}
	return v, nil
}
func (v *PlaylistDetailView) View() string {
	if v.loading {
		return v.spinner.View() + theme.MutedStyle.Render(" Loading playlist...")
	}
	if v.err != nil {
		return theme.ErrorStyle.Render("Error: " + v.err.Error())
	}
	var b strings.Builder
	pl := v.playlist
	b.WriteString(theme.TitleStyle.Render(pl.Name))
	b.WriteString("\n\n")
	dur := fmt.Sprintf("%d:%02d", int(pl.TotalDuration)/60, int(pl.TotalDuration)%60)
	info := fmt.Sprintf("%s tracks  |  %s", lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true).Render(fmt.Sprintf("%d", pl.TracksCount)), theme.MutedStyle.Render(dur))
	b.WriteString(info)
	b.WriteString("\n")
	b.WriteString(v.search.View())
	b.WriteString("\n\n")
	listWidth := v.cWidth / 3
	if listWidth < 28 {
		listWidth = 28
	}
	visibleRows := v.cHeight - 10
	if visibleRows < 6 {
		visibleRows = 6
	}
	items := make([]ListItem, 0, len(v.playlist.Tracks))
	for i := range v.playlist.Tracks {
		pt := v.playlist.Tracks[i]
		items = append(items, ListItem{Title: fmt.Sprintf("%d. %s", pt.Position, pt.Track.Title), Subtitle: pt.Track.Artist.Name + " | " + pt.Track.Album.Title, Meta: fmt.Sprintf("%d:%02d", int(pt.Track.Duration)/60, int(pt.Track.Duration)%60), Selected: i == v.cursor})
	}
	listPane := renderListSection("TRACKS", items, "No playlist tracks found.", v.cursor, visibleRows, listWidth)
	detailPane := v.renderPlaylistTrackDetail(visibleRows)
	b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	return b.String()
}
func (v *PlaylistDetailView) selectedPlaylistTrack() *models.PlaylistTrack {
	if v.playlist == nil {
		return nil
	}
	return selectedAt(v.playlist.Tracks, v.cursor)
}
func (v *PlaylistDetailView) renderPlaylistTrackDetail(height int) string {
	selected := v.selectedPlaylistTrack()
	if selected == nil {
		return theme.MutedStyle.Render("Select a playlist track to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Track.Title))
	body.WriteString("\n")
	body.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(selected.Track.Artist.Name))
	body.WriteString(theme.MutedStyle.Render("\nAlbum: "))
	body.WriteString(theme.GhostStyle.Render(selected.Track.Album.Title))
	body.WriteString(theme.MutedStyle.Render("\nPosition: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d", selected.Position)))
	body.WriteString(theme.MutedStyle.Render("\nDuration: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d:%02d", int(selected.Track.Duration)/60, int(selected.Track.Duration)%60)))
	body.WriteString("\n\nEnter plays the selected track and queues the visible playlist page. Press d to remove it from the playlist.")
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *PlaylistDetailView) Title() string         { return "Playlist" }
func (v *PlaylistDetailView) HasFocusedInput() bool { return v.search.Focused }
func (v *PlaylistDetailView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.Enter, Keys.Delete, Keys.NextPage, Keys.PrevPage, Keys.Back}
}
