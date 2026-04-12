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
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

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

func (v *DashboardView) Refresh() tea.Cmd {
	v.loading = true
	v.err = nil
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
	b.WriteString(theme.RenderSynthwaveSun(70))
	b.WriteString("\n")
	b.WriteString(theme.RenderSmallLogo())
	b.WriteString("\n\n")

	if v.profile != nil {
		nameStyle := lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true)
		b.WriteString(nameStyle.Render(v.profile.Name))
		b.WriteString(theme.MutedStyle.Render("  " + v.profile.EmailAddress))
		b.WriteString("\n\n")

		statBox := func(label string, count int, color lipgloss.Color) string {
			num := lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("%d", count))
			lbl := theme.MutedStyle.Render(label)
			return lipgloss.JoinVertical(lipgloss.Center, num, lbl)
		}

		statsRow := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(14).Align(lipgloss.Center).Render(statBox("Artists", v.profile.Stats.ArtistsCount, theme.ColorHotPink)),
			lipgloss.NewStyle().Width(14).Align(lipgloss.Center).Render(statBox("Albums", v.profile.Stats.AlbumsCount, theme.ColorNeonCyan)),
			lipgloss.NewStyle().Width(14).Align(lipgloss.Center).Render(statBox("Tracks", v.profile.Stats.TracksCount, theme.ColorNeonPurple)),
			lipgloss.NewStyle().Width(14).Align(lipgloss.Center).Render(statBox("Playlists", v.profile.Stats.PlaylistsCount, theme.ColorChromeYellow)),
		)

		b.WriteString(lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorGridPurple).Padding(1, 2).Render(statsRow))
	}

	if v.stats != nil && v.stats.Listening.TotalPlays > 0 {
		b.WriteString("\n\n")
		b.WriteString(theme.SubtitleStyle.Render("Listening This Month"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s plays  |  %s streak  |  %s best",
			lipgloss.NewStyle().Foreground(theme.ColorHotPink).Bold(true).Render(fmt.Sprintf("%d", v.stats.Listening.TotalPlays)),
			lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true).Render(fmt.Sprintf("%d days", v.stats.Listening.CurrentStreak)),
			lipgloss.NewStyle().Foreground(theme.ColorChromeYellow).Bold(true).Render(fmt.Sprintf("%d days", v.stats.Listening.LongestStreak)),
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

func (v *DashboardView) Title() string            { return "Dashboard" }
func (v *DashboardView) ShortHelp() []key.Binding { return []key.Binding{Keys.Refresh} }

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

func (v *StatsView) Refresh() tea.Cmd {
	v.loading = true
	v.err = nil
	return v.Init()
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
	b.WriteString(theme.TitleStyle.Render("LIBRARY"))
	b.WriteString("\n\n")
	libStats := []struct {
		label string
		value int
		color lipgloss.Color
	}{{"Artists", s.Library.ArtistsCount, theme.ColorHotPink}, {"Albums", s.Library.AlbumsCount, theme.ColorNeonCyan}, {"Tracks", s.Library.TracksCount, theme.ColorNeonPurple}, {"Playlists", s.Library.PlaylistsCount, theme.ColorChromeYellow}}
	for _, ls := range libStats {
		num := lipgloss.NewStyle().Foreground(ls.color).Bold(true).Width(8).Align(lipgloss.Right).Render(fmt.Sprintf("%d", ls.value))
		b.WriteString(fmt.Sprintf("  %s %s\n", num, theme.MutedStyle.Render(ls.label)))
	}
	b.WriteString("\n")
	b.WriteString(theme.TitleStyle.Render("LISTENING (" + s.Listening.TimeRange + ")"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  Plays:  %s  |  Streak: %s days  |  Best: %s days\n",
		lipgloss.NewStyle().Foreground(theme.ColorHotPink).Bold(true).Render(fmt.Sprintf("%d", s.Listening.TotalPlays)),
		lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true).Render(fmt.Sprintf("%d", s.Listening.CurrentStreak)),
		lipgloss.NewStyle().Foreground(theme.ColorChromeYellow).Bold(true).Render(fmt.Sprintf("%d", s.Listening.LongestStreak)),
	))
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
			bar := lipgloss.NewStyle().Foreground(theme.ColorHotPink).Render(strings.Repeat("█", barWidth))
			name := lipgloss.NewStyle().Width(25).Render(t.Title)
			b.WriteString(fmt.Sprintf("  %s %s %s\n", name, bar, theme.MutedStyle.Render(fmt.Sprintf("%d", t.PlayCount))))
		}
	}
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
			bar := lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Render(strings.Repeat("█", barWidth))
			name := lipgloss.NewStyle().Width(25).Render(a.Name)
			b.WriteString(fmt.Sprintf("  %s %s %s\n", name, bar, theme.MutedStyle.Render(fmt.Sprintf("%d", a.PlayCount))))
		}
	}
	return b.String()
}

func (v *StatsView) Title() string            { return "Stats" }
func (v *StatsView) ShortHelp() []key.Binding { return []key.Binding{Keys.Refresh} }
