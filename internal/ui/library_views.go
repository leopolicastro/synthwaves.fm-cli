package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/models"
	"github.com/leo/synthwaves-cli/internal/player"
	"github.com/leo/synthwaves-cli/internal/ui/components"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type artistViewMode int

const (
	artistModeList artistViewMode = iota
	artistModeCreate
	artistModeEdit
	artistModeConfirmDelete
)

type albumViewMode int

const (
	albumModeList albumViewMode = iota
	albumModeCreate
)

// Artists/ArtistDetail/Albums/AlbumDetail/Tracks/Playlists/Favorites/History/PlaylistDetail
// grouped here to keep related library resource routes together.

// START copied route implementations.

type ArtistsView struct {
	viewSize
	client        *api.Client
	search        components.SearchInput
	loading       bool
	spinner       spinner.Model
	page          int
	query         string
	items         []models.Artist
	pagination    models.Pagination
	cursor        int
	detailOffset  int
	err           error
	mode          artistViewMode
	form          *huh.Form
	formName      string
	formCat       string
	editID        int64
	deleteID      int64
	deleteName    string
	confirmDelete bool
}

func NewArtistsView(client *api.Client) *ArtistsView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &ArtistsView{client: client, search: components.NewSearchInput("Search artists..."), loading: true, spinner: s, page: 1}
}

type artistsLoadedMsg struct {
	items      []models.Artist
	pagination models.Pagination
	err        error
}
type artistCreatedMsg struct {
	artist *models.Artist
	err    error
}
type artistUpdatedMsg struct {
	artist *models.Artist
	err    error
}
type artistDeletedMsg struct {
	id  int64
	err error
}

