package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/models"
	"github.com/leo/synthwaves-cli/internal/player"
	"github.com/leo/synthwaves-cli/internal/ui/components"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

// ---------------------------------------------------------------------------
// Content area size helpers
// ---------------------------------------------------------------------------

const sidebarWidth = 26 // sidebar (22) + border + padding

// contentSize computes the usable content area from the terminal dimensions.
func contentSize(termWidth, termHeight int) (width, height int) {
	width = termWidth - sidebarWidth - 4 // 4 for content padding
	if width < 40 {
		width = 40
	}
	height = termHeight - 6 // status bar + chrome
	if height < 10 {
		height = 10
	}
	return
}

// friendlyTime formats an ISO timestamp into a short human-readable string.
func friendlyTime(iso string) string {
	t, err := time.Parse(time.RFC3339Nano, iso)
	if err != nil {
		// Try without nanoseconds
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

// viewSize is embedded in every view to track available content dimensions.
type viewSize struct {
	cWidth  int
	cHeight int
}

func (s *viewSize) handleResize(msg tea.WindowSizeMsg) {
	s.cWidth, s.cHeight = contentSize(msg.Width, msg.Height)
}

// perPage returns how many rows to fetch to fill the table height.
func (s viewSize) perPage() int {
	h := s.tHeight() - 2 // subtract header + pagination line
	if h < 20 {
		return 20
	}
	if h > 100 {
		return 100 // API max
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

// ---------------------------------------------------------------------------
// Shared messages
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Dashboard View
// ---------------------------------------------------------------------------

type DashboardView struct {
	client  *api.Client
	profile *models.Profile
	stats   *models.Stats
	loading bool
	spinner spinner.Model
	err     error
}

func NewDashboardView(client *api.Client) *DashboardView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &DashboardView{client: client, loading: true, spinner: s}
}

func (v *DashboardView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

type dashboardDataMsg struct {
	profile *models.Profile
	stats   *models.Stats
	err     error
}

func (v *DashboardView) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		profileSvc := api.NewProfileService(v.client)
		statsSvc := api.NewStatsService(v.client)
		p, err := profileSvc.Get(ctx)
		if err != nil {
			return dashboardDataMsg{err: err}
		}
		s, _ := statsSvc.Get(ctx, "month")
		return dashboardDataMsg{profile: p, stats: s}
	}
}

func (v *DashboardView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case dashboardDataMsg:
		v.loading = false
		v.profile = msg.profile
		v.stats = msg.stats
		v.err = msg.err
		return v, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *DashboardView) View() string {
	if v.loading {
		return v.spinner.View() + theme.MutedStyle.Render(" Loading dashboard...")
	}
	if v.err != nil {
		return theme.ErrorStyle.Render("Error: " + v.err.Error())
	}

	var b strings.Builder

	// Synthwave sun
	b.WriteString(theme.RenderSynthwaveSun(70))
	b.WriteString("\n")
	b.WriteString(theme.RenderSmallLogo())
	b.WriteString("\n\n")

	if v.profile != nil {
		// Profile card
		nameStyle := lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true)
		b.WriteString(nameStyle.Render(v.profile.Name))
		b.WriteString(theme.MutedStyle.Render("  " + v.profile.EmailAddress))
		b.WriteString("\n\n")

		// Library stats in a grid
		statBox := func(label string, count int, color lipgloss.Color) string {
			num := lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("%d", count))
			lbl := theme.MutedStyle.Render(label)
			return lipgloss.JoinVertical(lipgloss.Center, num, lbl)
		}

		statsRow := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(14).Align(lipgloss.Center).Render(
				statBox("Artists", v.profile.Stats.ArtistsCount, theme.ColorHotPink)),
			lipgloss.NewStyle().Width(14).Align(lipgloss.Center).Render(
				statBox("Albums", v.profile.Stats.AlbumsCount, theme.ColorNeonCyan)),
			lipgloss.NewStyle().Width(14).Align(lipgloss.Center).Render(
				statBox("Tracks", v.profile.Stats.TracksCount, theme.ColorNeonPurple)),
			lipgloss.NewStyle().Width(14).Align(lipgloss.Center).Render(
				statBox("Playlists", v.profile.Stats.PlaylistsCount, theme.ColorChromeYellow)),
		)

		b.WriteString(lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorGridPurple).
			Padding(1, 2).
			Render(statsRow))
	}

	if v.stats != nil && v.stats.Listening.TotalPlays > 0 {
		b.WriteString("\n\n")
		b.WriteString(theme.SubtitleStyle.Render("Listening This Month"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s plays  |  %s streak  |  %s best",
			lipgloss.NewStyle().Foreground(theme.ColorHotPink).Bold(true).
				Render(fmt.Sprintf("%d", v.stats.Listening.TotalPlays)),
			lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true).
				Render(fmt.Sprintf("%d days", v.stats.Listening.CurrentStreak)),
			lipgloss.NewStyle().Foreground(theme.ColorChromeYellow).Bold(true).
				Render(fmt.Sprintf("%d days", v.stats.Listening.LongestStreak)),
		))

		if len(v.stats.Listening.TopArtists) > 0 {
			b.WriteString("\n\n")
			b.WriteString(theme.MutedStyle.Render("Top Artists: "))
			var names []string
			for i, a := range v.stats.Listening.TopArtists {
				if i >= 5 {
					break
				}
				names = append(names, lipgloss.NewStyle().Foreground(theme.ColorGhostWhite).Render(a.Name))
			}
			b.WriteString(strings.Join(names, theme.MutedStyle.Render(", ")))
		}
	}

	return b.String()
}

func (v *DashboardView) Title() string          { return "Dashboard" }
func (v *DashboardView) ShortHelp() []key.Binding { return []key.Binding{Keys.Refresh} }

// ---------------------------------------------------------------------------
// Artists View (with embedded create/edit forms)
// ---------------------------------------------------------------------------

type artistViewMode int

const (
	artistModeList artistViewMode = iota
	artistModeCreate
	artistModeEdit
	artistModeConfirmDelete
)

type ArtistsView struct {
	viewSize
	client     *api.Client
	table      components.DataTable
	search     components.SearchInput
	loading    bool
	spinner    spinner.Model
	page       int
	query      string
	items      []models.Artist
	pagination models.Pagination
	err        error
	ready      bool

	// Form state
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
	return &ArtistsView{
		client:  client,
		search:  components.NewSearchInput("Search artists..."),
		loading: true,
		spinner: s,
		page:    1,
	}
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

func (v *ArtistsView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadArtists())
}

func (v *ArtistsView) loadArtists() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewArtistService(v.client)
		resp, err := svc.List(context.Background(), api.ArtistListParams{
			Query: v.query, Page: v.page, PerPage: v.perPage(),
		})
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
	v.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Artist Name").
				Description("The name of the artist").
				Value(&v.formName),
			huh.NewSelect[string]().
				Title("Category").
				Options(
					huh.NewOption("Music", "music"),
					huh.NewOption("Podcast", "podcast"),
				).
				Value(&v.formCat),
		),
	).WithTheme(theme.SynthwaveHuhTheme()).WithShowHelp(true)
	return v.form.Init()
}

