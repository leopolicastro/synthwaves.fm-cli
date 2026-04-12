package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

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

type RadioView struct {
	viewSize
	client       *api.Client
	loading      bool
	spinner      spinner.Model
	stations     []models.RadioStation
	cursor       int
	detailOffset int
	err          error
}

func NewRadioView(client *api.Client) *RadioView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &RadioView{client: client, loading: true, spinner: s}
}

type radioLoadedMsg struct {
	stations []models.RadioStation
	err      error
}

func (v *RadioView) Init() tea.Cmd { return tea.Batch(v.spinner.Tick, v.load()) }
func (v *RadioView) load() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewRadioStationService(v.client)
		stations, err := svc.List(context.Background())
		return radioLoadedMsg{stations: stations, err: err}
	}
}
func (v *RadioView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case radioLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.stations = msg.stations
		if v.cursor >= len(v.stations) {
			v.cursor = len(v.stations) - 1
		}
		if v.cursor < 0 {
			v.cursor = 0
		}
		return v, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if handleListDetailNavigation(keyMsg, len(v.stations), &v.cursor, &v.detailOffset, v.cHeight) {
			return v, nil
		}
		switch keyMsg.String() {
		}
	}
	return v, nil
}
func (v *RadioView) View() string {
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("RADIO STATIONS"))
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
		items := make([]ListItem, 0, len(v.stations))
		for i := range v.stations {
			station := v.stations[i]
			nowPlaying := "-"
			if station.CurrentTrack != nil {
				nowPlaying = station.CurrentTrack.Title
			}
			items = append(items, ListItem{Title: station.Name, Subtitle: station.PlaybackMode + " | " + nowPlaying, Meta: station.Status, Selected: i == v.cursor})
		}
		listPane := renderListSection("RADIO STATIONS", items, "No radio stations found.", v.cursor, visibleRows, listWidth)
		detailPane := v.renderRadioDetail(visibleRows)
		b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	}
	return b.String()
}
func (v *RadioView) renderRadioDetail(height int) string {
	selected := selectedAt(v.stations, v.cursor)
	if selected == nil {
		return theme.MutedStyle.Render("Select a station to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Name))
	body.WriteString("\n")
	body.WriteString(theme.GhostStyle.Render(selected.Status))
	body.WriteString(theme.MutedStyle.Render("\nMode: "))
	body.WriteString(theme.GhostStyle.Render(selected.PlaybackMode))
	body.WriteString(theme.MutedStyle.Render("\nBitrate: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%dk", selected.Bitrate)))
	body.WriteString(theme.MutedStyle.Render("\nMount: "))
	body.WriteString(theme.GhostStyle.Render(selected.MountPoint))
	if selected.CurrentTrack != nil {
		body.WriteString(theme.MutedStyle.Render("\nNow Playing: "))
		body.WriteString(theme.GhostStyle.Render(selected.CurrentTrack.Title))
	}
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *RadioView) Title() string            { return "Radio" }
func (v *RadioView) ShortHelp() []key.Binding { return []key.Binding{Keys.Enter} }

type searchItem struct {
	kind     string
	id       int64
	line     string
	title    string
	artist   string
	album    string
	duration float64
}
type SearchView struct {
	viewSize
	client       *api.Client
	search       components.SearchInput
	loading      bool
	spinner      spinner.Model
	result       *models.SearchResult
	items        []searchItem
	cursor       int
	detailOffset int
	err          error
}

func NewSearchView(client *api.Client) *SearchView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	si := components.NewSearchInput("Search artists, albums, tracks...")
	si = si.Focus()
	return &SearchView{client: client, search: si, spinner: s}
}

type searchResultMsg struct {
	result *models.SearchResult
	err    error
}

