package tui

import (
	"github.com/M0hammadUsman/letschat/internal/client"
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
	chatTxtarea      textarea.Model
	chatViewport     ChatViewport
	focusIdx         int // 0 -> chatTxtarea, 1 -> chatViewport
	focus            bool
	client           *client.Client
	selConvoUsername string
}

func InitialChatModel() ChatModel {
	return ChatModel{
		chatTxtarea:  newChatTxtArea(),
		chatViewport: InitialChatViewport(),
	}
}

func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	if !m.focus {
		m.chatTxtarea.Blur()
		m.updateChatTxtareaAndViewportDimensions()
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateChatTxtareaAndViewportDimensions()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+t":
			cmd := m.chatTxtarea.Focus()
			m.updateChatTxtareaAndViewportDimensions()
			return m, cmd
		case "ctrl+s":
			m.chatTxtarea.Reset()
			m.chatTxtarea.Blur()
			m.chatTxtarea.SetHeight(2) // we initialize with 0
		case "esc":
			m.chatTxtarea.Blur()
			return m, nil
		}
		//m.updateChatTxtareaAndViewportDimensions()
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			if zone.Get("helllo").InBounds(msg) {
				return m, tea.Quit
			}
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
			m.updateChatTxtareaAndViewportDimensions()
		}
		if zone.Get(chatTxtarea).InBounds(msg) {
			cmd := m.chatTxtarea.Focus()
			m.updateChatTxtareaAndViewportDimensions()
			return m, cmd
		}
	}
	return m, tea.Batch(m.handleChatTextareaUpdate(msg), m.handleChatViewportUpdate(msg))
}

func (m ChatModel) View() string {
	if m.selConvoUsername == "" {
		return chatContainerStyle.
			Width(chatWidth()).
			Height(chatHeight()).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render(rabbit)
	}
	h := renderChatHeader(m.selConvoUsername)
	chatHeaderHeight = lipgloss.Height(h)
	ta := zone.Mark(chatTxtarea, m.chatTxtarea.View())
	ta = renderChatTextarea(ta, m.chatTxtarea.Focused())
	chatTextareaHeight = lipgloss.Height(ta)
	m.chatViewport.vp.Height = chatHeight() - (chatHeaderHeight + chatTextareaHeight)
	chatView := m.chatViewport.View()
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
	ta.CharLimit = 1000
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

func (m *ChatModel) handleChatViewportUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.chatViewport, cmd = m.chatViewport.Update(msg)
	return cmd
}

func (m *ChatModel) updateChatTxtareaAndViewportDimensions() {
	m.chatTxtarea.SetWidth(chatWidth() - chatTxtareaStyle.GetHorizontalFrameSize())
	if m.chatTxtarea.Focused() {
		m.chatTxtarea.SetHeight(5)
	} else {
		m.chatTxtarea.SetHeight(0)
	}
	m.chatViewport.updateDimensions()
}