func (v *ArtistsView) enterEditMode(id int64, name, category string) tea.Cmd {
	v.mode = artistModeEdit
	v.editID = id
	v.formName = name
	v.formCat = category
	v.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Artist Name").
				Value(&v.formName),
			huh.NewSelect[string]().
				Title("Category").
				Options(
					huh.NewOption("Music", "music"),
					huh.NewOption("Podcast", "podcast"),
				).
				Value(&v.formCat),
		),
	).WithTheme(theme.SynthwaveHuhTheme()).WithShowHelp(true)
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
	// Form modes: route everything to the form
	if v.mode == artistModeCreate || v.mode == artistModeEdit {
		return v.updateForm(msg)
	}
	if v.mode == artistModeConfirmDelete {
		return v.updateConfirmDelete(msg)
	}

	// Handle async results regardless of focus
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		if v.ready {
			v.table = v.table.Resize(v.tWidth(), v.tHeight())
		}
		return v, nil

	case artistsLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.items = msg.items
		v.pagination = msg.pagination
		v.buildTable()
		return v, nil

	case artistCreatedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadArtists(), func() tea.Msg {
			return components.ToastMsg{Message: "Artist \"" + msg.artist.Name + "\" created!"}
		})

	case artistUpdatedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadArtists(), func() tea.Msg {
			return components.ToastMsg{Message: "Artist updated!"}
		})

	case artistDeletedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadArtists(), func() tea.Msg {
			return components.ToastMsg{Message: "Artist deleted."}
		})

	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadArtists())

	case components.RowSelectMsg:
		if len(msg.Row) > 0 {
			var id int64
			fmt.Sscanf(msg.Row[0], "%d", &id)
			return v, func() tea.Msg {
				return NavigateMsg{View: NewArtistDetailView(v.client, id)}
			}
		}

	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadArtists())

	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	// When search is focused, route ALL messages to it (keys + debounce ticks)
	if v.search.Focused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				v.search = v.search.Blur()
				return v, nil
			}
			if keyMsg.String() == "enter" {
				// Enter triggers immediate search and blurs
				v.query = v.search.Value()
				v.page = 1
				v.loading = true
				v.search = v.search.Blur()
				return v, tea.Batch(v.spinner.Tick, v.loadArtists())
			}
		}
		var cmd tea.Cmd
		v.search, cmd = v.search.Update(msg)
		return v, cmd
	}

	// Search is NOT focused: handle action keys
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "/":
			v.search = v.search.Focus()
			return v, nil
		case "n":
			return v, v.enterCreateMode()
		case "e":
			if v.ready {
				row := v.table.SelectedRow()
				if row != nil && len(row) >= 3 {
					var id int64
					fmt.Sscanf(row[0], "%d", &id)
					return v, v.enterEditMode(id, row[1], row[2])
				}
			}
		case "d":
			if v.ready {
				row := v.table.SelectedRow()
				if row != nil && len(row) >= 2 {
					var id int64
					fmt.Sscanf(row[0], "%d", &id)
					v.deleteID = id
					v.deleteName = row[1]
					v.confirmDelete = false
					v.mode = artistModeConfirmDelete
					v.form = huh.NewForm(
						huh.NewGroup(
							huh.NewConfirm().
								Title("Delete artist \"" + row[1] + "\"?").
								Description("This cannot be undone.").
								Affirmative("Delete").
								Negative("Cancel").
								Value(&v.confirmDelete),
						),
					).WithTheme(theme.SynthwaveHuhTheme())
					return v, v.form.Init()
				}
			}
		}
	}

	// Pass remaining messages to table (navigation keys: j/k/up/down/enter)
	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
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
	// Esc cancels the form
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		v.mode = artistModeList
		v.form = nil
		return v, nil
	}

	form, cmd := v.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		v.form = f
	}

	// Check if form completed
	if v.form.State == huh.StateCompleted {
		wasEditing := v.mode == artistModeEdit
		v.mode = artistModeList
		v.loading = true
		if wasEditing {
			return v, tea.Batch(v.spinner.Tick, v.submitUpdate())
		}
		return v, tea.Batch(v.spinner.Tick, v.submitCreate())
	}

	// Check if form was aborted
	if v.form.State == huh.StateAborted {
		v.mode = artistModeList
		v.form = nil
		return v, nil
	}

	return v, cmd
}

func (v *ArtistsView) buildTable() {
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Name", Width: 30},
		{Title: "Category", Width: 12},
		{Title: "Albums", Width: 8},
		{Title: "Tracks", Width: 8},
	}
	rows := make([]table.Row, len(v.items))
	for i, a := range v.items {
		rows[i] = table.Row{
			fmt.Sprintf("%d", a.ID), a.Name, a.Category,
			fmt.Sprintf("%d", a.AlbumsCount), fmt.Sprintf("%d", a.TracksCount),
		}
	}
	v.table = components.NewDataTable(columns, rows, v.pagination, v.tHeight(), v.tWidth())
	v.ready = true
}

func (v *ArtistsView) View() string {
	// Form/confirm modes -- show the form in a neon box
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

	// List mode
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("ARTISTS"))
	b.WriteString("\n\n")
	b.WriteString(v.search.View())
	b.WriteString("\n\n")

	if v.loading {
		b.WriteString(v.spinner.View() + theme.MutedStyle.Render(" Loading..."))
	} else if v.err != nil {
		b.WriteString(theme.ErrorStyle.Render("Error: " + v.err.Error()))
	} else if v.ready {
		b.WriteString(v.table.View())
	}
	return b.String()
}

func (v *ArtistsView) Title() string            { return "Artists" }
func (v *ArtistsView) HasFocusedInput() bool     { return v.search.Focused }
func (v *ArtistsView) ShortHelp() []key.Binding {
	if v.mode != artistModeList {
		return nil
	}
	return []key.Binding{Keys.Search, Keys.New, Keys.Enter, Keys.Delete, Keys.NextPage, Keys.PrevPage}
}

// ---------------------------------------------------------------------------
// Artist Detail View
// ---------------------------------------------------------------------------

type ArtistDetailView struct {
	viewSize
	client  *api.Client
	id      int64
	artist  *models.ArtistDetail
	loading bool
	spinner spinner.Model
	table   components.DataTable
	err     error
	ready   bool
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
		if v.artist != nil && len(v.artist.Albums) > 0 {
			v.buildAlbumsTable()
		}
		return v, nil
	case components.RowSelectMsg:
		if len(msg.Row) > 0 {
			var id int64
			fmt.Sscanf(msg.Row[0], "%d", &id)
			return v, func() tea.Msg {
				return NavigateMsg{View: NewAlbumDetailView(v.client, id)}
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *ArtistDetailView) buildAlbumsTable() {
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Title", Width: 30},
		{Title: "Year", Width: 6},
		{Title: "Genre", Width: 15},
		{Title: "Tracks", Width: 8},
	}
	rows := make([]table.Row, len(v.artist.Albums))
	for i, a := range v.artist.Albums {
		rows[i] = table.Row{
			fmt.Sprintf("%d", a.ID), a.Title,
			fmt.Sprintf("%d", a.Year), a.Genre,
			fmt.Sprintf("%d", a.TracksCount),
		}
	}
	v.table = components.NewDataTable(columns, rows, models.Pagination{}, v.tHeight(), v.tWidth())
	v.ready = true
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

	info := fmt.Sprintf("%s albums  |  %s tracks",
		lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true).Render(fmt.Sprintf("%d", a.AlbumsCount)),
		lipgloss.NewStyle().Foreground(theme.ColorNeonPurple).Bold(true).Render(fmt.Sprintf("%d", a.TracksCount)),
	)
	b.WriteString(info)
	b.WriteString("\n\n")

	if v.ready {
		b.WriteString(theme.SubtitleStyle.Render("Albums"))
		b.WriteString("\n")
		b.WriteString(v.table.View())
	}

	return b.String()
}

