package tui

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

const (
	letschatConversation = "letschatConversation"
	letschatChat         = "letschatChat"
)

// global for performance reasons, and ease of logic
var (
	// selected user from conversations
	selUserID, selUsername string
	selUserTyping          bool
	// if false msg will not be sent, and ConversationModel will not call for createConvoIfNotExist()
	validMsgForSend bool
)

type LetschatModel struct {
	conversation ConversationModel
	chat         ChatModel
	focus        bool
	client       *client.Client
}

func InitialLetschatModel(c *client.Client) LetschatModel {
	return LetschatModel{
		conversation: InitialConversationModel(c),
		chat:         InitialChatModel(c),
		client:       c,
	}
}

func (m LetschatModel) Init() tea.Cmd {
	return tea.Batch(m.conversation.Init(), m.chat.Init())
}

func (m LetschatModel) Update(msg tea.Msg) (LetschatModel, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.MouseMsg:
		m.conversation.focus = false
		m.chat.focus = false
		if zone.Get(letschatConversation).InBounds(msg) && m.focus {
			m.conversation.focus = true
			m.chat.focus = false

		} else if zone.Get(letschatChat).InBounds(msg) && m.focus {
			m.chat.focus = true
			m.conversation.focus = false
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+t":
			m.chat.focus = true
			m.conversation.focus = false
		case "ctrl+x":
			selUserID, selUsername, selUserTyping = "", "", false
		}
	}
	return m, tea.Batch(m.handleConversationUpdate(msg), m.handleChatUpdate(msg))
}

func (m LetschatModel) View() string {
	convo := m.conversation.View()
	convo = zone.Mark(letschatConversation, convo)
	chat := m.chat.View()
	chat = zone.Mark(letschatChat, chat)
	return lipgloss.JoinHorizontal(lipgloss.Left, convo, chat)
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func (m *LetschatModel) handleConversationUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.conversation, cmd = m.conversation.Update(msg)
	return cmd
}

func (m *LetschatModel) handleChatUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.chat, cmd = m.chat.Update(msg)
	return cmd
}
