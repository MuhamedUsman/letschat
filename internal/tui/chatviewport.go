package tui

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log/slog"
	"slices"
	"strings"
	"time"
)

type msgPage struct {
	msgs []*domain.Message
	meta *domain.Metadata
}

type msgBroadcast struct {
	ch    <-chan *domain.Message
	token int
}

type ChatViewportModel struct {
	vp                        viewport.Model
	msgs                      []*domain.Message
	currPage, lastPage        int
	selUsrID                  string
	focus                     bool
	fetching                  bool
	recvTypingTimer           timer.Model
	lastTypingStateReceivedAt time.Time
	prevLineCount             int
	client                    *client.Client
	mb                        msgBroadcast
}

func InitialChatViewport(c *client.Client) ChatViewportModel {
	token, ch := c.RecvMsgs.Subscribe()
	return ChatViewportModel{
		vp:              viewport.New(0, 0),
		msgs:            make([]*domain.Message, 0),
		client:          c,
		recvTypingTimer: timer.New(2 * time.Second),
		mb: msgBroadcast{
			ch:    ch,
			token: token,
		},
	}
}

func (m ChatViewportModel) Init() tea.Cmd {
	m.fetching = true
	return tea.Batch(m.listenForMessages(), m.recvTypingTimer.Init())
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
		// Once we set the content it takes us to the top, we want to go to the point where the user were before
		c := m.vp.TotalLineCount() - m.prevLineCount
		m.vp.LineDown(c)
		// Now update the prev line count
		m.prevLineCount = m.vp.TotalLineCount()
		return m, m.handleChatViewportUpdate(msg)

	case *domain.Message:
		var cmd tea.Cmd
		switch msg.Operation {
		case domain.CreateMsg:
			m.msgs = append([]*domain.Message{msg}, m.msgs...)
			m.vp.SetContent(m.renderChatViewport())
			m.vp.GotoBottom()
			// set it as read also | nil check, if the terminal focus is not supported, just set the msg as read
			if msg.SenderID == selUserID && msg.ReadAt == nil && (terminalFocus == nil || *terminalFocus) {
				t := time.Now()
				msg.ReadAt = &t
				m.setMsgAsRead(msg)
			}
		case domain.UpdateMsg:
			m.updateMsgInMsgs(msg)
			// the above op will update the msgs so we need to rerender
			m.vp.SetContent(m.renderChatViewport())
		case domain.UserTypingMsg:
			selUserTyping = true
		default:
		}

		switch msg.Confirmation {
		case domain.MsgDeliveredConfirmed, domain.MsgReadConfirmed:
			for i, ms := range m.msgs {
				if ms.ID == msg.ID {
					m.msgs[i] = msg
					m.vp.SetContent(m.renderChatViewport())
					break
				}
			}
		default:
		}

		return m, tea.Batch(m.handleChatViewportUpdate(msg), m.listenForMessages(), cmd)

	case SentMsg: // the message we'll send gets here once delivered successfully
		m.msgs = append([]*domain.Message{msg}, m.msgs...)
		m.vp.SetContent(m.renderChatViewport())
		m.vp.LineDown(3) // GotoBottom does not work here as intended
		return m, m.handleChatViewportUpdate(msg)

	case timer.TickMsg:
		if m.recvTypingTimer.ID() == msg.ID {
			var cmd tea.Cmd
			m.recvTypingTimer, cmd = m.recvTypingTimer.Update(msg)
			return m, cmd
		}

	case timer.TimeoutMsg:
		if m.recvTypingTimer.ID() == msg.ID {
			selUserTyping = false
			m.recvTypingTimer.Timeout = 3 * time.Second
		}

	}

	return m, tea.Batch(m.handleChatViewportUpdate(msg))
}

func (m ChatViewportModel) View() string {
	return m.vp.View()
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func (m *ChatViewportModel) renderChatViewport() string {
	var sb strings.Builder
	var prevMsgDay int
	l, err := time.LoadLocation("Local")
	if err != nil {
		slog.Error(err.Error())
	}
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
				Render(msg.SentAt.In(l).Format("January 02, 2006"))
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

	var status string
	if msg.SentAt != nil {
		status = "⁎"
	}
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
		for {
			msg := <-m.mb.ch
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
		msgs, meta, err := m.client.GetMessagesAsPageAndMarkAsRead(selUserID, p)
		if err != nil {
			return &errMsg{
				err:  "Unable to fetch initial chat for this user...",
				code: 0,
			}
		}
		return msgPage{msgs, meta}
	}
}

func (m *ChatViewportModel) updateMsgInMsgs(msg *domain.Message) {
	for i, imsg := range m.msgs {
		if imsg.ID == msg.ID {
			if msg.DeliveredAt != nil {
				imsg.DeliveredAt = msg.DeliveredAt
			}
			if msg.ReadAt != nil {
				imsg.ReadAt = msg.ReadAt
			}
			m.msgs[i] = imsg
			break
		}
	}
}

func (m ChatViewportModel) setMsgAsRead(msg *domain.Message) {
	m.client.BT.Run(func(shtdwnCtx context.Context) {
		_ = m.client.SetMsgAsRead(msg) // ignore as we can retry
	})
}