func (v *ArtistDetailView) Title() string { return "Artist" }
func (v *ArtistDetailView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Enter, Keys.Back}
}

// ---------------------------------------------------------------------------
// Albums View
// ---------------------------------------------------------------------------

type albumViewMode int

const (
	albumModeList albumViewMode = iota
	albumModeCreate
)

type AlbumsView struct {
	viewSize
	client     *api.Client
	table      components.DataTable
	search     components.SearchInput
	loading    bool
	spinner    spinner.Model
	page       int
	query      string
	items      []models.Album
	pagination models.Pagination
	err        error
	ready      bool

	// Form state
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
	return &AlbumsView{
		client:  client,
		search:  components.NewSearchInput("Search albums..."),
		loading: true,
		spinner: s,
		page:    1,
	}
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

type albumDeletedMsg struct {
	err error
}

func (v *AlbumsView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadAlbums())
}

func (v *AlbumsView) loadAlbums() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewAlbumService(v.client)
		resp, err := svc.List(context.Background(), api.AlbumListParams{
			Query: v.query, Page: v.page, PerPage: v.perPage(),
		})
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
	v.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Album Title").Value(&v.formTitle),
			huh.NewInput().Title("Artist ID").Value(&v.formArtistID),
			huh.NewInput().Title("Year").Value(&v.formYear),
			huh.NewInput().Title("Genre").Value(&v.formGenre),
		),
	).WithTheme(theme.SynthwaveHuhTheme()).WithShowHelp(true)
	return v.form.Init()
}

func (v *AlbumsView) Update(msg tea.Msg) (View, tea.Cmd) {
	// Form mode -- route to form
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

	// Async results
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.handleResize(msg)
		return v, nil
	case albumsLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.items = msg.items
		v.pagination = msg.pagination
		v.buildTable()
		return v, nil
	case albumCreatedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadAlbums(), func() tea.Msg {
			return components.ToastMsg{Message: "Album \"" + msg.album.Title + "\" created!"}
		})
	case albumDeletedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadAlbums(), func() tea.Msg {
			return components.ToastMsg{Message: "Album deleted."}
		})
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadAlbums())
	case components.RowSelectMsg:
		if len(msg.Row) > 0 {
			var id int64
			fmt.Sscanf(msg.Row[0], "%d", &id)
			return v, func() tea.Msg {
				return NavigateMsg{View: NewAlbumDetailView(v.client, id)}
			}
		}
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadAlbums())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	// When search is focused, route ALL messages to it
	if v.search.Focused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				v.search = v.search.Blur()
				return v, nil
			}
			if keyMsg.String() == "enter" {
				v.query = v.search.Value()
				v.page = 1
				v.loading = true
				v.search = v.search.Blur()
				return v, tea.Batch(v.spinner.Tick, v.loadAlbums())
			}
		}
		var cmd tea.Cmd
		v.search, cmd = v.search.Update(msg)
		return v, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "/":
			v.search = v.search.Focus()
			return v, nil
		case "n":
			return v, v.enterCreateMode()
		}
	}

	// Table navigation
	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *AlbumsView) buildTable() {
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Title", Width: 25},
		{Title: "Artist", Width: 20},
		{Title: "Year", Width: 6},
		{Title: "Genre", Width: 12},
		{Title: "Tracks", Width: 8},
	}
	rows := make([]table.Row, len(v.items))
	for i, a := range v.items {
		rows[i] = table.Row{
			fmt.Sprintf("%d", a.ID), a.Title, a.Artist.Name,
			fmt.Sprintf("%d", a.Year), a.Genre,
			fmt.Sprintf("%d", a.TracksCount),
		}
	}
	v.table = components.NewDataTable(columns, rows, v.pagination, v.tHeight(), v.tWidth())
	v.ready = true
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
	} else if v.ready {
		b.WriteString(v.table.View())
	}
	return b.String()
}

func (v *AlbumsView) Title() string            { return "Albums" }
func (v *AlbumsView) HasFocusedInput() bool     { return v.search.Focused }
func (v *AlbumsView) ShortHelp() []key.Binding {
	if v.mode != albumModeList {
		return nil
	}
	return []key.Binding{Keys.Search, Keys.New, Keys.Enter, Keys.Delete, Keys.NextPage, Keys.PrevPage}
}

// ---------------------------------------------------------------------------
// Album Detail View
// ---------------------------------------------------------------------------

type AlbumDetailView struct {
	viewSize
	client  *api.Client
	id      int64
	album   *models.AlbumDetail
	loading bool
	spinner spinner.Model
	table   components.DataTable
	err     error
	ready   bool
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
		if v.album != nil && len(v.album.Tracks) > 0 {
			v.buildTable()
		}
		return v, nil
	case components.RowSelectMsg:
		// Queue all album tracks and start from selected
		if v.album != nil && len(msg.Row) > 0 {
			var trackNum int
			fmt.Sscanf(msg.Row[0], "%d", &trackNum)
			selectedIndex := 0
			tracks := make([]player.Track, len(v.album.Tracks))
			for i, t := range v.album.Tracks {
				tracks[i] = player.Track{
					ID: t.ID, Title: t.Title,
					Artist: v.album.Artist.Name, Album: v.album.Title,
					Duration: t.Duration,
				}
				if t.TrackNumber == trackNum {
					selectedIndex = i
				}
			}
			return v, func() tea.Msg {
				return PlayQueueMsg{Tracks: tracks, Index: selectedIndex}
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *AlbumDetailView) buildTable() {
	columns := []table.Column{
		{Title: "#", Width: 4},
		{Title: "Title", Width: 35},
		{Title: "Duration", Width: 10},
		{Title: "Format", Width: 8},
	}
	rows := make([]table.Row, len(v.album.Tracks))
	for i, t := range v.album.Tracks {
		dur := fmt.Sprintf("%d:%02d", int(t.Duration)/60, int(t.Duration)%60)
		rows[i] = table.Row{
			fmt.Sprintf("%d", t.TrackNumber), t.Title, dur, t.FileFormat,
		}
	}
	v.table = components.NewDataTable(columns, rows, models.Pagination{}, v.tHeight(), v.tWidth())
	v.ready = true
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
	info := fmt.Sprintf("%s  |  %s  |  %s  |  %s tracks",
		lipgloss.NewStyle().Foreground(theme.ColorChromeYellow).Render(fmt.Sprintf("%d", a.Year)),
		lipgloss.NewStyle().Foreground(theme.ColorNeonPurple).Render(a.Genre),
		theme.MutedStyle.Render(dur),
		lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true).Render(fmt.Sprintf("%d", a.TracksCount)),
	)
	b.WriteString(info)
	b.WriteString("\n\n")

	if v.ready {
		b.WriteString(v.table.View())
	}

	return b.String()
}

func (v *AlbumDetailView) Title() string          { return "Album" }
func (v *AlbumDetailView) ShortHelp() []key.Binding { return []key.Binding{Keys.Back} }

// ---------------------------------------------------------------------------
// Tracks View
// ---------------------------------------------------------------------------

type TracksView struct {
	viewSize
	client     *api.Client
	table      components.DataTable
	search     components.SearchInput
	loading    bool
	spinner    spinner.Model
	page       int
	query      string
	items      []models.Track
	pagination models.Pagination
	err        error
	ready      bool
}

func NewTracksView(client *api.Client) *TracksView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &TracksView{
		client:  client,
		search:  components.NewSearchInput("Search tracks..."),
		loading: true,
		spinner: s,
		page:    1,
	}
}

