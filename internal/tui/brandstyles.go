package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	LetschatLogo = lipgloss.NewStyle().
			Border(lipgloss.InnerHalfBlockBorder()).
			BorderForeground(secondaryColor).
			Foreground(lipgloss.Color("#202020")).
			Background(secondaryColor).
			Italic(true).
			Bold(true).
			Align(lipgloss.Center).
			Width(10)

	primaryColor   = lipgloss.AdaptiveColor{Light: "#133CCA", Dark: "#133CCA"}
	secondaryColor = lipgloss.AdaptiveColor{Light: "#00B597", Dark: "#2FD1B2"}
	inactiveColor  = lipgloss.AdaptiveColor{Light: "#404040", Dark: "#505050"}

	WindowBorder = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(secondaryColor).
			Align(lipgloss.Center).
			Margin(10).
			Width(100).
			Height(30)

	InputBorder = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, false, true).
			Padding(1, 0, 1, 2).
			Margin(1, 0, 1, 1).
			Italic(true)

	InputInactiveBorder = InputBorder.
				BorderLeftForeground(inactiveColor).
				Foreground(inactiveColor)

	InputActiveBorder = InputBorder.
				BorderLeftForeground(secondaryColor).
				Foreground(primaryColor)

	InputButton = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			Margin(33, 0, 0, 4).
			Align(lipgloss.Center).
			Width(10).
			Inline(true)

	InputInactiveButton = InputButton.
				Foreground(lipgloss.Color("#202020")).
				Background(lipgloss.Color("#808080"))

	ContinueButton = InputButton.
			Foreground(lipgloss.Color("#FFFCE4")).
			Background(lipgloss.Color("#3115FF"))

	QuitButton = InputButton.
			Foreground(lipgloss.Color("#FFFCE4")).
			Background(lipgloss.Color("#ff2473"))
)