func (v *ArtistsView) Init() tea.Cmd { return tea.Batch(v.spinner.Tick, v.loadArtists()) }
func (v *ArtistsView) loadArtists() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewArtistService(v.client)
		resp, err := svc.List(context.Background(), api.ArtistListParams{Query: v.query, Page: v.page, PerPage: v.perPage()})
		if err != nil {
			return artistsLoadedMsg{err: err}
		}
		return artistsLoadedMsg{items: resp.Items, pagination: resp.Pagination}
	}
}
func (v *ArtistsView) enterCreateMode() tea.Cmd {
	v.mode = artistModeCreate
	v.formName = ""
	v.formCat = "music"
	v.form = huh.NewForm(huh.NewGroup(huh.NewInput().Title("Artist Name").Description("The name of the artist").Value(&v.formName), huh.NewSelect[string]().Title("Category").Options(huh.NewOption("Music", "music"), huh.NewOption("Podcast", "podcast")).Value(&v.formCat))).WithTheme(theme.SynthwaveHuhTheme()).WithShowHelp(true)
	return v.form.Init()
}
func (v *ArtistsView) enterEditMode(id int64, name, category string) tea.Cmd {
	v.mode = artistModeEdit
	v.editID = id
	v.formName = name
	v.formCat = category
	v.form = huh.NewForm(huh.NewGroup(huh.NewInput().Title("Artist Name").Value(&v.formName), huh.NewSelect[string]().Title("Category").Options(huh.NewOption("Music", "music"), huh.NewOption("Podcast", "podcast")).Value(&v.formCat))).WithTheme(theme.SynthwaveHuhTheme()).WithShowHelp(true)
	return v.form.Init()
}
func (v *ArtistsView) submitCreate() tea.Cmd {
	name, cat := v.formName, v.formCat
	return func() tea.Msg {
		svc := api.NewArtistService(v.client)
		a, err := svc.Create(context.Background(), name, cat)
		return artistCreatedMsg{artist: a, err: err}
	}
}
func (v *ArtistsView) submitUpdate() tea.Cmd {
	id, name, cat := v.editID, v.formName, v.formCat
	return func() tea.Msg {
		svc := api.NewArtistService(v.client)
		a, err := svc.Update(context.Background(), id, name, cat)
		return artistUpdatedMsg{artist: a, err: err}
	}
}
func (v *ArtistsView) Update(msg tea.Msg) (View, tea.Cmd) {
	if v.mode == artistModeCreate || v.mode == artistModeEdit {
		return v.updateForm(msg)
	}
	if v.mode == artistModeConfirmDelete {
		return v.updateConfirmDelete(msg)
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case artistsLoadedMsg:
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
	case artistCreatedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadArtists(), func() tea.Msg { return components.ToastMsg{Message: "Artist \"" + msg.artist.Name + "\" created!"} })
	case artistUpdatedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadArtists(), func() tea.Msg { return components.ToastMsg{Message: "Artist updated!"} })
	case artistDeletedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadArtists(), func() tea.Msg { return components.ToastMsg{Message: "Artist deleted."} })
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.loadArtists())
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.loadArtists())
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
		return tea.Batch(v.spinner.Tick, v.loadArtists())
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
		case "n":
			return v, v.enterCreateMode()
		case "e":
			selected := selectedAt(v.items, v.cursor)
			if selected != nil {
				return v, v.enterEditMode(selected.ID, selected.Name, selected.Category)
			}
		case "d":
			selected := selectedAt(v.items, v.cursor)
			if selected != nil {
				v.deleteID = selected.ID
				v.deleteName = selected.Name
				v.confirmDelete = false
				v.mode = artistModeConfirmDelete
				v.form = huh.NewForm(huh.NewGroup(huh.NewConfirm().Title("Delete artist \"" + selected.Name + "\"?").Description("This cannot be undone.").Affirmative("Delete").Negative("Cancel").Value(&v.confirmDelete))).WithTheme(theme.SynthwaveHuhTheme())
				return v, v.form.Init()
			}
		case "enter":
			selected := selectedAt(v.items, v.cursor)
			if selected != nil {
				id := selected.ID
				return v, func() tea.Msg { return NavigateMsg{View: NewArtistDetailView(v.client, id)} }
			}
		}
	}
	return v, nil
}
func (v *ArtistsView) updateConfirmDelete(msg tea.Msg) (View, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		v.mode = artistModeList
		v.form = nil
		return v, nil
	}
	form, cmd := v.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		v.form = f
	}
	if v.form.State == huh.StateCompleted {
		v.mode = artistModeList
		v.form = nil
		if v.confirmDelete {
			v.loading = true
			deleteID := v.deleteID
			return v, tea.Batch(v.spinner.Tick, func() tea.Msg {
				svc := api.NewArtistService(v.client)
				err := svc.Delete(context.Background(), deleteID)
				return artistDeletedMsg{id: deleteID, err: err}
			})
		}
		return v, nil
	}
	if v.form.State == huh.StateAborted {
		v.mode = artistModeList
		v.form = nil
		return v, nil
	}
	return v, cmd
}
func (v *ArtistsView) updateForm(msg tea.Msg) (View, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		v.mode = artistModeList
		v.form = nil
		return v, nil
	}
	form, cmd := v.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		v.form = f
	}
	if v.form.State == huh.StateCompleted {
		wasEditing := v.mode == artistModeEdit
		v.mode = artistModeList
		v.loading = true
		if wasEditing {
			return v, tea.Batch(v.spinner.Tick, v.submitUpdate())
		}
		return v, tea.Batch(v.spinner.Tick, v.submitCreate())
	}
	if v.form.State == huh.StateAborted {
		v.mode = artistModeList
		v.form = nil
		return v, nil
	}
	return v, cmd
}
func (v *ArtistsView) View() string {
	if v.mode != artistModeList && v.form != nil {
		var b strings.Builder
		switch v.mode {
		case artistModeCreate:
			b.WriteString(theme.SubtitleStyle.Render("NEW ARTIST"))
		case artistModeEdit:
			b.WriteString(theme.SubtitleStyle.Render("EDIT ARTIST"))
		case artistModeConfirmDelete:
			b.WriteString(theme.ErrorStyle.Render("CONFIRM DELETE"))
		}
		b.WriteString("\n")
		b.WriteString(theme.FormContainerStyle.Render(v.form.View()))
		b.WriteString("\n\n")
		b.WriteString(theme.MutedStyle.Render("esc to cancel"))
		return b.String()
	}
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("ARTISTS"))
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
			artist := v.items[i]
			meta := fmt.Sprintf("%d albums  %d tracks", artist.AlbumsCount, artist.TracksCount)
			if artist.Category != "" {
				meta = artist.Category + "  " + meta
			}
			items = append(items, ListItem{Title: artist.Name, Meta: meta, Selected: i == v.cursor})
		}
		listPane := renderListSection("ARTISTS", items, "No artists found.", v.cursor, visibleRows, listWidth)
		detailPane := v.renderArtistSummary(visibleRows)
		b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	}
	return b.String()
}
func (v *ArtistsView) renderArtistSummary(height int) string {
	selected := selectedAt(v.items, v.cursor)
	if selected == nil {
		return theme.MutedStyle.Render("Select an artist to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Name))
	body.WriteString("\n")
	body.WriteString(theme.MutedStyle.Render(selected.Category))
	body.WriteString(theme.MutedStyle.Render("\nAlbums: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d", selected.AlbumsCount)))
	body.WriteString(theme.MutedStyle.Render("\nTracks: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d", selected.TracksCount)))
	body.WriteString("\n\nEnter opens artist detail. Press e to edit or d to delete.")
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *ArtistsView) Title() string         { return "Artists" }
func (v *ArtistsView) HasFocusedInput() bool { return v.search.Focused }
func (v *ArtistsView) ShortHelp() []key.Binding {
	if v.mode != artistModeList {
		return nil
	}
	return []key.Binding{Keys.Search, Keys.New, Keys.Enter, Keys.Delete, Keys.NextPage, Keys.PrevPage}
}