type tracksLoadedMsg struct {
	items      []models.Track
	pagination models.Pagination
	err        error
}

func (v *TracksView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadTracks())
}

func (v *TracksView) loadTracks() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewTrackService(v.client)
		resp, err := svc.List(context.Background(), api.TrackListParams{
			Query: v.query, Page: v.page, PerPage: v.perPage(),
		})
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
		v.buildTable()
		return v, nil
	case components.RowSelectMsg:
		// Queue all current page tracks and start from selected
		if len(msg.Row) > 0 {
			var id int64
			fmt.Sscanf(msg.Row[0], "%d", &id)
			selectedIndex := 0
			tracks := make([]player.Track, len(v.items))
			for i, t := range v.items {
				tracks[i] = player.Track{
					ID: t.ID, Title: t.Title,
					Artist: t.Artist.Name, Album: t.Album.Title,
					Duration: t.Duration,
				}
				if t.ID == id {
					selectedIndex = i
				}
			}
			return v, func() tea.Msg {
				return PlayQueueMsg{Tracks: tracks, Index: selectedIndex}
			}
		}
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadTracks())
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.loadTracks())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	if v.search.Focused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				v.search = v.search.Blur()
				return v, nil
			}
			if keyMsg.String() == "enter" {
				v.query = v.search.Value()
				v.page = 1
				v.loading = true
				v.search = v.search.Blur()
				return v, tea.Batch(v.spinner.Tick, v.loadTracks())
			}
		}
		var cmd tea.Cmd
		v.search, cmd = v.search.Update(msg)
		return v, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "/" {
			v.search = v.search.Focus()
			return v, nil
		}
	}

	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *TracksView) buildTable() {
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Title", Width: 25},
		{Title: "Artist", Width: 18},
		{Title: "Album", Width: 18},
		{Title: "Duration", Width: 8},
		{Title: "Format", Width: 6},
	}
	rows := make([]table.Row, len(v.items))
	for i, t := range v.items {
		dur := fmt.Sprintf("%d:%02d", int(t.Duration)/60, int(t.Duration)%60)
		rows[i] = table.Row{
			fmt.Sprintf("%d", t.ID), t.Title, t.Artist.Name,
			t.Album.Title, dur, t.FileFormat,
		}
	}
	v.table = components.NewDataTable(columns, rows, v.pagination, v.tHeight(), v.tWidth())
	v.ready = true
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
	} else if v.ready {
		b.WriteString(v.table.View())
	}
	return b.String()
}

func (v *TracksView) Title() string            { return "Tracks" }
func (v *TracksView) HasFocusedInput() bool     { return v.search.Focused }
func (v *TracksView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.Enter, Keys.NextPage, Keys.PrevPage}
}

// ---------------------------------------------------------------------------
// Playlists View
// ---------------------------------------------------------------------------

type PlaylistsView struct {
	viewSize
	client     *api.Client
	table      components.DataTable
	search     components.SearchInput
	loading    bool
	spinner    spinner.Model
	page       int
	query      string
	items      []models.Playlist
	pagination models.Pagination
	err        error
	ready      bool
}

func NewPlaylistsView(client *api.Client) *PlaylistsView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &PlaylistsView{
		client:  client,
		search:  components.NewSearchInput("Search playlists..."),
		loading: true,
		spinner: s,
		page:    1,
	}
}

type playlistsLoadedMsg struct {
	items      []models.Playlist
	pagination models.Pagination
	err        error
}

func (v *PlaylistsView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.load())
}

func (v *PlaylistsView) load() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewPlaylistService(v.client)
		resp, err := svc.List(context.Background(), api.PlaylistListParams{
			Query: v.query, Page: v.page, PerPage: v.perPage(),
		})
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
		v.buildTable()
		return v, nil
	case components.RowSelectMsg:
		if len(msg.Row) > 0 {
			var id int64
			fmt.Sscanf(msg.Row[0], "%d", &id)
			return v, func() tea.Msg {
				return NavigateMsg{View: NewPlaylistDetailView(v.client, id)}
			}
		}
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.load())
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.load())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	if v.search.Focused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				v.search = v.search.Blur()
				return v, nil
			}
			if keyMsg.String() == "enter" {
				v.query = v.search.Value()
				v.page = 1
				v.loading = true
				v.search = v.search.Blur()
				return v, tea.Batch(v.spinner.Tick, v.load())
			}
		}
		var cmd tea.Cmd
		v.search, cmd = v.search.Update(msg)
		return v, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "/" {
			v.search = v.search.Focus()
			return v, nil
		}
	}

	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *PlaylistsView) buildTable() {
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Name", Width: 30},
		{Title: "Tracks", Width: 8},
		{Title: "Updated", Width: 20},
	}
	rows := make([]table.Row, len(v.items))
	for i, p := range v.items {
		rows[i] = table.Row{
			fmt.Sprintf("%d", p.ID), p.Name,
			fmt.Sprintf("%d", p.TracksCount), friendlyTime(p.UpdatedAt),
		}
	}
	v.table = components.NewDataTable(columns, rows, v.pagination, v.tHeight(), v.tWidth())
	v.ready = true
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
	} else if v.ready {
		b.WriteString(v.table.View())
	}
	return b.String()
}

func (v *PlaylistsView) Title() string            { return "Playlists" }
func (v *PlaylistsView) HasFocusedInput() bool     { return v.search.Focused }
func (v *PlaylistsView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.Enter, Keys.NextPage, Keys.PrevPage}
}

// ---------------------------------------------------------------------------
// Favorites View
// ---------------------------------------------------------------------------

type FavoritesView struct {
	viewSize
	client     *api.Client
	table      components.DataTable
	search     components.SearchInput
	loading    bool
	spinner    spinner.Model
	page       int
	query      string
	favType    string
	items      []models.Favorite
	pagination models.Pagination
	err        error
	ready      bool
}

func NewFavoritesView(client *api.Client) *FavoritesView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &FavoritesView{
		client:  client,
		search:  components.NewSearchInput("Search favorites..."),
		loading: true,
		spinner: s,
		page:    1,
	}
}

type favoritesLoadedMsg struct {
	items      []models.Favorite
	pagination models.Pagination
	err        error
}