func (v *SearchView) Init() tea.Cmd { return nil }
func (v *SearchView) buildItems() {
	v.items = nil
	v.cursor = 0
	v.detailOffset = 0
	if v.result == nil {
		return
	}
	for _, a := range v.result.Artists {
		v.items = append(v.items, searchItem{kind: "Artist", id: a.ID, title: a.Name})
	}
	for _, a := range v.result.Albums {
		v.items = append(v.items, searchItem{kind: "Album", id: a.ID, title: a.Title, artist: a.Artist.Name})
	}
	for _, t := range v.result.Tracks {
		v.items = append(v.items, searchItem{kind: "Track", id: t.ID, title: t.Title, artist: t.Artist.Name, album: t.Album.Title, duration: t.Duration})
	}
}
func (v *SearchView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case components.SearchQueryMsg:
		if msg.Query == "" {
			v.result = nil
			v.items = nil
			v.cursor = 0
			v.detailOffset = 0
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, func() tea.Msg {
			svc := api.NewSearchService(v.client)
			r, err := svc.Search(context.Background(), api.SearchParams{Query: msg.Query})
			return searchResultMsg{result: r, err: err}
		})
	case searchResultMsg:
		v.loading = false
		v.result = msg.result
		v.err = msg.err
		v.buildItems()
		if len(v.items) > 0 {
			v.search = v.search.Blur()
		}
		return v, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	if updated, handled, cmd := handleFocusedSearchInput(msg, v.search, func(query string) tea.Cmd {
		if query == "" {
			v.result = nil
			v.items = nil
			v.cursor = 0
			v.detailOffset = 0
			return nil
		}
		v.loading = true
		return tea.Batch(v.spinner.Tick, func() tea.Msg {
			svc := api.NewSearchService(v.client)
			r, err := svc.Search(context.Background(), api.SearchParams{Query: query})
			return searchResultMsg{result: r, err: err}
		})
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
			if v.cursor < len(v.items) {
				item := v.items[v.cursor]
				switch item.kind {
				case "Track":
					return v, func() tea.Msg {
						tracks, selectedIndex := searchTrackQueue(v.items, item.id)
						return PlayQueueMsg{Tracks: tracks, Index: selectedIndex}
					}
				case "Album":
					return v, func() tea.Msg { return NavigateMsg{View: NewAlbumDetailView(v.client, item.id)} }
				case "Artist":
					return v, func() tea.Msg { return NavigateMsg{View: NewArtistDetailView(v.client, item.id)} }
				}
			}
		}
	}
	return v, nil
}
func (v *SearchView) View() string {
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("SEARCH"))
	b.WriteString("\n")
	b.WriteString(v.search.View())
	b.WriteString("\n\n")
	if v.loading {
		b.WriteString(v.spinner.View() + theme.MutedStyle.Render(" Searching..."))
		return b.String()
	}
	if v.err != nil {
		b.WriteString(theme.ErrorStyle.Render("Error: " + v.err.Error()))
		return b.String()
	}
	if v.result == nil {
		b.WriteString(theme.MutedStyle.Render("Type to search across your library."))
		return b.String()
	}
	listWidth := v.cWidth / 3
	if listWidth < 28 {
		listWidth = 28
	}
	detailWidth := v.cWidth - listWidth - 2
	if detailWidth < 34 {
		detailWidth = 34
	}
	visibleRows := v.cHeight - 8
	if visibleRows < 6 {
		visibleRows = 6
	}
	items := make([]ListItem, 0, len(v.items))
	for i := range v.items {
		item := v.items[i]
		title := item.title
		subtitle := item.kind
		meta := ""
		switch item.kind {
		case "Album":
			subtitle = item.artist
		case "Track":
			subtitle = item.artist
			meta = fmt.Sprintf("%d:%02d", int(item.duration)/60, int(item.duration)%60)
		}
		items = append(items, ListItem{Title: title, Subtitle: subtitle, Meta: meta, Selected: i == v.cursor})
	}
	listPane := renderListSection("RESULTS", items, "No results found.", v.cursor, visibleRows, listWidth)
	detailPane := v.renderSearchDetail(detailWidth, visibleRows)
	b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 34, StackAt: 92}))
	if !v.search.Focused {
		b.WriteString("\n\n")
		b.WriteString(theme.MutedStyle.Render("j/k:move  enter:open or play  /:search  ctrl+u/d:scroll detail"))
	}
	return b.String()
}
func (v *SearchView) renderSearchDetail(width, height int) string {
	selected := selectedAt(v.items, v.cursor)
	if selected == nil {
		return theme.MutedStyle.Render("Select a result to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.title))
	body.WriteString("\n")
	body.WriteString(theme.MutedStyle.Render(selected.kind))
	body.WriteString("\n\n")
	switch selected.kind {
	case "Artist":
		body.WriteString("Open artist details with enter.")
	case "Album":
		body.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(selected.artist))
		body.WriteString("\n\nOpen album details with enter.")
	case "Track":
		body.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(selected.artist))
		if selected.album != "" {
			body.WriteString(theme.MutedStyle.Render("\nAlbum: "))
			body.WriteString(theme.GhostStyle.Render(selected.album))
		}
		body.WriteString(theme.MutedStyle.Render("\nDuration: "))
		body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d:%02d", int(selected.duration)/60, int(selected.duration)%60)))
		body.WriteString("\n\nEnter plays the selected track and queues the track results in this search.")
	}
	body.WriteString("\n\n")
	body.WriteString(theme.SubtitleStyle.Render("Result Breakdown"))
	body.WriteString("\n")
	body.WriteString(fmt.Sprintf("Artists: %d\nAlbums: %d\nTracks: %d", len(v.result.Artists), len(v.result.Albums), len(v.result.Tracks)))
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *SearchView) Title() string            { return "Search" }
func (v *SearchView) HasFocusedInput() bool    { return v.search.Focused }
func (v *SearchView) ShortHelp() []key.Binding { return []key.Binding{Keys.Search, Keys.Enter} }

