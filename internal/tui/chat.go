package tui

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	zone "github.com/lrstanley/bubblezone"
	"net/http"
	"strings"
	"time"
)

const (
	chatHeaderContainer = "chatHeaderContainer"
	chatMenu            = "chatMenu"
	menuGotoFirstMsgBtn = "menuGotoFirstMsgBtn"
	menuClearConvoBtn   = "menuClearConvoBtn"
	chatViewport        = "chatViewport"
	chatTxtarea         = "chatTxtarea"
)

type ChatModel struct {
	chatTxtarea    textarea.Model
	chatViewport   ChatViewportModel
	focus          bool
	prevChatLength int
	// menu buttons, -1 -> None Selected | 0 -> Goto First Msg | 1 -> Clear Conversation
	menuBtnIdx int
	client     *client.Client
	cb         convosBroadcast
}

func InitialChatModel(c *client.Client) ChatModel {
	return ChatModel{
		chatTxtarea:  newChatTxtArea(),
		chatViewport: InitialChatViewport(c),
		menuBtnIdx:   -1,
		client:       c,
	}
}

func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.chatViewport.Init(), echoTypingCmd()) // not using timer -> it have bugs
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

	var typingCmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.updateChatTxtareaAndViewportDimensions()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+t":
			typingCmd = m.chatTxtarea.Focus()
			m.menuBtnIdx = -1
			m.updateChatTxtareaAndViewportDimensions()
		case "ctrl+s":
			s := m.chatTxtarea.Value()
			s = strings.TrimSpace(s)
			if len(s) == 0 {
				return m, nil
			}
			m.chatTxtarea.Reset()
			return m, tea.Batch(m.sendMessage(s), m.handleChatTextareaUpdate(msg), m.handleChatViewportUpdate(msg))
		case "left":
			if m.menuBtnIdx == 1 {
				m.menuBtnIdx--
			}
		case "right":
			if m.menuBtnIdx == 0 {
				m.menuBtnIdx++
			}
		case "tab":
			if m.menuBtnIdx > -1 && m.menuBtnIdx <= 1 {
				m.menuBtnIdx = (m.menuBtnIdx + 1) % 2
			}
		case "esc", "ctrl+f":
			if m.menuBtnIdx != -1 {
				m.menuBtnIdx = -1
			}
			m.chatTxtarea.Blur()
			m.updateChatTxtareaAndViewportDimensions()
		case "enter":
			switch m.menuBtnIdx {
			case 0:
				m.chatViewport.gotoFirstMsg = true
				m.menuBtnIdx = -1
			case 1:
				m.menuBtnIdx = -1
				return m, m.deleteAllMsgsForConvo(m.client.CurrentUsr.ID, selUserID)
			}
		}

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
		case tea.MouseButtonLeft:
			if zone.Get(menuGotoFirstMsgBtn).InBounds(msg) {
				m.menuBtnIdx = 0
			}
			if zone.Get(menuClearConvoBtn).InBounds(msg) {
				m.menuBtnIdx = 1
			}
		default:
		}

		if zone.Get(chatMenu).InBounds(msg) &&
			msg.Button == tea.MouseButtonLeft &&
			msg.Action == tea.MouseActionRelease {
			m.menuBtnIdx = 0
		}

		if !zone.Get(chatHeaderContainer).InBounds(msg) {
			m.menuBtnIdx = -1
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

	case echoTypingMsg:
		var cmd tea.Cmd
		if m.prevChatLength < m.chatTxtarea.Length() && !selUserTyping {
			m.prevChatLength = m.chatTxtarea.Length()
			cmd = m.sendTypingStatus()
		}
		// echo again to continue the cycle
		return m, tea.Batch(cmd, echoTypingCmd())

	}

	return m, tea.Batch(typingCmd, m.handleChatTextareaUpdate(msg), m.handleChatViewportUpdate(msg))
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
	h := renderChatHeader(selUsername, selUserTyping)
	if m.menuBtnIdx != -1 {
		h = renderMenuBtns(m.menuBtnIdx)
	}
	chatHeaderHeight = lipgloss.Height(h)
	ta := zone.Mark(chatTxtarea, m.chatTxtarea.View())
	ta = renderChatTextarea(ta, m.chatTxtarea.Focused())
	chatTextareaHeight = lipgloss.Height(ta)
	m.chatViewport.chatVp.Height = chatHeight() - (chatHeaderHeight + chatTextareaHeight)
	if m.chatTxtarea.Focused() { // only works after setting chatVp height
		m.chatViewport.chatVp.GotoBottom()
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

func renderChatHeader(name string, typing bool) string {
	c := chatHeaderStyle.Width(chatWidth())
	menu := zone.Mark(chatMenu, "⚙️")
	sub := c.GetHorizontalFrameSize() + lipgloss.Width(name) + lipgloss.Width(menu)
	menuMarginLeft := max(0, c.GetWidth()-sub)
	menu = lipgloss.NewStyle().
		MarginLeft(menuMarginLeft).
		Render(menu)
	name = lipgloss.NewStyle().Blink(typing).Render(name)
	return zone.Mark(chatHeaderContainer, c.Render(name, menu))
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

func renderMenuBtns(selection int) string {
	if selection == -1 {
		return ""
	}

	gotoFirstMsgBtn := renderGotoFirstMsgBtn(true)
	clearConvoBtn := renderClearConvoBtn(false)
	if selection == 1 {
		gotoFirstMsgBtn = renderGotoFirstMsgBtn(false)
		clearConvoBtn = renderClearConvoBtn(true)
	}
	gotoFirstMsgBtn = zone.Mark(menuGotoFirstMsgBtn, gotoFirstMsgBtn)
	clearConvoBtn = zone.Mark(menuClearConvoBtn, clearConvoBtn)
	btnContainer := chatMenuBtnContainerStyle.Render(gotoFirstMsgBtn, clearConvoBtn)

	c := chatHeaderStyle.Width(chatWidth())
	content := lipgloss.PlaceHorizontal(chatWidth()-c.GetHorizontalFrameSize(), lipgloss.Center, btnContainer)

	return zone.Mark(chatHeaderContainer, c.Render(content))
}

func renderGotoFirstMsgBtn(focus bool) string {
	bg := primaryColor
	fg := primaryContrastColor
	if !focus {
		bg = darkGreyColor
		fg = lightGreyColor
	}
	return chatMenuBtnStyle.
		Background(bg).
		Foreground(fg).
		Render("GOTO FIRST MESSAGE")
}

func renderClearConvoBtn(focus bool) string {
	bg := dangerColor
	fg := whiteColor
	if !focus {
		bg = darkGreyColor
		fg = lightGreyColor
	}
	return chatMenuBtnStyle.
		Background(bg).
		Foreground(fg).
		Render("CLEAR CONVERSATION")
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
		ID:         uuid.New().String(),
		SenderID:   m.client.CurrentUsr.ID,
		ReceiverID: selUserID,
		Body:       msg,
		SentAt:     &t,
		Operation:  domain.CreateMsg,
	}
	return func() tea.Msg {
		if m.client.WsConnState.Get() != client.Connected {
			return &errMsg{
				err:  "No Connection, Unable to send message.",
				code: http.StatusRequestTimeout,
			}
		}
		go m.client.SendMessage(msgToSnd)
		// will be used in ChatViewportModel's update method
		return SentMsg(&msgToSnd)
	}
}

func (m *ChatModel) sendTypingStatus() tea.Cmd {
	t := time.Now()
	msgToSnd := domain.Message{
		ID:           uuid.New().String(),
		SenderID:     m.client.CurrentUsr.ID,
		ReceiverID:   selUserID,
		SentAt:       &t,
		Operation:    domain.TypingMsg,
		Confirmation: 0,
	}
	return func() tea.Msg {
		m.client.SendTypingStatus(msgToSnd)
		return nil
	}
}

func (m ChatModel) deleteAllMsgsForConvo(currUsrId, selUsrId string) tea.Cmd {
	return func() tea.Msg {
		if err := m.client.DeleteForMeAllMsgsForConversation(currUsrId, selUsrId); err != nil {
			return &errMsg{
				err:  "Unable to clear conversation",
				code: 0,
			}
		}
		return clearConvoSuccess{}
	}
}