func (v *FavoritesView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.load())
}

func (v *FavoritesView) load() tea.Cmd {
	return func() tea.Msg {
		svc := api.NewFavoriteService(v.client)
		resp, err := svc.List(context.Background(), api.FavoriteListParams{
			Type: v.favType, Query: v.query, Page: v.page, PerPage: v.perPage(),
		})
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
		// Client-side filter if the API doesn't support search
		if v.query != "" {
			q := strings.ToLower(v.query)
			filtered := v.items[:0]
			for _, f := range v.items {
				if strings.Contains(strings.ToLower(f.Favorable.DisplayName()), q) ||
					strings.Contains(strings.ToLower(f.Favorable.ArtistName()), q) {
					filtered = append(filtered, f)
				}
			}
			v.items = filtered
		}
		v.buildTable()
		return v, nil
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.load())
	case components.RowSelectMsg:
		if len(msg.Row) >= 2 {
			favType := msg.Row[0]
			name := msg.Row[1]
			for _, fav := range v.items {
				if fav.FavorableType != favType || fav.Favorable.DisplayName() != name {
					continue
				}
				f := fav
				switch favType {
				case "Track":
					return v, func() tea.Msg {
						return PlayTrackMsg{
							TrackID: f.Favorable.ID,
							Title:   f.Favorable.DisplayName(),
							Artist:  f.Favorable.ArtistName(),
						}
					}
				case "Album":
					return v, func() tea.Msg {
						return NavigateMsg{View: NewAlbumDetailView(v.client, f.Favorable.ID)}
					}
				case "Artist":
					return v, func() tea.Msg {
						return NavigateMsg{View: NewArtistDetailView(v.client, f.Favorable.ID)}
					}
				}
				break
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			v.favType = "Track"
			v.page = 1
			v.loading = true
			return v, tea.Batch(v.spinner.Tick, v.load())
		case "2":
			v.favType = "Album"
			v.page = 1
			v.loading = true
			return v, tea.Batch(v.spinner.Tick, v.load())
		case "3":
			v.favType = "Artist"
			v.page = 1
			v.loading = true
			return v, tea.Batch(v.spinner.Tick, v.load())
		case "0":
			v.favType = ""
			v.page = 1
			v.loading = true
			return v, tea.Batch(v.spinner.Tick, v.load())
		}
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.load())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	if v.search.Focused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				v.search = v.search.Blur()
				return v, nil
			}
			if keyMsg.String() == "enter" {
				v.query = v.search.Value()
				v.page = 1
				v.loading = true
				v.search = v.search.Blur()
				return v, tea.Batch(v.spinner.Tick, v.load())
			}
		}
		var cmd tea.Cmd
		v.search, cmd = v.search.Update(msg)
		return v, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "/" {
			v.search = v.search.Focus()
			return v, nil
		}
	}

	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *FavoritesView) buildTable() {
	columns := []table.Column{
		{Title: "Type", Width: 8},
		{Title: "Name", Width: 30},
		{Title: "Artist", Width: 20},
		{Title: "Favorited", Width: 14},
	}
	rows := make([]table.Row, len(v.items))
	for i, f := range v.items {
		rows[i] = table.Row{
			f.FavorableType,
			f.Favorable.DisplayName(),
			f.Favorable.ArtistName(),
			friendlyTime(f.CreatedAt),
		}
	}
	v.table = components.NewDataTable(columns, rows, v.pagination, v.tHeight(), v.tWidth())
	v.ready = true
}

func (v *FavoritesView) View() string {
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("FAVORITES"))
	b.WriteString("  ")

	// Tab-like filter buttons
	tabs := []struct{ key, label, val string }{
		{"0", "All", ""}, {"1", "Tracks", "Track"}, {"2", "Albums", "Album"}, {"3", "Artists", "Artist"},
	}
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
	} else if v.ready {
		b.WriteString(v.table.View())
	}
	return b.String()
}

func (v *FavoritesView) Title() string            { return "Favorites" }
func (v *FavoritesView) HasFocusedInput() bool     { return v.search.Focused }
func (v *FavoritesView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.Enter, Keys.NextPage, Keys.PrevPage}
}

// ---------------------------------------------------------------------------
// History View
// ---------------------------------------------------------------------------

type HistoryView struct {
	viewSize
	client     *api.Client
	table      components.DataTable
	search     components.SearchInput
	loading    bool
	spinner    spinner.Model
	page       int
	query      string
	items      []models.PlayHistory
	pagination models.Pagination
	err        error
	ready      bool
}

func NewHistoryView(client *api.Client) *HistoryView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &HistoryView{
		client:  client,
		search:  components.NewSearchInput("Search history..."),
		loading: true,
		spinner: s,
		page:    1,
	}
}

type historyLoadedMsg struct {
	items      []models.PlayHistory
	pagination models.Pagination
	err        error
}

func (v *HistoryView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.load())
}

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
		if v.query != "" {
			q := strings.ToLower(v.query)
			filtered := v.items[:0]
			for _, ph := range v.items {
				if strings.Contains(strings.ToLower(ph.Track.Title), q) ||
					strings.Contains(strings.ToLower(ph.Track.Artist.Name), q) {
					filtered = append(filtered, ph)
				}
			}
			v.items = filtered
		}
		v.buildTable()
		return v, nil
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.load())
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.load())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	if v.search.Focused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				v.search = v.search.Blur()
				return v, nil
			}
			if keyMsg.String() == "enter" {
				v.query = v.search.Value()
				v.page = 1
				v.loading = true
				v.search = v.search.Blur()
				return v, tea.Batch(v.spinner.Tick, v.load())
			}
		}
		var cmd tea.Cmd
		v.search, cmd = v.search.Update(msg)
		return v, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "/" {
			v.search = v.search.Focus()
			return v, nil
		}
	}

	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *HistoryView) buildTable() {
	columns := []table.Column{
		{Title: "Track", Width: 25},
		{Title: "Artist", Width: 20},
		{Title: "Played At", Width: 22},
	}
	rows := make([]table.Row, len(v.items))
	for i, ph := range v.items {
		rows[i] = table.Row{ph.Track.Title, ph.Track.Artist.Name, ph.PlayedAt}
	}
	v.table = components.NewDataTable(columns, rows, v.pagination, v.tHeight(), v.tWidth())
	v.ready = true
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
	} else if v.ready {
		b.WriteString(v.table.View())
	}
	return b.String()
}

func (v *HistoryView) Title() string            { return "History" }
func (v *HistoryView) HasFocusedInput() bool     { return v.search.Focused }
func (v *HistoryView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.NextPage, Keys.PrevPage}
}

// ---------------------------------------------------------------------------
// Radio View
// ---------------------------------------------------------------------------

