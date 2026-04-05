package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/player"
	"github.com/leo/synthwaves-cli/internal/ui/components"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

// View is the interface all TUI views implement.
type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View() string
	Title() string
	ShortHelp() []key.Binding
}

// InputFocuser is optionally implemented by views with text inputs.
type InputFocuser interface {
	HasFocusedInput() bool
}

// ---------------------------------------------------------------------------
// Messages
// ---------------------------------------------------------------------------

type NavigateMsg struct{ View View }
type ReplaceViewMsg struct{ View View }
type BackMsg struct{}

// PlayTrackMsg plays a single track (queues it alone).
type PlayTrackMsg struct {
	TrackID  int64
	Title    string
	Artist   string
	Album    string
	Duration float64
}

// PlayQueueMsg queues a list of tracks and starts playing at Index.
type PlayQueueMsg struct {
	Tracks []player.Track
	Index  int
}

// PlayStreamMsg plays a direct stream URL (e.g., Icecast radio).
type PlayStreamMsg struct {
	URL      string
	Title    string
	Artist   string
	Album    string
}

type TrackPlayingMsg struct{}
type TrackStoppedMsg struct{}
type TrackAdvancedMsg struct{ Track player.Track }
type TrackErrorMsg struct{ Err error }
type PlayerTickMsg struct{}

// ---------------------------------------------------------------------------
// Nav sections
// ---------------------------------------------------------------------------

type NavSection string

const (
	SectionDashboard NavSection = "dashboard"
	SectionArtists   NavSection = "artists"
	SectionAlbums    NavSection = "albums"
	SectionTracks    NavSection = "tracks"
	SectionPlaylists NavSection = "playlists"
	SectionFavorites NavSection = "favorites"
	SectionHistory   NavSection = "history"
	SectionLiveRadio NavSection = "live_radio"
	SectionRadio     NavSection = "radio"
	SectionSearch    NavSection = "search"
	SectionStats     NavSection = "stats"
)

// ---------------------------------------------------------------------------
// App
// ---------------------------------------------------------------------------

type App struct {
	Client *api.Client
	Player *player.Player

	nav         components.Nav
	statusBar   components.StatusBar
	toast       components.Toast
	progressBar progress.Model

	activeView     View
	viewStack      []View
	sidebarFocused bool

	width  int
	height int
	ready  bool
}

func NewApp(client *api.Client) *App {
	navItems := []components.NavItem{
		{Label: "Dashboard", Key: string(SectionDashboard)},
		{Label: "Artists", Key: string(SectionArtists)},
		{Label: "Albums", Key: string(SectionAlbums)},
		{Label: "Tracks", Key: string(SectionTracks)},
		{Label: "Playlists", Key: string(SectionPlaylists)},
		{Label: "Favorites", Key: string(SectionFavorites)},
		{Label: "History", Key: string(SectionHistory)},
		{Label: "Live Radio", Key: string(SectionLiveRadio)},
		{Label: "My Radio", Key: string(SectionRadio)},
		{Label: "Search", Key: string(SectionSearch)},
		{Label: "Stats", Key: string(SectionStats)},
	}

	pb := progress.New(
		progress.WithScaledGradient("#ff2d95", "#00fff2"),
		progress.WithoutPercentage(),
	)

	return &App{
		Client:         client,
		Player:         player.New(),
		nav:            components.NewNav(navItems),
		statusBar:      components.NewStatusBar(),
		toast:          components.NewToast(),
		progressBar:    pb,
		sidebarFocused: true,
	}
}

func (a *App) Init() tea.Cmd {
	return nil
}

// isInputActive returns true if a text input in the current view has focus.
func (a *App) isInputActive() bool {
	if a.sidebarFocused {
		return false
	}
	if f, ok := a.activeView.(InputFocuser); ok {
		return f.HasFocusedInput()
	}
	return false
}

func (a *App) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return PlayerTickMsg{}
	})
}

func (a *App) waitForPlayerEvent() tea.Cmd {
	p := a.Player
	return func() tea.Msg {
		event := p.WaitForEvent()
		switch event.Event {
		case player.EventAdvance:
			return TrackAdvancedMsg{Track: event.Track}
		default:
			return TrackStoppedMsg{}
		}
	}
}

