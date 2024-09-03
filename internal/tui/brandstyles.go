package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var ( // Global Styling

	// These will be updated by any of the activeTab TabContainerModel
	terminalWidth  = 100
	terminalHeight = 20

	primaryColor         = lipgloss.AdaptiveColor{Light: "#3c3000", Dark: "#FFC700"}
	primaryContrastColor = lipgloss.AdaptiveColor{Light: "#FFC700", Dark: "#3c3000"}
	dangerColor          = lipgloss.AdaptiveColor{Light: "#ff7b4e", Dark: "#FF5C00"}
	whiteColor           = lipgloss.AdaptiveColor{Light: "#202020", Dark: "#FFFCE4"}
	blackColor           = lipgloss.AdaptiveColor{Light: "#FFFCE4", Dark: "#202020"}
	darkGreyColor        = lipgloss.AdaptiveColor{Light: "#808080", Dark: "#383838"}
	lightGreyColor       = lipgloss.AdaptiveColor{Light: "#404040", Dark: "#afafaf"}

	letschatLogo = lipgloss.NewStyle().
			Border(lipgloss.InnerHalfBlockBorder(), true).
			BorderForeground(primaryColor).
			Background(primaryColor).
			Foreground(primaryContrastColor).
			Width(10).
			MarginBottom(2).
			Align(lipgloss.Center).
			Italic(true).
			Render("Letschat")
)

var ( // Form Styling

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(darkGreyColor).
			Foreground(darkGreyColor).
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
				Foreground(primaryContrastColor)

	buttonStyle = lipgloss.NewStyle().
			Background(darkGreyColor).
			Foreground(whiteColor).
			Width(10).
			Align(lipgloss.Center).
			Inline(true)

	activeButtonStyleWithColor = func(foreground, background lipgloss.AdaptiveColor) lipgloss.Style {
		return buttonStyle.
			Foreground(foreground).
			Background(background)
	}

	formContainer = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(primaryColor).
			Width(70).
			Height(25).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center)
	formContainerCentered = func(content string) string {
		return lipgloss.Place(terminalWidth, terminalHeight,
			lipgloss.Center, lipgloss.Center,
			content,
			lipgloss.WithWhitespaceChars("+"),
			lipgloss.WithWhitespaceForeground(darkGreyColor))
	}

	infoTxtStyle = lipgloss.NewStyle().
			Margin(1, 0, 2, 0).
			Padding(0, 1, 0, 1).
			AlignHorizontal(lipgloss.Center).
			Foreground(whiteColor)

	otpInputStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, true, false).
			BorderForeground(darkGreyColor).
			Padding(0, 1, 0, 1).
			Margin(1, 0, 1, 0).
			Width(10).
			Align(lipgloss.Center)
)

var ( // Tab Container Styling

	tabContainer = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			BorderForeground(primaryColor).
			AlignHorizontal(lipgloss.Center)

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(primaryColor).
		Foreground(lightGreyColor).
		Padding(0, 1)

	activeTab = tab.Border(activeTabBorder, true).
			Foreground(primaryColor)

	tabGap = lipgloss.NewStyle().
		BorderForeground(primaryColor).
		BorderBottom(true).
		Padding(0, 1).
		Align(lipgloss.Center)

	tabGapLeft  = tabGap.Border(lipgloss.Border{Bottom: "─", BottomLeft: "╭", BottomRight: "─"})
	tabGapRight = tabGap.Border(lipgloss.Border{Bottom: "─", BottomRight: "╮", BottomLeft: "─"})

	statusText = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(lightGreyColor).
			Background(primaryContrastColor).
			Italic(true).
			Align(lipgloss.Center)

	errContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(dangerColor).
				Foreground(dangerColor).
				Width(61).
				Padding(1, 2)

	errHeaderStyle = lipgloss.NewStyle().
			Background(dangerColor).
			Foreground(whiteColor).
			Padding(0, 1)

	errDescStyle = lipgloss.NewStyle().
			Foreground(dangerColor).
			MarginTop(1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)
)

var ( // Discover Styling

	// discoverBar -> SearchBar
	discoverBar = inputStyle.Width(51).
			Border(lipgloss.RoundedBorder())

	activeDiscoverBar = activeInputStyle.Width(51).
				Border(lipgloss.RoundedBorder()).
				Align(lipgloss.Center)

	discoverTableStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor)
)

var (
	bunny = lipgloss.NewStyle().
		Foreground(primaryContrastColor).
		Render(lipgloss.JoinVertical(lipgloss.Center), b, bunnyText)

	bunnyText = lipgloss.NewStyle().
			Foreground(primaryColor).
			Align(lipgloss.Center).
			MarginTop(1).
			Render(" Houston, we have a problem.\nNo results in this rabbit hole!")
	b = `
....▓▓▓▓
..▓▓......▓
..▓▓......▓▓..................▓▓▓▓
..▓▓......▓▓..............▓▓......▓▓▓▓
..▓▓....▓▓..............▓......▓▓......▓▓
....▓▓....▓............▓....▓▓....▓▓▓....▓▓
......▓▓....▓........▓....▓▓..........▓▓....▓
........▓▓..▓▓....▓▓..▓▓................▓▓
........▓▓......▓▓....▓▓
.......▓......................▓
.....▓.........................▓
....▓......^..........^......▓
....▓...........🤎............▓
....▓..........................▓
......▓........ ٮ ..........▓
..........▓▓..........▓▓
`
)