type RadioView struct {
	viewSize
	client   *api.Client
	table    components.DataTable
	loading  bool
	spinner  spinner.Model
	stations []models.RadioStation
	err      error
	ready    bool
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

func (v *RadioView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.load())
}

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
		v.buildTable()
		return v, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *RadioView) buildTable() {
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Name", Width: 22},
		{Title: "Status", Width: 8},
		{Title: "Mode", Width: 10},
		{Title: "Bitrate", Width: 8},
		{Title: "Now Playing", Width: 25},
	}
	rows := make([]table.Row, len(v.stations))
	for i, s := range v.stations {
		nowPlaying := "-"
		if s.CurrentTrack != nil {
			nowPlaying = s.CurrentTrack.Title
		}
		rows[i] = table.Row{
			fmt.Sprintf("%d", s.ID), s.Name,
			theme.StatusDot(s.Status) + " " + s.Status,
			s.PlaybackMode, fmt.Sprintf("%dk", s.Bitrate),
			nowPlaying,
		}
	}
	v.table = components.NewDataTable(columns, rows, models.Pagination{}, 20, 80)
	v.ready = true
}

func (v *RadioView) View() string {
	var b strings.Builder
	b.WriteString(theme.SubtitleStyle.Render("RADIO STATIONS"))
	b.WriteString("\n\n")
	if v.loading {
		b.WriteString(v.spinner.View() + theme.MutedStyle.Render(" Loading..."))
	} else if v.err != nil {
		b.WriteString(theme.ErrorStyle.Render("Error: " + v.err.Error()))
	} else if v.ready {
		b.WriteString(v.table.View())
	}
	return b.String()
}

func (v *RadioView) Title() string          { return "Radio" }
func (v *RadioView) ShortHelp() []key.Binding { return []key.Binding{Keys.Enter} }

// ---------------------------------------------------------------------------
// Search View
// ---------------------------------------------------------------------------

type searchItem struct {
	kind string // "Track", "Album", "Artist"
	id   int64
	line string // pre-rendered display line
	// Track-specific fields for playback
	title    string
	artist   string
	album    string
	duration float64
}

type SearchView struct {
	viewSize
	client  *api.Client
	search  components.SearchInput
	loading bool
	spinner spinner.Model
	result  *models.SearchResult
	items   []searchItem // flattened results for navigation
	cursor  int
	err     error
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

func (v *SearchView) Init() tea.Cmd {
	return nil
}

func (v *SearchView) buildItems() {
	v.items = nil
	v.cursor = 0
	if v.result == nil {
		return
	}
	for _, a := range v.result.Artists {
		v.items = append(v.items, searchItem{
			kind:  "Artist",
			id:    a.ID,
			title: a.Name,
		})
	}
	for _, a := range v.result.Albums {
		v.items = append(v.items, searchItem{
			kind:   "Album",
			id:     a.ID,
			title:  a.Title,
			artist: a.Artist.Name,
		})
	}
	for _, t := range v.result.Tracks {
		v.items = append(v.items, searchItem{
			kind:     "Track",
			id:       t.ID,
			title:    t.Title,
			artist:   t.Artist.Name,
			album:    t.Album.Title,
			duration: t.Duration,
		})
	}
}

func (v *SearchView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case components.SearchQueryMsg:
		if msg.Query == "" {
			v.result = nil
			v.items = nil
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

	// When search is focused, route keys to the search input
	if v.search.Focused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				v.search = v.search.Blur()
				return v, nil
			}
		}
		var cmd tea.Cmd
		v.search, cmd = v.search.Update(msg)
		return v, cmd
	}

	// When search is blurred, handle navigation
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "/":
			v.search = v.search.Focus()
			return v, nil
		case "j", "down":
			if v.cursor < len(v.items)-1 {
				v.cursor++
			}
			return v, nil
		case "k", "up":
			if v.cursor > 0 {
				v.cursor--
			}
			return v, nil
		case "enter":
			if v.cursor < len(v.items) {
				item := v.items[v.cursor]
				switch item.kind {
				case "Track":
					return v, func() tea.Msg {
						return PlayTrackMsg{
							TrackID:  item.id,
							Title:    item.title,
							Artist:   item.artist,
							Album:    item.album,
							Duration: item.duration,
						}
					}
				case "Album":
					return v, func() tea.Msg {
						return NavigateMsg{View: NewAlbumDetailView(v.client, item.id)}
					}
				case "Artist":
					return v, func() tea.Msg {
						return NavigateMsg{View: NewArtistDetailView(v.client, item.id)}
					}
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
	b.WriteString("\n")

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

	if len(v.items) == 0 {
		b.WriteString(theme.MutedStyle.Render("No results found."))
		return b.String()
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(theme.ColorDeepNavy).
		Background(theme.ColorNeonCyan).
		Bold(true)

	idx := 0
	writeItem := func(label string) {
		if idx == v.cursor {
			b.WriteString(selectedStyle.Render("> " + label))
		} else {
			b.WriteString("  " + label)
		}
		b.WriteString("\n")
		idx++
	}

	if len(v.result.Artists) > 0 {
		b.WriteString(theme.TitleStyle.Render(fmt.Sprintf("Artists (%d)", len(v.result.Artists))))
		b.WriteString("\n")
		for _, a := range v.result.Artists {
			writeItem(theme.GhostStyle.Render(a.Name))
		}
		b.WriteString("\n")
	}

	if len(v.result.Albums) > 0 {
		b.WriteString(theme.TitleStyle.Render(fmt.Sprintf("Albums (%d)", len(v.result.Albums))))
		b.WriteString("\n")
		for _, a := range v.result.Albums {
			writeItem(theme.GhostStyle.Render(a.Title) +
				theme.MutedStyle.Render(" by ") +
				lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(a.Artist.Name))
		}
		b.WriteString("\n")
	}

	if len(v.result.Tracks) > 0 {
		b.WriteString(theme.TitleStyle.Render(fmt.Sprintf("Tracks (%d)", len(v.result.Tracks))))
		b.WriteString("\n")
		for _, t := range v.result.Tracks {
			dur := fmt.Sprintf("%d:%02d", int(t.Duration)/60, int(t.Duration)%60)
			writeItem(theme.GhostStyle.Render(t.Title) +
				theme.MutedStyle.Render(" - ") +
				lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(t.Artist.Name) +
				theme.MutedStyle.Render(" "+dur))
		}
	}

	if !v.search.Focused {
		b.WriteString("\n")
		b.WriteString(theme.MutedStyle.Render("  j/k:navigate  enter:select  /:search again"))
	}

	return b.String()
}

func (v *SearchView) Title() string            { return "Search" }
func (v *SearchView) HasFocusedInput() bool     { return v.search.Focused }
func (v *SearchView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.Enter}
}

// ---------------------------------------------------------------------------
// Stats View
// ---------------------------------------------------------------------------

type StatsView struct {
	client  *api.Client
	loading bool
	spinner spinner.Model
	stats   *models.Stats
	err     error
}

func NewStatsView(client *api.Client) *StatsView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &StatsView{client: client, loading: true, spinner: s}
}

type statsLoadedMsg struct {
	stats *models.Stats
	err   error
}

func (v *StatsView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, func() tea.Msg {
		svc := api.NewStatsService(v.client)
		s, err := svc.Get(context.Background(), "month")
		return statsLoadedMsg{stats: s, err: err}
	})
}

