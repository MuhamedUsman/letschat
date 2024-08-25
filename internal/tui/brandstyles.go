package tui

import "github.com/charmbracelet/lipgloss"

var (
	// These will be updated by any of the active model
	terminalWidth  = 100
	terminalHeight = 20

	primaryColor   = lipgloss.AdaptiveColor{Light: "#3115FF", Dark: "#3115FF"}
	secondaryColor = lipgloss.AdaptiveColor{Light: "#FF2473", Dark: "#FF2473"}
	whiteColor     = lipgloss.AdaptiveColor{Light: "#202020", Dark: "#FFFCE4"}
	blackColor     = lipgloss.AdaptiveColor{Light: "#FFFCE4", Dark: "#202020"}
	greyColor      = lipgloss.AdaptiveColor{Light: "#808080", Dark: "#383838"}

	letschatLogo = lipgloss.NewStyle().
			Border(lipgloss.InnerHalfBlockBorder(), true).
			BorderForeground(primaryColor).
			Background(primaryColor).
			Width(10).
			MarginBottom(2).
			Align(lipgloss.Center).
			Italic(true).
			Render("Letschat")

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(greyColor).
			Foreground(greyColor).
			Padding(0, 2, 0, 3).
			Margin(1, 0, 1, 0).
			Align(lipgloss.Center)
	activeInputStyle = inputStyle.
				Border(lipgloss.ThickBorder(), false, false, true, false).
				BorderForeground(primaryColor).
				Foreground(primaryColor)

	btnInputStyle = inputStyle.
			Border(lipgloss.HiddenBorder()).
			MarginBottom(0)
	activeBtnInputStyle = btnInputStyle.
				BorderForeground(primaryColor).
				Foreground(primaryColor)

	buttonStyle = lipgloss.NewStyle().
			Background(greyColor).
			Foreground(whiteColor).
			Width(10).
			Align(lipgloss.Center).
			Inline(true)

	activeButtonStyleWithColor = func(foreground, background lipgloss.AdaptiveColor) lipgloss.Style {
		return buttonStyle.
			Foreground(foreground).
			Background(background)
	}

	container = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(primaryColor).
			Width(70).
			Height(25).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center)
	containerCentered = func(content string) string {
		return lipgloss.Place(terminalWidth, terminalHeight,
			lipgloss.Center, lipgloss.Center,
			content,
			lipgloss.WithWhitespaceChars("▄▀"),
			lipgloss.WithWhitespaceForeground(greyColor))
	}

	infoTxtStyle = lipgloss.NewStyle().
			Margin(1, 0, 2, 0).
			Padding(0, 1, 0, 1).
			AlignHorizontal(lipgloss.Center).
			Foreground(whiteColor)
)
