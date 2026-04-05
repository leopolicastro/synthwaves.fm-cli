package theme

import "github.com/charmbracelet/lipgloss"

// Synthwave color palette
var (
	ColorHotPink       = lipgloss.Color("#ff2d95")
	ColorNeonCyan      = lipgloss.Color("#00fff2")
	ColorNeonPurple    = lipgloss.Color("#b026ff")
	ColorElectricBlue  = lipgloss.Color("#0066ff")
	ColorSunsetOrange  = lipgloss.Color("#ff6b35")
	ColorChromeYellow  = lipgloss.Color("#ffd700")
	ColorDeepNavy      = lipgloss.Color("#0a0a1a")
	ColorDarkPurple    = lipgloss.Color("#1a0a2e")
	ColorGridPurple    = lipgloss.Color("#2d1b4e")
	ColorMutedLavender = lipgloss.Color("#6c6c8a")
	ColorGhostWhite    = lipgloss.Color("#e8e8f0")
	ColorGreen         = lipgloss.Color("#44ff44")
	ColorRed           = lipgloss.Color("#ff4444")
)

// Text styles
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorHotPink).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorNeonCyan).
			Bold(true)

	AccentStyle = lipgloss.NewStyle().
			Foreground(ColorNeonPurple)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMutedLavender)

	GhostStyle = lipgloss.NewStyle().
			Foreground(ColorGhostWhite)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorSunsetOrange)

	GoldStyle = lipgloss.NewStyle().
			Foreground(ColorChromeYellow)
)

// Layout styles
var (
	SidebarStyle = lipgloss.NewStyle().
			Width(22).
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).
			BorderForeground(ColorGridPurple).
			Padding(1, 1)

	ContentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMutedLavender).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(ColorGridPurple).
			Padding(0, 1)
)

// Nav item styles
var (
	NavItemStyle = lipgloss.NewStyle().
			Foreground(ColorMutedLavender).
			PaddingLeft(2)

	NavActiveStyle = lipgloss.NewStyle().
			Foreground(ColorNeonCyan).
			Bold(true).
			PaddingLeft(1)

	NavActivePrefix = lipgloss.NewStyle().
			Foreground(ColorHotPink).
			Bold(true).
			Render(">")
)

// StatusDot returns a colored dot based on status.
func StatusDot(status string) string {
	switch status {
	case "active":
		return SuccessStyle.Render("*")
	case "stopped":
		return ErrorStyle.Render("*")
	default:
		return WarningStyle.Render("*")
	}
}

// GradientText renders text with a pink-to-cyan gradient.
func GradientText(text string) string {
	colors := []lipgloss.Color{
		"#ff2d95", "#e826a8", "#d11fbb", "#ba18ce",
		"#a311e1", "#8c0af4", "#6617f4", "#4024f4",
		"#1a31f4", "#00fff2",
	}
	if len(text) == 0 {
		return ""
	}
	result := ""
	for i, ch := range text {
		idx := 0
		if len(text) > 1 {
			idx = i * (len(colors) - 1) / (len(text) - 1)
		}
		if idx >= len(colors) {
			idx = len(colors) - 1
		}
		result += lipgloss.NewStyle().Foreground(colors[idx]).Render(string(ch))
	}
	return result
}

// RenderSmallLogo returns the styled app name.
func RenderSmallLogo() string {
	return TitleStyle.Render("SYNTHWAVES") +
		AccentStyle.Render(".") +
		SubtitleStyle.Render("FM")
}