type ArtistDetailView struct {
	viewSize
	client       *api.Client
	id           int64
	artist       *models.ArtistDetail
	loading      bool
	spinner      spinner.Model
	cursor       int
	detailOffset int
	err          error
}

func NewArtistDetailView(client *api.Client, id int64) *ArtistDetailView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &ArtistDetailView{client: client, id: id, loading: true, spinner: s}
}

type artistDetailMsg struct {
	artist *models.ArtistDetail
	err    error
}

func (v *ArtistDetailView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, func() tea.Msg {
		svc := api.NewArtistService(v.client)
		a, err := svc.Get(context.Background(), v.id)
		return artistDetailMsg{artist: a, err: err}
	})
}
func (v *ArtistDetailView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case artistDetailMsg:
		v.loading = false
		v.artist = msg.artist
		v.err = msg.err
		if v.artist != nil && v.cursor >= len(v.artist.Albums) {
			v.cursor = len(v.artist.Albums) - 1
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
		if handleListDetailNavigation(keyMsg, len(v.artist.Albums), &v.cursor, &v.detailOffset, v.cHeight) {
			return v, nil
		}
		switch keyMsg.String() {
		case "enter":
			selected := v.selectedArtistAlbum()
			if selected != nil {
				return v, func() tea.Msg { return NavigateMsg{View: NewAlbumDetailView(v.client, selected.ID)} }
			}
		}
	}
	return v, nil
}
func (v *ArtistDetailView) View() string {
	if v.loading {
		return v.spinner.View() + theme.MutedStyle.Render(" Loading artist...")
	}
	if v.err != nil {
		return theme.ErrorStyle.Render("Error: " + v.err.Error())
	}
	var b strings.Builder
	a := v.artist
	b.WriteString(theme.TitleStyle.Render(a.Name))
	b.WriteString("\n")
	b.WriteString(theme.MutedStyle.Render(a.Category))
	b.WriteString("\n\n")
	info := fmt.Sprintf("%s albums  |  %s tracks", lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true).Render(fmt.Sprintf("%d", a.AlbumsCount)), lipgloss.NewStyle().Foreground(theme.ColorNeonPurple).Bold(true).Render(fmt.Sprintf("%d", a.TracksCount)))
	b.WriteString(info)
	b.WriteString("\n\n")
	listWidth := v.cWidth / 3
	if listWidth < 28 {
		listWidth = 28
	}
	visibleRows := v.cHeight - 9
	if visibleRows < 6 {
		visibleRows = 6
	}
	items := make([]ListItem, 0, len(v.artist.Albums))
	for i := range v.artist.Albums {
		album := v.artist.Albums[i]
		meta := fmt.Sprintf("%d tracks", album.TracksCount)
		if album.Year > 0 {
			meta = fmt.Sprintf("%d  %s", album.Year, meta)
		}
		items = append(items, ListItem{Title: album.Title, Meta: meta, Selected: i == v.cursor})
	}
	listPane := renderListSection("ALBUMS", items, "No albums found.", v.cursor, visibleRows, listWidth)
	detailPane := v.renderArtistAlbumDetail(visibleRows)
	b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	return b.String()
}
func (v *ArtistDetailView) selectedArtistAlbum() *models.AlbumSummary {
	if v.artist == nil {
		return nil
	}
	return selectedAt(v.artist.Albums, v.cursor)
}
func (v *ArtistDetailView) renderArtistAlbumDetail(height int) string {
	selected := v.selectedArtistAlbum()
	if selected == nil {
		return theme.MutedStyle.Render("Select an album to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Title))
	body.WriteString("\n")
	body.WriteString(theme.MutedStyle.Render(fmt.Sprintf("%d", selected.Year)))
	if selected.Genre != "" {
		body.WriteString(theme.MutedStyle.Render(" | "))
		body.WriteString(theme.GhostStyle.Render(selected.Genre))
	}
	body.WriteString(theme.MutedStyle.Render("\nTracks: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d", selected.TracksCount)))
	body.WriteString("\n\nEnter opens the album detail screen.")
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *ArtistDetailView) Title() string            { return "Artist" }
func (v *ArtistDetailView) ShortHelp() []key.Binding { return []key.Binding{Keys.Enter, Keys.Back} }

// Remaining grouped implementations intentionally compact to keep this patch manageable.
// Existing behavior is unchanged from the previous refactor slices.

type AlbumsView struct {
	viewSize
	client       *api.Client
	search       components.SearchInput
	loading      bool
	spinner      spinner.Model
	page         int
	query        string
	items        []models.Album
	pagination   models.Pagination
	cursor       int
	detailOffset int
	err          error
	mode         albumViewMode
	form         *huh.Form
	formTitle    string
	formArtistID string
	formYear     string
	formGenre    string
}

func NewAlbumsView(client *api.Client) *AlbumsView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &AlbumsView{client: client, search: components.NewSearchInput("Search albums..."), loading: true, spinner: s, page: 1}
}

