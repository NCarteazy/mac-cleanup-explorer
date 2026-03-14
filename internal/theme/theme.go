package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

// Tokyo Night color palette
const (
	BgColor        = "#1a1b26"
	PrimaryColor   = "#7dcfff"
	SecondaryColor = "#bb9af7"
	SuccessColor   = "#9ece6a"
	WarningColor   = "#e0af68"
	DangerColor    = "#f7768e"
	MutedColor     = "#565f89"
	TextColor      = "#c0caf5"
	SurfaceColor   = "#24283b"
	OverlayColor   = "#414868"
)

// Reusable styles
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PrimaryColor)).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(SecondaryColor))

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(MutedColor))

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(SuccessColor))

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(WarningColor))

	DangerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(DangerColor))

	TextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextColor))

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(PrimaryColor)).
			Padding(1, 2)

	ActivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(SecondaryColor)).
				Padding(1, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(SurfaceColor)).
			Foreground(lipgloss.Color(TextColor)).
			Padding(0, 1)

	BreadcrumbStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(SurfaceColor)).
			Foreground(lipgloss.Color(PrimaryColor)).
			Padding(0, 1)

	SelectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(OverlayColor)).
				Foreground(lipgloss.Color(TextColor))

	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(PrimaryColor)).
				Bold(true).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(OverlayColor))
)

// FormatSize formats a byte count into a human-readable string using SI units.
func FormatSize(bytes int64) string {
	return humanize.Bytes(uint64(bytes))
}

// SizeBarColor returns the appropriate color for a size bar based on the fill ratio.
func SizeBarColor(ratio float64) string {
	switch {
	case ratio >= 0.7:
		return DangerColor
	case ratio >= 0.4:
		return WarningColor
	default:
		return SuccessColor
	}
}

// SizeBar renders a visual bar representing a ratio, colored by severity.
func SizeBar(width int, ratio float64, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 30
	}
	filled := int(ratio * float64(maxWidth))
	if filled < 1 && ratio > 0 {
		filled = 1
	}
	color := SizeBarColor(ratio)
	bar := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(repeatChar('\u2588', filled))
	empty := lipgloss.NewStyle().
		Foreground(lipgloss.Color(MutedColor)).
		Render(repeatChar('\u2591', maxWidth-filled))
	return bar + empty
}

func repeatChar(ch rune, count int) string {
	if count <= 0 {
		return ""
	}
	s := make([]rune, count)
	for i := range s {
		s[i] = ch
	}
	return string(s)
}
