package tui

import (
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	zone "github.com/lrstanley/bubblezone"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// once ioStatus is not zero valued & spinnerSpinCmd is returned,
// TabContainerModel.spinner will spin with ioStatus until spinnerResetCmd
var ioStatus string

// TabContainerModel -> main TUI model for this application
type TabContainerModel struct {
	discover  DiscoverModel
	letschat  LetschatModel
	tabs      []string
	activeTab int
	errMsg    *errMsg
	timer     timer.Model
	stopwatch stopwatch.Model
	spinner   *spinner.Model
	client    *client.Client
}

func InitialTabContainerModel() TabContainerModel {
	t := []string{
		"🔎 Discover",
		"💭 Conversations",
		"⚙️ Preferences",
	}
	c := client.Get()
	s := spinner.New(spinner.WithStyle(spinnerStyle), spinner.WithSpinner(spinner.Points))
	return TabContainerModel{
		discover:  InitialDiscoverModel(c),
		letschat:  InitialLetschatModel(c),
		tabs:      t,
		activeTab: 1,
		timer:     timer.New(0),
		stopwatch: stopwatch.New(),
		spinner:   &s,
		client:    c,
	}
}

func (m TabContainerModel) Init() tea.Cmd {
	return tea.Batch(
		m.discover.Init(),
		m.letschat.Init(),
		m.stopwatch.Init(),
		m.readOnUsrLoggedInChan(),
		m.runStartUpProcesses(),
	)
}

func (m TabContainerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.setChildModelFocus()
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		terminalHeight = msg.Height
		terminalWidth = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if err := m.client.BT.Shutdown(5 * time.Second); err != nil {
				slog.Error(err.Error())
			}
			return m, tea.Quit
		case "shift+tab":
			if m.activeTab == len(m.tabs)-1 {
				m.activeTab = 0
			} else {
				m.activeTab++
			}
		case "enter":
			if !m.timer.Timedout() {
				return m, nil
			}
		}

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			for i, t := range m.tabs {
				if zone.Get(t).InBounds(msg) {
					m.activeTab = i
				}
			}
		default:
		}

	case client.LoginState:
		if !msg {
			// case requireAuthMsg will take over, also we again listen for state change
			return m, tea.Batch(requireAuthCmd, m.readOnUsrLoggedInChan())
		}

	case requireAuthMsg:
		loginModel := InitialLoginModel()
		return loginModel, loginModel.Init()

	case *errMsg:
		m.resetSpinner()
		m.errMsg = msg
		if m.timer.Timedout() {
			m.timer = timer.New(3 * time.Second)
			return m, m.timer.Init()
		}

	case timer.TickMsg:
		return m, m.handleTimerUpdate(msg)

	case timer.TimeoutMsg:
		m.errMsg = nil

	case spinMsg:
		return m, m.spinner.Tick

	case spinner.TickMsg:
		return m, m.handleSpinnerUpdate(msg)

	case resetSpinnerMsg:
		m.resetSpinner()
	}

	return m, tea.Batch(m.handleChildModelUpdates(msg), m.handleStopwatchUpdate(msg))
}

