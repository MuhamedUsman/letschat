package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Every width calculation going to be experimental (visually) to some extent

var ( // Global Styling

	// These will be updated by any of the activeTab TabContainerModel
	terminalWidth  int
	terminalHeight int
	terminalFocus  *bool // only read the msg once the terminal in focus

	primaryColor           = lipgloss.AdaptiveColor{Light: "#4b3b00", Dark: "#FFC700"}
	primarySubtleDarkColor = lipgloss.AdaptiveColor{Light: "#6c5300", Dark: "#8b7000"}
	primaryContrastColor   = lipgloss.AdaptiveColor{Light: "#FFC700", Dark: "#4b3b00"}
	dangerColor            = lipgloss.AdaptiveColor{Light: "#ff7b4e", Dark: "#FF5C00"}
	whiteColor             = lipgloss.AdaptiveColor{Light: "#202020", Dark: "#FFFCE4"}
	blackColor             = lipgloss.AdaptiveColor{Light: "#FFFCE4", Dark: "#202020"}
	darkGreyColor          = lipgloss.AdaptiveColor{Light: "#808080", Dark: "#404040"}
	lightGreyColor         = lipgloss.AdaptiveColor{Light: "#404040", Dark: "#afafaf"}
	redColor               = lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}
	orangeColor            = lipgloss.AdaptiveColor{Light: "#ffa000", Dark: "#ffa000"}
	greenColor             = lipgloss.AdaptiveColor{Light: "#00a300", Dark: "#00ff00"}

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
			AlignHorizontal(lipgloss.Left)

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

	statusTextStyle = lipgloss.NewStyle().
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
	discoverBar = inputStyle.Width(68).
			Border(lipgloss.RoundedBorder())

	activeDiscoverBar = activeInputStyle.Width(71).
				Border(lipgloss.RoundedBorder()).
				Align(lipgloss.Center)

	discoverTableStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor)
)

var ( // Conversation Styling

	// updated by TabContainerModel so we can keep the divider proportional to the gap
	tabGapLeftWidth int

	conversationWidth  = func() int { return tabGapLeftWidth - 1 }
	conversationHeight = func() int { return terminalHeight - 4 }

	conversationContainerStyle = lipgloss.NewStyle().
					Padding(0, 1).
					BorderStyle(lipgloss.NormalBorder()).
					BorderRight(true).
					BorderForeground(darkGreyColor)

	conversationSearchBarStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder(), true).
					Padding(0, 1).
					BorderForeground(primarySubtleDarkColor)

	conversationActiveSearchBarStyle = conversationSearchBarStyle.
						BorderForeground(primaryColor)

	conversationOnlineIndicator = lipgloss.NewStyle().
					Foreground(greenColor).
					Render(" *")
)

var (
	tabGapRightWithTabsWidth int
	chatWidth                = func() int { return tabGapRightWithTabsWidth - 2 }
	chatHeight               = func() int { return terminalHeight - 6 }

	chatContainerStyle = lipgloss.NewStyle()

	chatHeaderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(darkGreyColor).
			Foreground(primaryColor).
			Bold(true).
			Margin(1, 3, 0, 3)

	chatHeaderHeight, chatTextareaHeight int // used by ChatModel.chatViewport for its height calculations

	chatTxtareaStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderTop(true).
				BorderForeground(darkGreyColor).
				Margin(0, 3).
				Padding(1, 0)

	chatBubbleContainer = lipgloss.NewStyle().
				Margin(0, 1)

	chatBubbleLBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╰",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}

	chatBubbleRBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╯",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}

	chatBubbleLStyle = lipgloss.NewStyle().
				Border(chatBubbleLBorder, true).
				BorderForeground(whiteColor).
				Padding(0, 1)

	chatBubbleRStyle = lipgloss.NewStyle().
				Border(chatBubbleRBorder, true).
				BorderForeground(primaryColor).
				Padding(0, 1).
				Foreground(primaryColor)

	chatMenuBtnContainerStyle = lipgloss.NewStyle().
					Margin(0, 2)

	chatMenuBtnStyle = lipgloss.NewStyle().
				Margin(0, 1).
				Padding(0, 2)
)

var (
	msgInfoHeaderStyle = lipgloss.NewStyle().
				Background(primaryContrastColor).
				Foreground(primaryColor).
				Margin(2, 5, 0, 5).
				Padding(0, 2).
				Italic(true)

	msgInfoBodyStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.ThickBorder()).
				BorderLeft(true).
				BorderForeground(primaryColor).
				Foreground(primaryColor).
				Margin(2, 5, 0, 5).
				PaddingLeft(2).
				Italic(true)

	msgInfoFooterStyle = lipgloss.NewStyle().
				Margin(2, 5).
				Foreground(primarySubtleDarkColor)

	msgInfoContainerBtn = lipgloss.NewStyle().
				Margin(2, 5, 1, 5)

	msgInfoBtnStyle = lipgloss.NewStyle().
			MarginRight(1).
			Padding(0, 2)
)

var ( // Bunny Stying

	bunnyColor = lipgloss.AdaptiveColor{Light: "#602c1a", Dark: "#6d4534"}

	bunnyText = lipgloss.NewStyle().
			Foreground(primaryColor).
			Align(lipgloss.Center).
			MarginTop(1).
			Render(" Houston, we have a problem.\nNo results in this rabbit hole!")

	bunny = lipgloss.NewStyle().
		Foreground(bunnyColor).
		Render(lipgloss.JoinVertical(lipgloss.Center), b, bunnyText)

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

var (
	rabbitColor = bunnyColor

	rabbit = lipgloss.NewStyle().
		Foreground(primaryColor).
		MarginTop(1).
		Blink(true).
		Render(r)

	r = `
      _........_
   .d888888888888b.
  d888+"""""+888888.
   8b' o  o  '88888b
_.8P (_)    .888888
d88P .____, .d88888P
888b.       :88888P'
'"""""'-----'""""' mh
`
)
