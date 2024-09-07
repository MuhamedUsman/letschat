package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

const (
	chatMenu     = "chatMenu"
	chatViewport = "chatViewport"
	chatTxtarea  = "chatTxtarea"
)

type ChatModel struct {
	chatTxtarea textarea.Model
	focusIdx    int // 0 -> chatTxtarea, 1 -> chatViewport
	focus       bool
}

func InitialChatModel() ChatModel {
	return ChatModel{chatTxtarea: newChatTxtArea()}
}

func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	if !m.focus {
		m.chatTxtarea.Blur()
		m.updateChatTxtareaDimensions()
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateChatTxtareaDimensions()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+t":
			if m.chatTxtarea.Focused() {
				m.chatTxtarea.Blur()
			} else {
				cmd := m.chatTxtarea.Focus()
				m.updateChatTxtareaDimensions()
				return m, cmd
			}
		case "ctrl+s":
			m.chatTxtarea.Reset()
			m.chatTxtarea.Blur()
			m.chatTxtarea.SetHeight(2) // we initialize with 0
		case "esc":
			m.chatTxtarea.Blur()
		}
		m.updateChatTxtareaDimensions()
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
		case tea.MouseButtonWheelDown:
			if zone.Get(chatTxtarea).InBounds(msg) {
				m.chatTxtarea.CursorDown()
			}
		case tea.MouseButtonWheelUp:
			if zone.Get(chatTxtarea).InBounds(msg) {
				m.chatTxtarea.CursorUp()
			}
		case tea.MouseButtonWheelRight:
			if zone.Get(chatTxtarea).InBounds(msg) {
				m.chatTxtarea.SetCursor(m.chatTxtarea.LineInfo().CharOffset + 1)
			}
		case tea.MouseButtonWheelLeft:
			if zone.Get(chatTxtarea).InBounds(msg) {
				m.chatTxtarea.SetCursor(max(0, m.chatTxtarea.LineInfo().CharOffset-1))
			}
		default:
		}
		if zone.Get(chatViewport).InBounds(msg) {
			m.chatTxtarea.Blur()
			m.updateChatTxtareaDimensions()
		}
		if zone.Get(chatTxtarea).InBounds(msg) {
			cmd := m.chatTxtarea.Focus()
			m.updateChatTxtareaDimensions()
			return m, cmd
		}
	}
	return m, m.handleChatTextareaUpdate(msg)
}

func (m ChatModel) View() string {
	h := renderChatHeader("Muhammad Usman")
	ta := zone.Mark(chatTxtarea, m.chatTxtarea.View())
	ta = renderChatTextarea(ta, m.chatTxtarea.Focused())
	chatViewportHeight := chatHeight() - (lipgloss.Height(h) + lipgloss.Height(ta))
	chatView := renderChatViewport("", chatViewportHeight)
	chatView = zone.Mark(chatViewport, chatView)
	c := lipgloss.JoinVertical(lipgloss.Top, h, chatView, ta)
	c = chatContainerStyle.
		Width(chatWidth()).
		Height(chatHeight()).
		Render(c)
	return c
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func newChatTxtArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Prompt = ""
	ta.CharLimit = 2000
	ta.ShowLineNumbers = false
	ta.SetHeight(0)
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(primaryColor)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Foreground(whiteColor)
	return ta
}

func renderChatHeader(name string) string {
	c := chatHeaderStyle.Width(chatWidth())
	menu := zone.Mark(chatMenu, "⚙️")
	menuMarginLeft := max(0, c.GetWidth()-(c.GetHorizontalFrameSize()+lipgloss.Width(name)+lipgloss.Width(menu)))
	menu = lipgloss.NewStyle().
		MarginLeft(menuMarginLeft).
		Render(menu)
	return c.Render(name, menu)
}

func renderChatViewport(content string, h int) string {
	return chatViewportStyle.
		Height(h).
		Width(chatWidth()).
		Render(content)
}

func renderChatTextarea(ta string, padding bool) string {
	cStyle := chatTxtareaStyle.
		Width(chatWidth())
	if padding {
		cStyle = cStyle.UnsetPadding()
		cStyle = cStyle.Height(5)
	}
	return cStyle.Render(ta)
}

func (m *ChatModel) handleChatTextareaUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.chatTxtarea, cmd = m.chatTxtarea.Update(msg)
	return cmd
}

func (m *ChatModel) updateChatTxtareaDimensions() {
	m.chatTxtarea.SetWidth(chatWidth() - chatTxtareaStyle.GetHorizontalFrameSize())
	if m.chatTxtarea.Focused() {
		m.chatTxtarea.SetHeight(5)
	} else {
		m.chatTxtarea.SetHeight(0)
	}
}