// playTrack fetches the stream URL and starts playback.
func (a *App) playTrack(track player.Track) tea.Cmd {
	client := a.Client
	p := a.Player
	return func() tea.Msg {
		svc := api.NewTrackService(client)
		// Fill in missing metadata (e.g. Favorites/Search lack Duration)
		if track.Duration == 0 && track.ID > 0 {
			if full, err := svc.Get(context.Background(), track.ID); err == nil {
				track.Duration = full.Duration
				if track.Artist == "" {
					track.Artist = full.Artist.Name
				}
				if track.Album == "" {
					track.Album = full.Album.Title
				}
			}
		}
		info, err := svc.Stream(context.Background(), track.ID)
		if err != nil {
			return TrackErrorMsg{Err: err}
		}
		if err := p.Play(info.URL, track); err != nil {
			return TrackErrorMsg{Err: err}
		}
		return TrackPlayingMsg{}
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.nav.Height = msg.Height - 4
		a.ready = true
		if a.activeView != nil {
			var cmd tea.Cmd
			a.activeView, cmd = a.activeView.Update(msg)
			return a, cmd
		}
		return a, nil

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			if msg.X < 22 {
				a.sidebarFocused = true
				a.nav.Focused = true
			} else {
				a.sidebarFocused = false
				a.nav.Focused = false
			}
		}

	case tea.KeyMsg:
		switch {
		// Quit
		case msg.String() == "ctrl+c" || msg.String() == "ctrl+q":
			a.Player.Cleanup()
			return a, tea.Quit
		case msg.String() == "q" && a.sidebarFocused:
			a.Player.Cleanup()
			return a, tea.Quit

		// Player controls (only when player active and no text input focused)
		case key.Matches(msg, Keys.Stop) && !a.isInputActive():
			if a.Player.IsPlaying() {
				a.Player.Stop()
				a.Player.Queue().Clear()
				return a, nil
			}
		case key.Matches(msg, Keys.PlayPause) && !a.isInputActive():
			if a.Player.IsPlaying() {
				a.Player.TogglePause()
				return a, nil
			}
		case key.Matches(msg, Keys.NextTrack) && !a.isInputActive():
			if a.Player.IsPlaying() {
				a.Player.SkipNext()
				return a, nil
			}
		case key.Matches(msg, Keys.PrevTrack) && !a.isInputActive():
			if a.Player.IsPlaying() {
				prev := a.Player.SkipPrevious()
				if prev != nil {
					return a, a.playTrack(*prev)
				}
				return a, nil
			}
		case key.Matches(msg, Keys.Shuffle) && !a.isInputActive():
			if a.Player.IsPlaying() {
				q := a.Player.Queue()
				if q.Mode() == player.QueueShuffle {
					q.SetMode(player.QueueNormal)
				} else {
					q.SetMode(player.QueueShuffle)
				}
				return a, nil
			}
		case key.Matches(msg, Keys.Repeat) && !a.isInputActive():
			if a.Player.IsPlaying() {
				q := a.Player.Queue()
				if q.Mode() == player.QueueRepeatOne {
					q.SetMode(player.QueueNormal)
				} else {
					q.SetMode(player.QueueRepeatOne)
				}
				return a, nil
			}
		case (msg.String() == "+" || msg.String() == "=") && !a.isInputActive():
			if a.Player.IsPlaying() {
				a.Player.AdjustVolume(0.5)
				return a, nil
			}
		case msg.String() == "-" && !a.isInputActive():
			if a.Player.IsPlaying() {
				a.Player.AdjustVolume(-0.5)
				return a, nil
			}
		case (msg.String() == "." || msg.String() == ">") && !a.isInputActive():
			if a.Player.IsPlaying() {
				a.Player.Seek(5 * time.Second)
				return a, nil
			}
		case (msg.String() == "," || msg.String() == "<") && !a.isInputActive():
			if a.Player.IsPlaying() {
				a.Player.Seek(-5 * time.Second)
				return a, nil
			}

		// Navigation
		case key.Matches(msg, Keys.Tab):
			a.sidebarFocused = !a.sidebarFocused
			a.nav.Focused = a.sidebarFocused
			return a, nil
		case key.Matches(msg, Keys.Back) && !a.isInputActive():
			if !a.sidebarFocused && len(a.viewStack) > 0 {
				a.activeView = a.viewStack[len(a.viewStack)-1]
				a.viewStack = a.viewStack[:len(a.viewStack)-1]
				return a, nil
			}
		}

	// --------------- Playback messages ---------------

	case PlayStreamMsg:
		// Direct URL streaming (radio) -- no TrackService.Stream() needed
		p := a.Player
		track := player.Track{
			Title: msg.Title, Artist: msg.Artist,
			Album: msg.Album, Duration: 0,
		}
		a.Player.Queue().Clear()
		return a, func() tea.Msg {
			if err := p.Play(msg.URL, track); err != nil {
				return TrackErrorMsg{Err: err}
			}
			return TrackPlayingMsg{}
		}

	case PlayTrackMsg:
		track := player.Track{
			ID: msg.TrackID, Title: msg.Title,
			Artist: msg.Artist, Album: msg.Album,
			Duration: msg.Duration,
		}
		a.Player.Queue().SetTracks([]player.Track{track}, 0)
		return a, a.playTrack(track)

	case PlayQueueMsg:
		a.Player.Queue().SetTracks(msg.Tracks, msg.Index)
		track := a.Player.Queue().Current()
		if track == nil {
			return a, nil
		}
		return a, a.playTrack(*track)

	case TrackPlayingMsg:
		return a, tea.Batch(a.tickCmd(), a.waitForPlayerEvent())

	case TrackAdvancedMsg:
		// Auto-advance: fetch stream URL for the next track
		return a, a.playTrack(msg.Track)

	case TrackStoppedMsg:
		return a, nil

	case TrackErrorMsg:
		return a, func() tea.Msg {
			return components.ToastMsg{Message: "Playback error: " + msg.Err.Error(), IsError: true}
		}

	case PlayerTickMsg:
		if a.Player.GetState() == player.Playing {
			return a, a.tickCmd()
		}
		return a, nil

	// --------------- Other messages ---------------

	case NavigateMsg:
		if a.activeView != nil {
			a.viewStack = append(a.viewStack, a.activeView)
		}
		a.activeView = msg.View
		a.sidebarFocused = false
		a.nav.Focused = false
		return a, a.activeView.Init()

	case ReplaceViewMsg:
		a.activeView = msg.View
		a.viewStack = nil
		a.sidebarFocused = false
		a.nav.Focused = false
		return a, a.activeView.Init()

	case BackMsg:
		if len(a.viewStack) > 0 {
			a.activeView = a.viewStack[len(a.viewStack)-1]
			a.viewStack = a.viewStack[:len(a.viewStack)-1]
		}
		return a, nil

	case components.NavSelectMsg:
		view := a.createViewForSection(NavSection(msg.Key))
		if view != nil {
			a.activeView = view
			a.viewStack = nil
			a.sidebarFocused = false
			a.nav.Focused = false
			return a, a.activeView.Init()
		}

	case components.ToastMsg:
		a.toast, _ = a.toast.Update(msg)
		return a, nil
	}

	// Update toast
	var toastCmd tea.Cmd
	a.toast, toastCmd = a.toast.Update(msg)
	if toastCmd != nil {
		cmds = append(cmds, toastCmd)
	}

	// Route to sidebar or active view
	if a.sidebarFocused {
		var navCmd tea.Cmd
		a.nav, navCmd = a.nav.Update(msg)
		if navCmd != nil {
			cmds = append(cmds, navCmd)
		}
	} else if a.activeView != nil {
		var viewCmd tea.Cmd
		a.activeView, viewCmd = a.activeView.Update(msg)
		if viewCmd != nil {
			cmds = append(cmds, viewCmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	sidebar := a.nav.View()

	contentWidth := a.width - lipgloss.Width(sidebar) - 2
	var content string
	if a.activeView != nil {
		content = a.activeView.View()
	} else {
		content = a.renderWelcome()
	}
	content = theme.ContentStyle.Width(contentWidth).Render(content)

	main := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	toastView := a.toast.View()
	if toastView != "" {
		main = lipgloss.JoinVertical(lipgloss.Left, toastView, main)
	}

	nowPlaying := a.renderNowPlaying()

	a.statusBar.Width = a.width
	a.statusBar.Breadcrumbs = a.getBreadcrumbs()
	a.statusBar.KeyHints = a.getKeyHints()
	bar := a.statusBar.View()

	if nowPlaying != "" {
		return lipgloss.JoinVertical(lipgloss.Left, main, nowPlaying, bar)
	}
	return lipgloss.JoinVertical(lipgloss.Left, main, bar)
}

func (a *App) renderNowPlaying() string {
	track := a.Player.Current()
	if track == nil {
		return ""
	}

	state := a.Player.GetState()

	// State icon
	var icon string
	switch state {
	case player.Playing:
		icon = lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true).Render(">> ")
	case player.Paused:
		icon = lipgloss.NewStyle().Foreground(theme.ColorChromeYellow).Bold(true).Render("|| ")
	default:
		return ""
	}

	// Track info
	title := lipgloss.NewStyle().Foreground(theme.ColorGhostWhite).Bold(true).Render(track.Title)
	sep := theme.MutedStyle.Render(" - ")
	artist := lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(track.Artist)

	// Queue position
	q := a.Player.Queue()
	queueInfo := ""
	if q.Len() > 1 {
		queueInfo = theme.MutedStyle.Render(fmt.Sprintf("  [%d/%d]", q.Index(), q.Len()))
	}

	var line2 string

	// Visualizer
	viz := components.RenderVisualizer(a.Player.Tap())

	// Volume indicator
	vol := a.Player.Volume()
	volPct := (vol + 5.0) / 6.0 // map [-5, 1] to [0, 1]
	if volPct < 0 {
		volPct = 0
	}
	if volPct > 1 {
		volPct = 1
	}
	volBars := int(volPct * 10)
	volStr := ""
	for i := 0; i < 10; i++ {
		if i < volBars {
			volStr += "|"
		} else {
			volStr += " "
		}
	}
	volView := theme.MutedStyle.Render(" vol[") +
		lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(volStr) +
		theme.MutedStyle.Render("]")

	if track.Duration == 0 {
		// Live stream -- no progress bar, show LIVE badge
		liveBadge := lipgloss.NewStyle().
			Foreground(theme.ColorDeepNavy).
			Background(theme.ColorRed).
			Bold(true).
			Padding(0, 1).
			Render("LIVE")
		line1 := icon + title + sep + artist + "  " + liveBadge + queueInfo
		if viz != "" {
			line1 += "  " + viz
		}
		hints := theme.MutedStyle.Render("  space:pause  +/-:vol  x:stop")
		line2 = hints + volView

		content := line1 + "\n" + line2
		return lipgloss.NewStyle().
			Width(a.width).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(theme.ColorHotPink).
			Padding(0, 1).
			Render(content)
	}

	// Track mode -- progress bar + time
	elapsed := a.Player.Elapsed()
	elapsedSec := int(elapsed.Seconds())
	totalSec := int(track.Duration)

	timeStr := fmt.Sprintf("%d:%02d / %d:%02d",
		elapsedSec/60, elapsedSec%60,
		totalSec/60, totalSec%60)

	pct := 0.0
	if track.Duration > 0 {
		pct = elapsed.Seconds() / track.Duration
		if pct > 1.0 {
			pct = 1.0
		}
	}

	barWidth := a.width - 50
	if barWidth < 15 {
		barWidth = 15
	}
	if barWidth > 60 {
		barWidth = 60
	}
	a.progressBar.Width = barWidth
	bar := a.progressBar.ViewAs(pct)

	timeView := theme.MutedStyle.Render(" " + timeStr)

	// Mode indicators
	modes := ""
	switch q.Mode() {
	case player.QueueShuffle:
		modes = lipgloss.NewStyle().Foreground(theme.ColorNeonPurple).Bold(true).Render("  [S]")
	case player.QueueRepeatOne:
		modes = lipgloss.NewStyle().Foreground(theme.ColorChromeYellow).Bold(true).Render("  [R1]")
	}

	// Line 1: icon + title + artist + queue pos + visualizer
	line1 := icon + title + sep + artist + queueInfo + modes
	if viz != "" {
		line1 += "  " + viz
	}

	// Line 2: progress bar + time + volume + hints
	hints := theme.MutedStyle.Render("  space:pause  ]/[:skip  ./:seek  +/-:vol  s:shuf  R:rpt  x:stop")
	line2 = bar + timeView + volView + hints

	content := line1 + "\n" + line2

	return lipgloss.NewStyle().
		Width(a.width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(theme.ColorHotPink).
		Padding(0, 1).
		Render(content)
}

func (a *App) renderWelcome() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		"",
		theme.RenderSmallLogo(),
		"",
		theme.MutedStyle.Render("Select a section from the sidebar to get started."),
		theme.MutedStyle.Render("Press Tab to switch between sidebar and content."),
	)
}

