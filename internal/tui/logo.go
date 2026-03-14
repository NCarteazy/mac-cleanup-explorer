package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

// renderLogo renders the ASCII art logo with a gradient effect.
// Extracted here so it can be reused in the scan screen and help overlay.
func renderLogo() string {
	// Tokyo Night gradient colors from cyan through blue to purple
	gradientColors := []string{
		"#7dcfff", // cyan (PrimaryColor)
		"#7cc4f5",
		"#7ab9eb",
		"#79aee1",
		"#7ba3dc",
		"#7e98d7",
		"#828dd2",
		"#8882cc",
		"#8f78c5",
		"#976dbd",
		"#a063b5",
		"#a95aad",
		"#b351a4",
		"#bb9af7", // purple (SecondaryColor)
	}

	logoLines := []string{
		"  ███╗   ███╗ █████╗  ██████╗",
		"  ████╗ ████║██╔══██╗██╔════╝",
		"  ██╔████╔██║███████║██║     ",
		"  ██║╚██╔╝██║██╔══██║██║     ",
		"  ██║ ╚═╝ ██║██║  ██║╚██████╗",
		"  ╚═╝     ╚═╝╚═╝  ╚═╝ ╚═════╝",
	}

	cleanupLines := []string{
		"   ██████╗██╗     ███████╗ █████╗ ███╗   ██╗██╗   ██╗██████╗ ",
		"  ██╔════╝██║     ██╔════╝██╔══██╗████╗  ██║██║   ██║██╔══██╗",
		"  ██║     ██║     █████╗  ███████║██╔██╗ ██║██║   ██║██████╔╝",
		"  ██║     ██║     ██╔══╝  ██╔══██║██║╚██╗██║██║   ██║██╔═══╝ ",
		"  ╚██████╗███████╗███████╗██║  ██║██║ ╚████║╚██████╔╝██║     ",
		"   ╚═════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝╚═╝     ",
	}

	explorerLines := []string{
		"  ███████╗██╗  ██╗██████╗ ██╗      ██████╗ ██████╗ ███████╗██████╗ ",
		"  ██╔════╝╚██╗██╔╝██╔══██╗██║     ██╔═══██╗██╔══██╗██╔════╝██╔══██╗",
		"  █████╗   ╚███╔╝ ██████╔╝██║     ██║   ██║██████╔╝█████╗  ██████╔╝",
		"  ██╔══╝   ██╔██╗ ██╔═══╝ ██║     ██║   ██║██╔══██╗██╔══╝  ██╔══██╗",
		"  ███████╗██╔╝ ██╗██║     ███████╗╚██████╔╝██║  ██║███████╗██║  ██║",
		"  ╚══════╝╚═╝  ╚═╝╚═╝     ╚══════╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝",
	}

	// Render each block of lines with gradient colors
	styledMAC := renderGradientBlock(logoLines, gradientColors[:6])
	styledCleanup := renderGradientBlock(cleanupLines, gradientColors[4:10])
	styledExplorer := renderGradientBlock(explorerLines, gradientColors[8:])

	// Subtitle under the big logo
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedColor)).
		Italic(true).
		Render("           Reclaim your disk space")

	// Decorative divider
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.OverlayColor)).
		Render("  " + strings.Repeat("─", 56))

	return lipgloss.JoinVertical(lipgloss.Left,
		styledMAC,
		styledCleanup,
		styledExplorer,
		"",
		subtitle,
		divider,
	)
}

// renderGradientBlock applies a vertical gradient to a block of text lines.
func renderGradientBlock(lines []string, colors []string) string {
	styled := make([]string, len(lines))
	for i, line := range lines {
		// Map line index to color index
		colorIdx := 0
		if len(lines) > 1 {
			colorIdx = i * (len(colors) - 1) / (len(lines) - 1)
		}
		if colorIdx >= len(colors) {
			colorIdx = len(colors) - 1
		}
		styled[i] = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors[colorIdx])).
			Render(line)
	}
	return strings.Join(styled, "\n")
}
