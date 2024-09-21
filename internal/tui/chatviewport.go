package tui

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"slices"
	"strings"
	"time"
)

type msgPage struct {
	msgs []*domain.Message
	meta *domain.Metadata
}

type ChatViewportModel struct {
	vp                 viewport.Model
	msgs               []*domain.Message
	currPage, lastPage int
	selUsrID           string
	focus              bool
	fetching           bool
	client             *client.Client
}

func InitialChatViewport(c *client.Client) ChatViewportModel {
	return ChatViewportModel{
		vp:     viewport.New(0, 0),
		msgs:   make([]*domain.Message, 0),
		client: c,
	}
}

func (m ChatViewportModel) Init() tea.Cmd {
	m.fetching = true
	return m.listenForMessages()
}

func (m ChatViewportModel) Update(msg tea.Msg) (ChatViewportModel, tea.Cmd) {
	if m.focus {
		m.vp.KeyMap = viewport.DefaultKeyMap()
		m.vp.MouseWheelEnabled = true
	} else {
		m.vp.KeyMap = viewport.KeyMap{}
		m.vp.MouseWheelEnabled = false
	}

	if m.selUsrID != selUserID {
		m.msgs = slices.Delete(m.msgs, 0, len(m.msgs))
		m.msgs = nil
		m.selUsrID = selUserID
		return m, m.getMsgAsPage(1)
	}
	if m.vp.AtTop() && !m.fetching {
		if m.lastPage != m.currPage {
			m.fetching = true
			return m, m.getMsgAsPage(m.currPage + 1)
		}
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.vp.SetContent(m.renderChatViewport())
		return m, m.handleChatViewportUpdate(msg)
	case msgPage:
		m.fetching = false
		m.msgs = append(m.msgs, msg.msgs...)
		m.currPage = msg.meta.CurrentPage
		m.lastPage = msg.meta.LastPage
		m.vp.SetContent(m.renderChatViewport())
		m.vp.GotoBottom()
		return m, m.handleChatViewportUpdate(msg)
	case *domain.Message:
		switch msg.Operation {
		case domain.CreateMsg:
			m.msgs = append(m.msgs, msg)
			m.vp.SetContent(m.renderChatViewport())
			m.vp.GotoBottom()
			return m, tea.Batch(m.listenForMessages(), m.handleChatViewportUpdate(msg))
		case domain.UpdateMsg:

		}
	}
	return m, m.handleChatViewportUpdate(msg)
}

func (m ChatViewportModel) View() string {
	return m.vp.View()
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func newChatViewport(w, h int) viewport.Model {
	vp := viewport.New(w, h)
	vp.MouseWheelEnabled = true
	return vp
}

func (m *ChatViewportModel) renderChatViewport() string {
	var sb strings.Builder
	var prevMsgDay int
	cb := chatBubbleContainer.Width(chatWidth() - chatBubbleContainer.GetHorizontalFrameSize())
	for i := len(m.msgs) - 1; i >= 0; i-- {
		msg := m.msgs[i]
		if msg.SentAt.Day() != prevMsgDay {
			prevMsgDay = msg.SentAt.Day()
			s := lipgloss.NewStyle().
				Foreground(primaryColor).
				Background(primaryContrastColor).
				Padding(0, 1).
				Italic(true).
				Render(msg.SentAt.Format("January 02, 2006"))
			sb.WriteString("\n\n")
			s = cb.Align(lipgloss.Center).Italic(true).Render(s)
			sb.WriteString(s)
			sb.WriteString("\n")
		}
		// bubble style
		align := lipgloss.Left
		if msg.SenderID == m.client.CurrentUsr.ID {
			align = lipgloss.Right
		}
		cb := cb.
			Align(align).
			Render(m.renderBubbleWithStatusInfo(msg))
		// TODO: Add functionality to show time and stuff
		sb.WriteString("\n")
		sb.WriteString(cb)
	}
	return sb.String()
}

func (m *ChatViewportModel) renderBubbleWithStatusInfo(msg *domain.Message) string {
	txtWidth := min(chatWidth()-20, lipgloss.Width(msg.Body)+2)
	bubble := chatBubbleLStyle.Width(txtWidth).Render(msg.Body)
	sentAt := lipgloss.NewStyle().Faint(true).Foreground(whiteColor).SetString(msg.SentAt.Format(time.Kitchen))

	status := "⁎"
	if msg.DeliveredAt != nil {
		status = "⁑"
	}
	if msg.ReadAt != nil {
		status = "⁂"
	}
	status = lipgloss.NewStyle().Faint(true).Foreground(primaryColor).Render(status)

	if msg.SenderID == m.client.CurrentUsr.ID {
		bubble = chatBubbleRStyle.Width(txtWidth).Render(msg.Body)
		sentAt = sentAt.Foreground(primaryColor)
		return lipgloss.JoinHorizontal(lipgloss.Center, status, " ", sentAt.Render(), " ", bubble)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, bubble, " ", sentAt.Render())
}

func (m *ChatViewportModel) updateDimensions() {
	m.vp.Width = chatWidth()
	m.vp.Height = chatHeight() - (chatHeaderHeight + chatTextareaHeight)
}

func (m *ChatViewportModel) handleChatViewportUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return cmd
}

func (m ChatViewportModel) listenForMessages() tea.Cmd {
	return func() tea.Msg {
		ctx := m.client.BT.GetShtdwnCtx()
		for {
			msg := m.client.RecvMsgs.WaitForStateChange()
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			if msg == nil {
				return nil
			}
			// if the msg has to do something with the selected chat then
			if msg.SenderID == selUserID || msg.ReceiverID == selUserID {
				return msg
			}
		}
	}
}

func (m ChatViewportModel) getMsgAsPage(p int) tea.Cmd {
	return func() tea.Msg {
		msgs, meta, err := m.client.GetMessagesAsPage(selUserID, p)
		if err != nil {
			return &errMsg{
				err:  "Unable to fetch initial chat for this user...",
				code: 0,
			}
		}
		return msgPage{msgs, meta}
	}
}