func (v *StatsView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case statsLoadedMsg:
		v.loading = false
		v.stats = msg.stats
		v.err = msg.err
		return v, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *StatsView) View() string {
	if v.loading {
		return v.spinner.View() + theme.MutedStyle.Render(" Loading stats...")
	}
	if v.err != nil {
		return theme.ErrorStyle.Render("Error: " + v.err.Error())
	}

	var b strings.Builder
	s := v.stats

	// Library section
	b.WriteString(theme.TitleStyle.Render("LIBRARY"))
	b.WriteString("\n\n")

	libStats := []struct{ label string; value int; color lipgloss.Color }{
		{"Artists", s.Library.ArtistsCount, theme.ColorHotPink},
		{"Albums", s.Library.AlbumsCount, theme.ColorNeonCyan},
		{"Tracks", s.Library.TracksCount, theme.ColorNeonPurple},
		{"Playlists", s.Library.PlaylistsCount, theme.ColorChromeYellow},
	}
	for _, ls := range libStats {
		num := lipgloss.NewStyle().Foreground(ls.color).Bold(true).
			Width(8).Align(lipgloss.Right).
			Render(fmt.Sprintf("%d", ls.value))
		b.WriteString(fmt.Sprintf("  %s %s\n", num, theme.MutedStyle.Render(ls.label)))
	}

	// Listening section
	b.WriteString("\n")
	b.WriteString(theme.TitleStyle.Render("LISTENING ("+s.Listening.TimeRange+")"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  Plays:  %s  |  Streak: %s days  |  Best: %s days\n",
		lipgloss.NewStyle().Foreground(theme.ColorHotPink).Bold(true).
			Render(fmt.Sprintf("%d", s.Listening.TotalPlays)),
		lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true).
			Render(fmt.Sprintf("%d", s.Listening.CurrentStreak)),
		lipgloss.NewStyle().Foreground(theme.ColorChromeYellow).Bold(true).
			Render(fmt.Sprintf("%d", s.Listening.LongestStreak)),
	))

	// Top tracks as simple bar chart
	if len(s.Listening.TopTracks) > 0 {
		b.WriteString("\n")
		b.WriteString(theme.SubtitleStyle.Render("Top Tracks"))
		b.WriteString("\n")
		maxPlays := s.Listening.TopTracks[0].PlayCount
		for i, t := range s.Listening.TopTracks {
			if i >= 10 {
				break
			}
			barWidth := 20
			if maxPlays > 0 {
				barWidth = t.PlayCount * 20 / maxPlays
			}
			if barWidth < 1 {
				barWidth = 1
			}
			bar := lipgloss.NewStyle().Foreground(theme.ColorHotPink).
				Render(strings.Repeat("█", barWidth))
			name := lipgloss.NewStyle().Width(25).Render(t.Title)
			b.WriteString(fmt.Sprintf("  %s %s %s\n",
				name, bar,
				theme.MutedStyle.Render(fmt.Sprintf("%d", t.PlayCount)),
			))
		}
	}

	// Top artists
	if len(s.Listening.TopArtists) > 0 {
		b.WriteString("\n")
		b.WriteString(theme.SubtitleStyle.Render("Top Artists"))
		b.WriteString("\n")
		maxPlays := s.Listening.TopArtists[0].PlayCount
		for i, a := range s.Listening.TopArtists {
			if i >= 10 {
				break
			}
			barWidth := 20
			if maxPlays > 0 {
				barWidth = a.PlayCount * 20 / maxPlays
			}
			if barWidth < 1 {
				barWidth = 1
			}
			bar := lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).
				Render(strings.Repeat("█", barWidth))
			name := lipgloss.NewStyle().Width(25).Render(a.Name)
			b.WriteString(fmt.Sprintf("  %s %s %s\n",
				name, bar,
				theme.MutedStyle.Render(fmt.Sprintf("%d", a.PlayCount)),
			))
		}
	}

	return b.String()
}

func (v *StatsView) Title() string          { return "Stats" }
func (v *StatsView) ShortHelp() []key.Binding { return []key.Binding{Keys.Refresh} }

// ---------------------------------------------------------------------------
// Playlist Detail View
// ---------------------------------------------------------------------------

type PlaylistDetailView struct {
	viewSize
	client   *api.Client
	id       int64
	playlist *models.PlaylistDetail
	search   components.SearchInput
	loading  bool
	spinner  spinner.Model
	table    components.DataTable
	page     int
	query    string
	err      error
	ready    bool
}

func NewPlaylistDetailView(client *api.Client, id int64) *PlaylistDetailView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return &PlaylistDetailView{
		client:  client,
		id:      id,
		search:  components.NewSearchInput("Search playlist..."),
		loading: true,
		spinner: s,
		page:    1,
	}
}

type playlistDetailMsg struct {
	playlist *models.PlaylistDetail
	err      error
}

func (v *PlaylistDetailView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.load())
}

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
		if v.playlist != nil && len(v.playlist.Tracks) > 0 {
			// Client-side filter if the API doesn't support search
			if v.query != "" {
				q := strings.ToLower(v.query)
				filtered := v.playlist.Tracks[:0]
				for _, pt := range v.playlist.Tracks {
					if strings.Contains(strings.ToLower(pt.Track.Title), q) ||
						strings.Contains(strings.ToLower(pt.Track.Artist.Name), q) ||
						strings.Contains(strings.ToLower(pt.Track.Album.Title), q) {
						filtered = append(filtered, pt)
					}
				}
				v.playlist.Tracks = filtered
			}
			v.buildTable()
		}
		return v, nil
	case components.RowSelectMsg:
		// Queue all playlist tracks and start from selected
		if v.playlist != nil && len(msg.Row) > 0 {
			var pos int
			fmt.Sscanf(msg.Row[0], "%d", &pos)
			selectedIndex := 0
			tracks := make([]player.Track, len(v.playlist.Tracks))
			for i, pt := range v.playlist.Tracks {
				tracks[i] = player.Track{
					ID: pt.Track.ID, Title: pt.Track.Title,
					Artist: pt.Track.Artist.Name, Album: pt.Track.Album.Title,
					Duration: pt.Track.Duration,
				}
				if pt.Position == pos {
					selectedIndex = i
				}
			}
			return v, func() tea.Msg {
				return PlayQueueMsg{Tracks: tracks, Index: selectedIndex}
			}
		}
	case components.PageChangeMsg:
		v.page = msg.Page
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.load())
	case components.SearchQueryMsg:
		v.query = msg.Query
		v.page = 1
		v.loading = true
		return v, tea.Batch(v.spinner.Tick, v.load())
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	if v.search.Focused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				v.search = v.search.Blur()
				return v, nil
			}
			if keyMsg.String() == "enter" {
				v.query = v.search.Value()
				v.page = 1
				v.loading = true
				v.search = v.search.Blur()
				return v, tea.Batch(v.spinner.Tick, v.load())
			}
		}
		var cmd tea.Cmd
		v.search, cmd = v.search.Update(msg)
		return v, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "/":
			v.search = v.search.Focus()
			return v, nil
		case "d":
			if v.playlist != nil && v.ready {
				row := v.table.SelectedRow()
				if row != nil {
					var pos int
					fmt.Sscanf(row[0], "%d", &pos)
					for _, pt := range v.playlist.Tracks {
						if pt.Position == pos {
							plID := v.playlist.ID
							ptID := pt.PlaylistTrackID
							client := v.client
							return v, func() tea.Msg {
								svc := api.NewPlaylistService(client)
								err := svc.RemoveTrack(context.Background(), plID, ptID)
								if err != nil {
									return components.ToastMsg{Message: "Remove failed: " + err.Error(), IsError: true}
								}
								// Reload the playlist to reflect the removal
								pl, err := svc.Get(context.Background(), plID, "", 1, 100)
								if err != nil {
									return components.ToastMsg{Message: "Track removed (refresh to see changes)"}
								}
								return playlistDetailMsg{playlist: pl}
							}
						}
					}
				}
			}
		}
	}

	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *PlaylistDetailView) buildTable() {
	columns := []table.Column{
		{Title: "#", Width: 4},
		{Title: "Title", Width: 25},
		{Title: "Artist", Width: 20},
		{Title: "Album", Width: 20},
		{Title: "Duration", Width: 8},
	}
	rows := make([]table.Row, len(v.playlist.Tracks))
	for i, pt := range v.playlist.Tracks {
		dur := fmt.Sprintf("%d:%02d", int(pt.Track.Duration)/60, int(pt.Track.Duration)%60)
		rows[i] = table.Row{
			fmt.Sprintf("%d", pt.Position),
			pt.Track.Title,
			pt.Track.Artist.Name,
			pt.Track.Album.Title,
			dur,
		}
	}
	v.table = components.NewDataTable(columns, rows, v.playlist.Pagination, v.tHeight(), v.tWidth())
	v.ready = true
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
	info := fmt.Sprintf("%s tracks  |  %s",
		lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true).Render(fmt.Sprintf("%d", pl.TracksCount)),
		theme.MutedStyle.Render(dur),
	)
	b.WriteString(info)
	b.WriteString("\n")
	b.WriteString(v.search.View())
	b.WriteString("\n")

	if v.ready {
		b.WriteString(v.table.View())
	}

	return b.String()
}