func (m TabContainerModel) View() string {
	if m.errMsg != nil {
		return renderErrContainer(m.errMsg.err, m.errMsg.code, m.timer.View())
	}

	tabs := make([]string, len(m.tabs))

	for i, t := range m.tabs {
		if i == m.activeTab {
			t = zone.Mark(t, activeTab.Render(t))
			tabs = append(tabs, t)
		} else {
			t = zone.Mark(t, tab.Render(t))
			tabs = append(tabs, t)
		}
	}

	t := lipgloss.JoinHorizontal(
		lipgloss.Center,
		tabs...,
	)
	s := "Session Uptime: " + m.stopwatch.View()
	if ioStatus != "" {
		s = ioStatus + " " + m.spinner.View()
	}
	if m.client.CurrentUsr != nil {
		t = renderTabsWithGapsAndText(t, m.client.CurrentUsr.Name, s, m.client.WsConnState.Get())
	} else {
		// conn State will be ignored in this case
		t = renderTabsWithGapsAndText(t, "", s, m.client.WsConnState.Get())
	}
	content := m.populateActiveTabContent()
	c := renderContainerWithTabs(t, content)
	return zone.Scan(c)
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func renderLeftText(txt string, s client.WsConnState) string {
	is := statusTextStyle
	switch s {
	case client.Disconnected:
		is = is.Foreground(redColor)
	case client.WaitingForReconnection:
		is = is.Foreground(redColor).Blink(true)
	case client.Reconnecting:
		is = is.Foreground(orangeColor).Blink(true)
	case client.Connected:
		is = is.Foreground(greenColor)
	}
	return fmt.Sprint(is.Render("●"), statusTextStyle.UnsetPadding().Render(txt), is.Render("●"))
}

func renderTabsWithGapsAndText(tabs, textL, textR string, state client.WsConnState) string {
	w := (terminalWidth - lipgloss.Width(tabs) - 4) / 2
	gapL := tabGapLeft.Width(w).Render(statusTextStyle.Render("Letschat"))
	// used for divider in conversations tab
	tabGapLeftWidth = lipgloss.Width(gapL)
	gapR := tabGapRight.Width(w).Render(statusTextStyle.Render(textR))
	// used for chat container in conversations tab
	tabGapRightWithTabsWidth = lipgloss.Width(gapR) + lipgloss.Width(tabs)
	if textL != "" {
		gapL = tabGapLeft.Width(w).Render(renderLeftText(textL, state))
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, gapL, tabs, gapR)
}

func renderContainerWithTabs(tabs string, content string) string {
	w := lipgloss.Width(tabs) - 2
	h := terminalHeight - lipgloss.Height(tabs) - 1
	c := tabContainer.Width(max(0, w)).Height(max(0, h)).Render(content)
	return lipgloss.JoinVertical(lipgloss.Center, tabs, c)
}

func renderErrContainer(err string, code int, timer string) string {
	h := errHeaderStyle.Render(strconv.Itoa(code), "-", http.StatusText(code))
	margin := errContainerStyle.GetWidth() - (lipgloss.Width(h) + 6)
	t := lipgloss.NewStyle().Foreground(dangerColor).MarginLeft(margin).Render(timer)
	h = lipgloss.JoinHorizontal(lipgloss.Left, h, t)
	d := errDescStyle.Render(ansi.Wordwrap(err, 58, " ")) // 58 -> sweet spot
	e := lipgloss.JoinVertical(lipgloss.Left, h, d)
	e = errContainerStyle.Render(e)
	return lipgloss.Place(terminalWidth, terminalHeight,
		lipgloss.Center, lipgloss.Center,
		e,
		lipgloss.WithWhitespaceChars("↯"),
		lipgloss.WithWhitespaceForeground(darkGreyColor))
}

func (m *TabContainerModel) setChildModelFocus() {
	m.discover.focus = false
	m.letschat.focus = false
	switch m.activeTab {
	case 0:
		m.discover.focus = true
	case 1:
		m.letschat.focus = true
	}
}

func (m *TabContainerModel) handleChildModelUpdates(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 2)
	m.discover, cmds[0] = m.discover.Update(msg)
	m.letschat, cmds[1] = m.letschat.Update(msg)
	return tea.Batch(cmds...)
}

func (m *TabContainerModel) handleTimerUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.timer, cmd = m.timer.Update(msg)
	return cmd
}

func (m *TabContainerModel) handleStopwatchUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.stopwatch, cmd = m.stopwatch.Update(msg)
	return cmd
}

func (m *TabContainerModel) handleSpinnerUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	*m.spinner, cmd = m.spinner.Update(msg)
	return cmd
}

func (m *TabContainerModel) resetSpinner() {
	s := spinner.New(
		spinner.WithStyle(spinnerStyle),
		spinner.WithSpinner(spinner.Points),
	)
	m.spinner = &s
	ioStatus = ""
}

func (m *TabContainerModel) populateActiveTabContent() string {
	switch m.activeTab {
	case 0:
		return m.discover.View()
	case 1:
		return m.letschat.View()
	default:
		return ""
	}
}

// user is not logged in return requireAuthMsg
func (m TabContainerModel) readOnUsrLoggedInChan() tea.Cmd {
	return func() tea.Msg {
		for {
			return m.client.LoginState.WaitForStateChange()
		}
	}
}

// it runs the processes only once, the func this calls is a sync.OnceFunc
func (m TabContainerModel) runStartUpProcesses() tea.Cmd {
	return func() tea.Msg {
		m.client.RunStartupProcesses()
		return nil
	}
}