type PublicRadioListView struct {
	viewSize
	client       *api.Client
	player       *player.Player
	loading      bool
	spinner      spinner.Model
	stations     []models.PublicRadioStation
	cursor       int
	detailOffset int
	err          error
}

func NewPublicRadioListView(client *api.Client, p *player.Player) *PublicRadioListView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &PublicRadioListView{client: client, player: p, loading: true, spinner: s}
}

type publicRadioListMsg struct {
	stations []models.PublicRadioStation
	err      error
}

func (v *PublicRadioListView) Init() tea.Cmd { return tea.Batch(v.spinner.Tick, v.load()) }
func (v *PublicRadioListView) load() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewPublicRadioService(v.client)
		stations, err := svc.List(context.Background())
		return publicRadioListMsg{stations: stations, err: err}
	}
}
func (v *PublicRadioListView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case publicRadioListMsg:
		v.loading = false
		v.err = msg.err
		v.stations = msg.stations
		if v.cursor >= len(v.stations) {
			v.cursor = len(v.stations) - 1
		}
		if v.cursor < 0 {
			v.cursor = 0
		}
		return v, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if handleListDetailNavigation(keyMsg, len(v.stations), &v.cursor, &v.detailOffset, v.cHeight) {
			return v, nil
		}
		switch keyMsg.String() {
		case "enter":
			selected := selectedAt(v.stations, v.cursor)
			if selected != nil {
				slug := selected.Slug
				return v, func() tea.Msg { return NavigateMsg{View: NewPublicRadioDetailView(v.client, v.player, slug)} }
			}
		}
	}
	return v, nil
}
func (v *PublicRadioListView) View() string {
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("LIVE RADIO"))
	b.WriteString("\n\n")
	if v.loading {
		b.WriteString(v.spinner.View() + theme.MutedStyle.Render(" Loading stations..."))
	} else if v.err != nil {
		b.WriteString(theme.ErrorStyle.Render("Error: " + v.err.Error()))
	} else if len(v.stations) == 0 {
		b.WriteString(theme.MutedStyle.Render("No live stations right now."))
	} else {
		listWidth := v.cWidth / 3
		if listWidth < 28 {
			listWidth = 28
		}
		visibleRows := v.cHeight - 8
		if visibleRows < 6 {
			visibleRows = 6
		}
		items := make([]ListItem, 0, len(v.stations))
		for i := range v.stations {
			station := v.stations[i]
			nowPlaying := "-"
			if station.CurrentTrack != nil {
				nowPlaying = station.CurrentTrack.Title + " - " + station.CurrentTrack.Artist.Name
			}
			items = append(items, ListItem{Title: station.Name, Subtitle: nowPlaying, Meta: fmt.Sprintf("%d", station.ListenerCount), Selected: i == v.cursor})
		}
		listPane := renderListSection("LIVE RADIO", items, "No live stations right now.", v.cursor, visibleRows, listWidth)
		detailPane := v.renderPublicRadioSummary(visibleRows)
		b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	}
	return b.String()
}
func (v *PublicRadioListView) renderPublicRadioSummary(height int) string {
	selected := selectedAt(v.stations, v.cursor)
	if selected == nil {
		return theme.MutedStyle.Render("Select a station to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Name))
	body.WriteString("\n")
	body.WriteString(theme.GhostStyle.Render(selected.Status))
	body.WriteString(theme.MutedStyle.Render("\nListeners: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d", selected.ListenerCount)))
	if selected.CurrentTrack != nil {
		body.WriteString(theme.MutedStyle.Render("\nNow Playing: "))
		body.WriteString(theme.GhostStyle.Render(selected.CurrentTrack.Title + " - " + selected.CurrentTrack.Artist.Name))
	}
	body.WriteString("\n\nEnter opens the live station detail screen.")
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *PublicRadioListView) Title() string            { return "Live Radio" }
func (v *PublicRadioListView) ShortHelp() []key.Binding { return []key.Binding{Keys.Enter, Keys.Back} }

type PublicRadioDetailView struct {
	viewSize
	client  *api.Client
	player  *player.Player
	slug    string
	station *models.PublicRadioStation
	loading bool
	spinner spinner.Model
	err     error
}

func NewPublicRadioDetailView(client *api.Client, p *player.Player, slug string) *PublicRadioDetailView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &PublicRadioDetailView{client: client, player: p, slug: slug, loading: true, spinner: s}
}

