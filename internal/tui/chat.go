package tui

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	zone "github.com/lrstanley/bubblezone"
	"slices"
	"strings"
	"time"
)

const (
	chatMenu     = "chatMenu"
	chatViewport = "chatViewport"
	chatTxtarea  = "chatTxtarea"
)

type ChatModel struct {
	chatTxtarea  textarea.Model
	chatViewport ChatViewportModel
	focus        bool
	client       *client.Client
	onlineUsrIds []string
	cb           convosBroadcast
}

func InitialChatModel(c *client.Client) ChatModel {
	return ChatModel{
		chatTxtarea:  newChatTxtArea(),
		chatViewport: InitialChatViewport(c),
		client:       c,
	}
}

func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.chatViewport.Init())
}

func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	if !m.focus {
		m.chatViewport.focus = false
		m.chatTxtarea.Blur()
		m.updateChatTxtareaAndViewportDimensions()
	} else if m.chatTxtarea.Focused() {
		m.chatViewport.focus = false
	} else {
		m.chatViewport.focus = true
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
			s := m.chatTxtarea.Value()
			s = strings.TrimSpace(s)
			if len(s) == 0 {
				return m, nil
			}
			m.chatTxtarea.Reset()
			return m, tea.Batch(m.sendMessage(s), m.handleChatTextareaUpdate(msg))
		case "esc":
			m.chatTxtarea.Blur()
			m.updateChatTxtareaAndViewportDimensions()
			return m, nil
		}
		//m.updateChatTxtareaAndViewportDimensions()
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			if zone.Get(chatTxtarea).InBounds(msg) && m.focus {
				m.chatTxtarea.CursorDown()
			}
		case tea.MouseButtonWheelUp:
			if zone.Get(chatTxtarea).InBounds(msg) && m.focus {
				m.chatTxtarea.CursorUp()
			}
		case tea.MouseButtonWheelRight:
			if zone.Get(chatTxtarea).InBounds(msg) && m.focus {
				m.chatTxtarea.SetCursor(m.chatTxtarea.LineInfo().CharOffset + 1)
			}
		case tea.MouseButtonWheelLeft:
			if zone.Get(chatTxtarea).InBounds(msg) && m.focus {
				m.chatTxtarea.SetCursor(max(0, m.chatTxtarea.LineInfo().CharOffset-1))
			}
		default:
		}
		if zone.Get(chatTxtarea).InBounds(msg) {
			cmd := m.chatTxtarea.Focus() // cmd must be fetched before the update of dimensions
			m.updateChatTxtareaAndViewportDimensions()
			return m, cmd
		}
		if zone.Get(chatViewport).InBounds(msg) {
			m.chatTxtarea.Blur()
			m.updateChatTxtareaAndViewportDimensions()
		}

	case UsrOnlineMsg:
		if !slices.Contains(m.onlineUsrIds, msg.SenderID) {
			m.onlineUsrIds = append(m.onlineUsrIds, msg.SenderID)
		}

	case UsrOfflineMsg:
		if slices.Contains(m.onlineUsrIds, msg.SenderID) {
			for i, id := range m.onlineUsrIds {
				if id == msg.SenderID {
					m.onlineUsrIds = append(m.onlineUsrIds[:i-1], m.onlineUsrIds[i+1:]...)
					break
				}
			}
		}

	}

	return m, tea.Batch(m.handleChatTextareaUpdate(msg), m.handleChatViewportUpdate(msg))
}

func (m ChatModel) View() string {
	if selUsername == "" {
		return chatContainerStyle.
			Width(chatWidth()).
			Height(chatHeight()).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render(rabbit)
	}
	online := false
	typing := false
	if slices.Contains(m.onlineUsrIds, selUserID) {
		online = true
	}
	h := renderChatHeader(selUsername, online, typing)
	chatHeaderHeight = lipgloss.Height(h)
	ta := zone.Mark(chatTxtarea, m.chatTxtarea.View())
	ta = renderChatTextarea(ta, m.chatTxtarea.Focused())
	chatTextareaHeight = lipgloss.Height(ta)
	m.chatViewport.vp.Height = chatHeight() - (chatHeaderHeight + chatTextareaHeight)
	if m.chatTxtarea.Focused() { // only works after setting vp height
		m.chatViewport.vp.GotoBottom()
	}
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

func renderChatHeader(name string, online, typing bool) string {
	c := chatHeaderStyle.Width(chatWidth())
	menu := zone.Mark(chatMenu, "⚙️")
	sub := c.GetHorizontalFrameSize() + lipgloss.Width(name) + lipgloss.Width(menu) + lipgloss.Width(onlineIndicator)
	menuMarginLeft := max(0, c.GetWidth()-sub)
	menu = lipgloss.NewStyle().
		MarginLeft(menuMarginLeft).
		Render(menu)
	s := onlineIndicator
	if !online {
		s = ""
	}
	return c.Render(name, s, menu)
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

func (m *ChatModel) sendMessage(msg string) tea.Cmd {
	t := time.Now()
	msgToSnd := domain.Message{
		ID:           uuid.New().String(),
		SenderID:     m.client.CurrentUsr.ID,
		ReceiverID:   selUserID,
		Body:         msg,
		SentAt:       &t,
		Operation:    domain.CreateMsg,
		Confirmation: 0,
	}
	go m.client.SendMessage(msgToSnd)
	return func() tea.Msg {
		// will be used in ChatViewportModel's update method
		return SentMsg(&msgToSnd)
	}
}