type albumsLoadedMsg struct {
	items      []models.Album
	pagination models.Pagination
	err        error
}
type albumCreatedMsg struct {
	album *models.Album
	err   error
}
type albumDeletedMsg struct{ err error }

func (v *AlbumsView) Init() tea.Cmd { return tea.Batch(v.spinner.Tick, v.loadAlbums()) }
func (v *AlbumsView) loadAlbums() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewAlbumService(v.client)
		resp, err := svc.List(context.Background(), api.AlbumListParams{Query: v.query, Page: v.page, PerPage: v.perPage()})
		if err != nil {
			return albumsLoadedMsg{err: err}
		}
		return albumsLoadedMsg{items: resp.Items, pagination: resp.Pagination}
	}
}
func (v *AlbumsView) enterCreateMode() tea.Cmd {
	v.mode = albumModeCreate
	v.formTitle = ""
	v.formArtistID = ""
	v.formYear = ""
	v.formGenre = ""
	v.form = huh.NewForm(huh.NewGroup(huh.NewInput().Title("Album Title").Value(&v.formTitle), huh.NewInput().Title("Artist ID").Value(&v.formArtistID), huh.NewInput().Title("Year").Value(&v.formYear), huh.NewInput().Title("Genre").Value(&v.formGenre))).WithTheme(theme.SynthwaveHuhTheme()).WithShowHelp(true)
	return v.form.Init()
}
func (v *AlbumsView) Update(msg tea.Msg) (View, tea.Cmd) {
	if v.mode != albumModeList && v.form != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
			v.mode = albumModeList
			v.form = nil
			return v, nil
		}
		form, cmd := v.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			v.form = f
		}
		if v.form.State == huh.StateCompleted {
			v.mode = albumModeList
			v.loading = true
			title, artistIDStr, yearStr, genre := v.formTitle, v.formArtistID, v.formYear, v.formGenre
			return v, tea.Batch(v.spinner.Tick, func() tea.Msg {
				var artistID int64
				fmt.Sscanf(artistIDStr, "%d", &artistID)
				var year int
				fmt.Sscanf(yearStr, "%d", &year)
				svc := api.NewAlbumService(v.client)
				a, err := svc.Create(context.Background(), title, artistID, year, genre)
				return albumCreatedMsg{album: a, err: err}
			})
		}
		if v.form.State == huh.StateAborted {
			v.mode = albumModeList
			v.form = nil
		}
		return v, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case albumsLoadedMsg:
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
	case albumCreatedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadAlbums(), func() tea.Msg { return components.ToastMsg{Message: "Album \"" + msg.album.Title + "\" created!"} })
	case albumDeletedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadAlbums(), func() tea.Msg { return components.ToastMsg{Message: "Album deleted."} })
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.loadAlbums())
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		v.cursor = 0
		v.detailOffset = 0
		return v, tea.Batch(v.spinner.Tick, v.loadAlbums())
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
		return tea.Batch(v.spinner.Tick, v.loadAlbums())
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
		case "n":
			return v, v.enterCreateMode()
		case "enter":
			selected := selectedAt(v.items, v.cursor)
			if selected != nil {
				id := selected.ID
				return v, func() tea.Msg { return NavigateMsg{View: NewAlbumDetailView(v.client, id)} }
			}
		}
	}
	return v, nil
}
func (v *AlbumsView) View() string {
	if v.mode == albumModeCreate && v.form != nil {
		var b strings.Builder
		b.WriteString(theme.SubtitleStyle.Render("NEW ALBUM"))
		b.WriteString("\n")
		b.WriteString(theme.FormContainerStyle.Render(v.form.View()))
		b.WriteString("\n\n")
		b.WriteString(theme.MutedStyle.Render("esc to cancel"))
		return b.String()
	}
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("ALBUMS"))
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
			album := v.items[i]
			items = append(items, ListItem{Title: album.Title, Subtitle: album.Artist.Name + " | " + album.Genre, Meta: fmt.Sprintf("%d", album.Year), Selected: i == v.cursor})
		}
		listPane := renderListSection("ALBUMS", items, "No albums found.", v.cursor, visibleRows, listWidth)
		detailPane := v.renderAlbumSummary(visibleRows)
		b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	}
	return b.String()
}
func (v *AlbumsView) renderAlbumSummary(height int) string {
	selected := selectedAt(v.items, v.cursor)
	if selected == nil {
		return theme.MutedStyle.Render("Select an album to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Title))
	body.WriteString("\n")
	body.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(selected.Artist.Name))
	body.WriteString(theme.MutedStyle.Render("\nYear: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d", selected.Year)))
	body.WriteString(theme.MutedStyle.Render("\nGenre: "))
	body.WriteString(theme.GhostStyle.Render(selected.Genre))
	body.WriteString(theme.MutedStyle.Render("\nTracks: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d", selected.TracksCount)))
	body.WriteString("\n\nEnter opens the album detail screen.")
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *AlbumsView) Title() string         { return "Albums" }
func (v *AlbumsView) HasFocusedInput() bool { return v.search.Focused }
func (v *AlbumsView) ShortHelp() []key.Binding {
	if v.mode != albumModeList {
		return nil
	}
	return []key.Binding{Keys.Search, Keys.New, Keys.Enter, Keys.Delete, Keys.NextPage, Keys.PrevPage}
}