type publicRadioDetailMsg struct {
	station *models.PublicRadioStation
	err     error
}
type publicRadioRefreshMsg struct{}

func (v *PublicRadioDetailView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.load(), v.refreshTimer())
}
func (v *PublicRadioDetailView) load() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewPublicRadioService(v.client)
		s, err := svc.Get(context.Background(), v.slug)
		return publicRadioDetailMsg{station: s, err: err}
	}
}
func (v *PublicRadioDetailView) refreshTimer() tea.Cmd {
	return tea.Tick(10*time.Second, func(time.Time) tea.Msg { return publicRadioRefreshMsg{} })
}
func (v *PublicRadioDetailView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case publicRadioDetailMsg:
		v.loading = false
		v.station = msg.station
		v.err = msg.err
		return v, nil
	case publicRadioRefreshMsg:
		return v, tea.Batch(v.load(), v.refreshTimer())
	case tea.KeyMsg:
		if msg.String() == "enter" && v.station != nil && v.station.ListenURL != "" {
			if v.isPlayingThisStation() {
				v.player.Stop()
				v.player.Queue().Clear()
				return v, nil
			}
			s := v.station
			trackTitle := s.Name
			trackArtist := ""
			if s.CurrentTrack != nil {
				trackTitle = s.CurrentTrack.Title
				trackArtist = s.CurrentTrack.Artist.Name
			}
			url := s.ListenURL
			return v, func() tea.Msg { return PlayStreamMsg{URL: url, Title: trackTitle, Artist: trackArtist, Album: s.Name} }
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	return v, nil
}
func (v *PublicRadioDetailView) View() string {
	if v.loading {
		return v.spinner.View() + theme.MutedStyle.Render(" Loading station...")
	}
	if v.err != nil {
		return theme.ErrorStyle.Render("Error: " + v.err.Error())
	}
	s := v.station
	var b strings.Builder
	b.WriteString(theme.RenderSunCompact())
	b.WriteString("\n\n")
	statusStyle := lipgloss.NewStyle().Foreground(theme.ColorDeepNavy).Background(theme.ColorGreen).Bold(true).Padding(0, 1)
	if s.Status != "active" {
		statusStyle = statusStyle.Background(theme.ColorRed)
	}
	b.WriteString(statusStyle.Render(strings.ToUpper(s.Status)))
	b.WriteString(theme.MutedStyle.Render(fmt.Sprintf("  %d listeners", s.ListenerCount)))
	b.WriteString("\n\n")
	b.WriteString(theme.GradientText(strings.ToUpper(s.Name)))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(theme.ColorGridPurple).Render(strings.Repeat("~", len(s.Name)+4)))
	b.WriteString("\n\n")
	if s.CurrentTrack != nil {
		b.WriteString(theme.MutedStyle.Render("Now Playing"))
		b.WriteString("\n")
		nowIcon := lipgloss.NewStyle().Foreground(theme.ColorHotPink).Bold(true).Render("  >> ")
		trackName := lipgloss.NewStyle().Foreground(theme.ColorGhostWhite).Bold(true).Render(s.CurrentTrack.Title)
		artistName := lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(s.CurrentTrack.Artist.Name)
		b.WriteString(nowIcon + trackName + theme.MutedStyle.Render(" - ") + artistName)
		b.WriteString("\n\n")
	}
	b.WriteString(theme.MutedStyle.Render("Stream URL"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonPurple).Render("  " + s.ListenURL))
	b.WriteString("\n\n")
	if v.isPlayingThisStation() {
		b.WriteString(lipgloss.NewStyle().Foreground(theme.ColorDeepNavy).Background(theme.ColorRed).Bold(true).Padding(0, 2).Render("ENTER: Stop Listening"))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(theme.ColorDeepNavy).Background(theme.ColorHotPink).Bold(true).Padding(0, 2).Render("ENTER: Start Listening"))
	}
	b.WriteString("\n")
	return b.String()
}
func (v *PublicRadioDetailView) isPlayingThisStation() bool {
	if v.player == nil || v.station == nil {
		return false
	}
	cur := v.player.Current()
	return cur != nil && cur.Album == v.station.Name && v.player.IsPlaying()
}
func (v *PublicRadioDetailView) Title() string { return "Station" }
func (v *PublicRadioDetailView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Enter, Keys.Back}
}