func (v *PlaylistDetailView) Title() string            { return "Playlist" }
func (v *PlaylistDetailView) HasFocusedInput() bool     { return v.search.Focused }
func (v *PlaylistDetailView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Search, Keys.Enter, Keys.Delete, Keys.NextPage, Keys.PrevPage, Keys.Back}
}

// ---------------------------------------------------------------------------
// Public Radio List View
// ---------------------------------------------------------------------------

type PublicRadioListView struct {
	viewSize
	client   *api.Client
	player   *player.Player
	table    components.DataTable
	loading  bool
	spinner  spinner.Model
	stations []models.PublicRadioStation
	err      error
	ready    bool
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

func (v *PublicRadioListView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.load())
}

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
		v.buildTable()
		return v, nil
	case components.RowSelectMsg:
		if len(msg.Row) > 0 {
			// Row[0] is the slug
			slug := msg.Row[0]
			return v, func() tea.Msg {
				return NavigateMsg{View: NewPublicRadioDetailView(v.client, v.player, slug)}
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	if v.ready {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *PublicRadioListView) buildTable() {
	columns := []table.Column{
		{Title: "Slug", Width: 20},
		{Title: "Name", Width: 25},
		{Title: "Status", Width: 10},
		{Title: "Listeners", Width: 10},
		{Title: "Now Playing", Width: 30},
	}
	rows := make([]table.Row, len(v.stations))
	for i, s := range v.stations {
		nowPlaying := "-"
		if s.CurrentTrack != nil {
			nowPlaying = s.CurrentTrack.Title + " - " + s.CurrentTrack.Artist.Name
		}
		rows[i] = table.Row{
			s.Slug,
			s.Name,
			theme.StatusDot(s.Status) + " " + s.Status,
			fmt.Sprintf("%d", s.ListenerCount),
			nowPlaying,
		}
	}
	v.table = components.NewDataTable(columns, rows, models.Pagination{}, v.tHeight(), v.tWidth())
	v.ready = true
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
	} else if v.ready {
		b.WriteString(v.table.View())
	}
	return b.String()
}

func (v *PublicRadioListView) Title() string          { return "Live Radio" }
func (v *PublicRadioListView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Enter, Keys.Back}
}

// ---------------------------------------------------------------------------
// Public Radio Detail View
// ---------------------------------------------------------------------------

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
	return tea.Tick(10*time.Second, func(time.Time) tea.Msg {
		return publicRadioRefreshMsg{}
	})
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
		// Auto-refresh to update current track
		return v, tea.Batch(v.load(), v.refreshTimer())
	case tea.KeyMsg:
		if msg.String() == "enter" && v.station != nil && v.station.ListenURL != "" {
			// If already playing this station, stop it
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
			return v, func() tea.Msg {
				return PlayStreamMsg{
					URL:    url,
					Title:  trackTitle,
					Artist: trackArtist,
					Album:  s.Name,
				}
			}
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

	// Compact sun as station header
	b.WriteString(theme.RenderSunCompact())
	b.WriteString("\n\n")

	// Status + listeners
	statusStyle := lipgloss.NewStyle().
		Foreground(theme.ColorDeepNavy).
		Background(theme.ColorGreen).
		Bold(true).
		Padding(0, 1)
	if s.Status != "active" {
		statusStyle = statusStyle.Background(theme.ColorRed)
	}
	b.WriteString(statusStyle.Render(strings.ToUpper(s.Status)))
	b.WriteString(theme.MutedStyle.Render(fmt.Sprintf("  %d listeners", s.ListenerCount)))
	b.WriteString("\n\n")

	// Station name -- big gradient title
	b.WriteString(theme.GradientText(strings.ToUpper(s.Name)))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(theme.ColorGridPurple).
		Render(strings.Repeat("~", len(s.Name)+4)))
	b.WriteString("\n\n")

	// Now playing
	if s.CurrentTrack != nil {
		b.WriteString(theme.MutedStyle.Render("Now Playing"))
		b.WriteString("\n")
		nowIcon := lipgloss.NewStyle().Foreground(theme.ColorHotPink).Bold(true).Render("  >> ")
		trackName := lipgloss.NewStyle().Foreground(theme.ColorGhostWhite).Bold(true).
			Render(s.CurrentTrack.Title)
		artistName := lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).
			Render(s.CurrentTrack.Artist.Name)
		b.WriteString(nowIcon + trackName + theme.MutedStyle.Render(" - ") + artistName)
		b.WriteString("\n\n")
	}

	// Stream URL
	b.WriteString(theme.MutedStyle.Render("Stream URL"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(theme.ColorNeonPurple).Render("  " + s.ListenURL))
	b.WriteString("\n\n")


	// Action button
	if v.isPlayingThisStation() {
		stopBtn := lipgloss.NewStyle().
			Foreground(theme.ColorDeepNavy).
			Background(theme.ColorRed).
			Bold(true).
			Padding(0, 2).
			Render("ENTER: Stop Listening")
		b.WriteString(stopBtn)
	} else {
		playBtn := lipgloss.NewStyle().
			Foreground(theme.ColorDeepNavy).
			Background(theme.ColorHotPink).
			Bold(true).
			Padding(0, 2).
			Render("ENTER: Start Listening")
		b.WriteString(playBtn)
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

func (v *PublicRadioDetailView) Title() string          { return "Station" }
func (v *PublicRadioDetailView) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Enter, Keys.Back}
}