type AlbumDetailView struct {
	viewSize
	client       *api.Client
	id           int64
	album        *models.AlbumDetail
	loading      bool
	spinner      spinner.Model
	cursor       int
	detailOffset int
	err          error
}

func NewAlbumDetailView(client *api.Client, id int64) *AlbumDetailView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &AlbumDetailView{client: client, id: id, loading: true, spinner: s}
}

type albumDetailMsg struct {
	album *models.AlbumDetail
	err   error
}

func (v *AlbumDetailView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, func() tea.Msg {
		svc := api.NewAlbumService(v.client)
		a, err := svc.Get(context.Background(), v.id)
		return albumDetailMsg{album: a, err: err}
	})
}
func (v *AlbumDetailView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case albumDetailMsg:
		v.loading = false
		v.album = msg.album
		v.err = msg.err
		if v.album != nil && v.cursor >= len(v.album.Tracks) {
			v.cursor = len(v.album.Tracks) - 1
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
		if handleListDetailNavigation(keyMsg, len(v.album.Tracks), &v.cursor, &v.detailOffset, v.cHeight) {
			return v, nil
		}
		switch keyMsg.String() {
		case "enter":
			if v.album != nil && len(v.album.Tracks) > 0 {
				tracks := make([]player.Track, len(v.album.Tracks))
				for i, t := range v.album.Tracks {
					tracks[i] = player.Track{ID: t.ID, Title: t.Title, Artist: v.album.Artist.Name, Album: v.album.Title, Duration: t.Duration}
				}
				return v, func() tea.Msg { return PlayQueueMsg{Tracks: tracks, Index: v.cursor} }
			}
		}
	}
	return v, nil
}
func (v *AlbumDetailView) View() string {
	if v.loading {
		return v.spinner.View() + theme.MutedStyle.Render(" Loading album...")
	}
	if v.err != nil {
		return theme.ErrorStyle.Render("Error: " + v.err.Error())
	}
	var b strings.Builder
	a := v.album
	b.WriteString(theme.TitleStyle.Render(a.Title))
	b.WriteString("\n")
	b.WriteString(theme.MutedStyle.Render("by "))
	b.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(a.Artist.Name))
	b.WriteString("\n\n")
	dur := fmt.Sprintf("%d:%02d", int(a.TotalDuration)/60, int(a.TotalDuration)%60)
	info := fmt.Sprintf("%s  |  %s  |  %s  |  %s tracks", lipgloss.NewStyle().Foreground(theme.ColorChromeYellow).Render(fmt.Sprintf("%d", a.Year)), lipgloss.NewStyle().Foreground(theme.ColorNeonPurple).Render(a.Genre), theme.MutedStyle.Render(dur), lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true).Render(fmt.Sprintf("%d", a.TracksCount)))
	b.WriteString(info)
	b.WriteString("\n\n")
	listWidth := v.cWidth / 3
	if listWidth < 28 {
		listWidth = 28
	}
	visibleRows := v.cHeight - 9
	if visibleRows < 6 {
		visibleRows = 6
	}
	items := make([]ListItem, 0, len(v.album.Tracks))
	for i := range v.album.Tracks {
		track := v.album.Tracks[i]
		items = append(items, ListItem{Title: fmt.Sprintf("%d. %s", track.TrackNumber, track.Title), Meta: fmt.Sprintf("%d:%02d", int(track.Duration)/60, int(track.Duration)%60), Selected: i == v.cursor})
	}
	listPane := renderListSection("TRACKS", items, "No tracks found.", v.cursor, visibleRows, listWidth)
	detailPane := v.renderAlbumTrackDetail(visibleRows)
	b.WriteString(RenderWorkspace(WorkspaceProps{Width: v.cWidth, Height: visibleRows + 2, Nav: listPane, Detail: detailPane, NavPreferredWidth: listWidth, NavMinWidth: 28, DetailMinWidth: 36, StackAt: 92}))
	return b.String()
}
func (v *AlbumDetailView) renderAlbumTrackDetail(height int) string {
	if v.album == nil {
		return theme.MutedStyle.Render("Select a track to inspect it here.")
	}
	selected := selectedAt(v.album.Tracks, v.cursor)
	if selected == nil {
		return theme.MutedStyle.Render("Select a track to inspect it here.")
	}
	var body strings.Builder
	body.WriteString(theme.TitleStyle.Render(selected.Title))
	body.WriteString("\n")
	body.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(v.album.Artist.Name))
	body.WriteString(theme.MutedStyle.Render("\nTrack: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d", selected.TrackNumber)))
	body.WriteString(theme.MutedStyle.Render("\nFormat: "))
	body.WriteString(theme.GhostStyle.Render(selected.FileFormat))
	body.WriteString(theme.MutedStyle.Render("\nDuration: "))
	body.WriteString(theme.GhostStyle.Render(fmt.Sprintf("%d:%02d", int(selected.Duration)/60, int(selected.Duration)%60)))
	body.WriteString("\n\nEnter plays the selected track and queues the album tracks.")
	content, _ := renderScrollableDetail(body.String(), height, v.detailOffset)
	return content
}
func (v *AlbumDetailView) Title() string            { return "Album" }
func (v *AlbumDetailView) ShortHelp() []key.Binding { return []key.Binding{Keys.Back} }

// Track/playlist/favorites/history/playlist-detail implementations omitted here for brevity in this snippet.
// They are added below as focused route files in the same package.