func (a *App) getBreadcrumbs() []string {
	var crumbs []string
	for _, v := range a.viewStack {
		crumbs = append(crumbs, v.Title())
	}
	if a.activeView != nil {
		crumbs = append(crumbs, a.activeView.Title())
	}
	return crumbs
}

func (a *App) getKeyHints() []key.Binding {
	hints := []key.Binding{Keys.Tab, Keys.Back, Keys.Quit}
	if a.activeView != nil {
		hints = append(a.activeView.ShortHelp(), hints...)
	}
	return hints
}

func (a *App) createViewForSection(section NavSection) View {
	switch section {
	case SectionDashboard:
		return NewDashboardView(a.Client)
	case SectionArtists:
		return NewArtistsView(a.Client)
	case SectionAlbums:
		return NewAlbumsView(a.Client)
	case SectionTracks:
		return NewTracksView(a.Client)
	case SectionPlaylists:
		return NewPlaylistsView(a.Client)
	case SectionFavorites:
		return NewFavoritesView(a.Client)
	case SectionHistory:
		return NewHistoryView(a.Client)
	case SectionLiveRadio:
		return NewPublicRadioListView(a.Client, a.Player)
	case SectionRadio:
		return NewRadioView(a.Client)
	case SectionSearch:
		return NewSearchView(a.Client)
	case SectionStats:
		return NewStatsView(a.Client)
	default:
		return nil
	}
}
