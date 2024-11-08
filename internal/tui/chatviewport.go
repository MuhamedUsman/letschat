package tui

import (
	"context"
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"log/slog"
	"slices"
	"strings"
	"time"
)

const (
	infoDialogBox               = "infoDialogBox"
	infoDialogCopyBtn           = "infoDialogCopyBtn"
	infoDialogDelForMeBtn       = "infoDialogDelForMeBtn"
	infoDialogDelForEveryoneBtn = "infoDialogDelForEveryoneBtn"
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
	chatVp             viewport.Model
	msgDialogVp        viewport.Model
	msgs               []*domain.Message
	currPage, lastPage int
	// used to determine the border style of the msg bubble, starting bubble will have a top corner protruding
	startingBubble bool
	// also used in the process of determine the border style of the msg bubble
	prevRenderedMsg *domain.Message
	selUsrID        string
	// currently selected msg (sent one) for info, we'll hide the dialog once the selMsgId is nil
	selMsgId *string
	// current button selection once the msg info dialog in focus,
	// 0 -> CopyBtn | 1 -> DeleteBtn
	selMsgDialogBtn           int // -1 when the selMsgId is nil
	focus                     bool
	fetching                  bool
	recvTypingTimer           timer.Model
	lastTypingStateReceivedAt time.Time
	// only used when msgPage is received
	prevLineCount int
	client        *client.Client
	mb            msgBroadcast
}

func InitialChatViewport(c *client.Client) ChatViewportModel {
	token, ch := c.RecvMsgs.Subscribe()
	return ChatViewportModel{
		chatVp: viewport.New(0, 0),
		//msgDialogVp:     viewport.New(0, 0),
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
		m.chatVp.KeyMap = viewport.DefaultKeyMap()
		m.msgDialogVp.KeyMap = viewport.DefaultKeyMap()
		m.chatVp.MouseWheelEnabled = true
		m.msgDialogVp.MouseWheelEnabled = false
		if m.selMsgId != nil {
			m.msgDialogVp.MouseWheelEnabled = true
			m.chatVp.MouseWheelEnabled = false
		}
	} else {
		m.chatVp.KeyMap = viewport.KeyMap{}
		m.msgDialogVp.KeyMap = viewport.KeyMap{}
		m.chatVp.MouseWheelEnabled = false
		m.msgDialogVp.MouseWheelEnabled = false
	}

	if m.selUsrID != selUserID {
		m.msgs = slices.Delete(m.msgs, 0, len(m.msgs))
		m.msgs = nil
		m.selUsrID = selUserID
		return m, m.getMsgAsPage(1)
	}
	if m.chatVp.AtTop() && !m.fetching {
		if m.lastPage != m.currPage {
			m.fetching = true
			return m, m.getMsgAsPage(m.currPage + 1)
		}
	}
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.chatVp.SetContent(m.renderChatViewport())
		return m, tea.Batch(m.handleChatViewportUpdate(msg), m.handleMsgDialogViewportUpdate(msg))

	case tea.KeyMsg:
		var selMsg *domain.Message
		if m.selMsgId != nil {
			selMsg = m.getSelMsgFromMsgSlice()
		}

		switch msg.String() {
		case "esc", "ctrl+t", "ctrl+f": // once user types or filters convos, hide the dialog
			m.selMsgId = nil
			m.selMsgDialogBtn = -1
		case "tab":
			if selMsg != nil {
				if m.selMsgDialogBtn >= 0 && m.selMsgDialogBtn <= 2 {
					if selMsg.SenderID == m.client.CurrentUsr.ID {
						m.selMsgDialogBtn = (m.selMsgDialogBtn + 1) % 3
					} else {
						m.selMsgDialogBtn = (m.selMsgDialogBtn + 1) % 2
					}
				}
				if m.selMsgDialogBtn > 2 {
					m.selMsgDialogBtn = 0
				}
				m.msgDialogVp.SetContent(m.renderMsgDialogViewport())
			}
		case "left":
			if selMsg != nil {
				if m.selMsgDialogBtn > 0 && m.selMsgDialogBtn <= 2 {
					if selMsg.SenderID == m.client.CurrentUsr.ID {
						m.selMsgDialogBtn--
					} else {
						m.selMsgDialogBtn--
						if m.selMsgDialogBtn < 0 {
							m.selMsgDialogBtn = 0
						}
					}
				}
				m.msgDialogVp.SetContent(m.renderMsgDialogViewport())
			}
		case "right":
			if selMsg != nil {
				if m.selMsgDialogBtn >= 0 && m.selMsgDialogBtn <= 1 {
					if selMsg.SenderID == m.client.CurrentUsr.ID {
						m.selMsgDialogBtn++
						if m.selMsgDialogBtn > 2 {
							m.selMsgDialogBtn = 2
						}
					} else {
						m.selMsgDialogBtn++
						if m.selMsgDialogBtn > 1 {
							m.selMsgDialogBtn = 1
						}
					}
				}
				m.msgDialogVp.SetContent(m.renderMsgDialogViewport())
			}
		case "enter":
			if m.selMsgId != nil {
				if m.selMsgDialogBtn == 0 {
					_ = clipboard.WriteAll(m.getSelMsgFromMsgSlice().Body)
				}
				if m.selMsgDialogBtn == 1 {
					if m.selMsgId != nil {
						return m, m.deleteForMe(*m.selMsgId)
					}
				}
				if m.selMsgDialogBtn == 2 {
					return m, m.deleteForEveryone(*m.selMsgId)
				}
			}
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonRight && msg.Action == tea.MouseActionRelease {
			for _, mesg := range m.msgs {
				if zone.Get(mesg.ID).InBounds(msg) {
					m.selMsgId = &mesg.ID
					m.selMsgDialogBtn = 0
					m.msgDialogVp.SetContent(m.renderMsgDialogViewport())
					m.msgDialogVp.GotoTop() // to remove previous render scroll position
					break
				}
			}
			return m, nil
		}

		if m.selMsgId != nil && msg.Button == tea.MouseButtonLeft {
			if msg.Action == tea.MouseActionPress {
				if zone.Get(infoDialogCopyBtn).InBounds(msg) {
					m.selMsgDialogBtn = 0
				}
				if zone.Get(infoDialogDelForMeBtn).InBounds(msg) {
					m.selMsgDialogBtn = 1
				}
				if zone.Get(infoDialogDelForEveryoneBtn).InBounds(msg) {
					m.selMsgDialogBtn = 2
				}
				m.msgDialogVp.SetContent(m.renderMsgDialogViewport())
			}
		}

		if msg.Action == tea.MouseActionMotion {
			if !zone.Get(infoDialogBox).InBounds(msg) {
				m.selMsgId = nil
				m.selMsgDialogBtn = -1
			}
		}

	case msgPage:
		m.fetching = false
		m.msgs = append(m.msgs, msg.msgs...)
		m.currPage = msg.meta.CurrentPage
		m.lastPage = msg.meta.LastPage
		m.chatVp.SetContent(m.renderChatViewport())
		// Once we set the content it takes us to the top, we want to go to the point where the user were before
		c := m.chatVp.TotalLineCount() - m.prevLineCount
		m.chatVp.LineDown(c)
		// Now update the prev line count
		m.prevLineCount = m.chatVp.TotalLineCount()
		return m, m.handleChatViewportUpdate(msg)

	case *domain.Message:
		var cmd tea.Cmd
		switch msg.Operation {

		case domain.CreateMsg:
			m.msgs = append([]*domain.Message{msg}, m.msgs...)
			m.chatVp.SetContent(m.renderChatViewport())
			m.chatVp.GotoBottom()
			// set it as read also | nil check, if the terminal focus is not supported, just set the msg as read
			if msg.SenderID == selUserID && msg.ReadAt == nil && (terminalFocus == nil || *terminalFocus) {
				t := time.Now()
				msg.ReadAt = &t
				m.setMsgAsRead(msg)
			}

		case domain.UpdateMsg:
			m.updateMsgInMsgs(msg)
			// the above op will update the msgs so we need to rerender
			if m.selMsgId != nil {
				m.chatVp.SetContent(m.renderMsgDialogViewport())
			} else {
				m.chatVp.SetContent(m.renderChatViewport())
			}

		case domain.DeleteMsg:
			m.deleteMsgInMsgs(msg.ID)
			// the deleted msg is selected then:
			if m.selMsgId != nil && *m.selMsgId == msg.ID {
				m.selMsgId = nil
				// rerender to remove the deleted msg
				m.msgDialogVp.SetContent(m.renderMsgDialogViewport())
			}
			prevLineCount := m.chatVp.TotalLineCount()
			m.chatVp.SetContent(m.renderChatViewport())
			currLineCount := m.chatVp.TotalLineCount()
			// for the viewport to go down, not show empty spaces
			if m.chatVp.PastBottom() {
				m.chatVp.GotoBottom()
			} else {
				m.chatVp.LineDown(max(0, prevLineCount-currLineCount))
			}

		case domain.UserTypingMsg:
			selUserTyping = true
		default:
		}

		switch msg.Confirmation {
		case domain.MsgDeliveredConfirmed, domain.MsgReadConfirmed:
			for i, ms := range m.msgs {
				if ms.ID == msg.ID {
					m.msgs[i] = msg
					m.chatVp.SetContent(m.renderChatViewport())
					break
				}
			}
		default:
		}

		return m, tea.Batch(m.handleChatViewportUpdate(msg), m.handleMsgDialogViewportUpdate(msg), m.listenForMessages(), cmd)

	case SentMsg: // the message we'll send gets here once delivered successfully
		m.msgs = append([]*domain.Message{msg}, m.msgs...)
		m.chatVp.SetContent(m.renderChatViewport())
		m.chatVp.LineDown(3) // GotoBottom does not work here as intended
		return m, m.handleChatViewportUpdate(msg)

	case deleteMsgSuccess:
		m.selMsgId = nil
		m.deleteMsgInMsgs(string(msg))
		prevLineCount := m.chatVp.TotalLineCount()
		m.msgDialogVp.SetContent(m.renderMsgDialogViewport())
		m.chatVp.SetContent(m.renderChatViewport())
		currLineCount := m.chatVp.TotalLineCount()
		// for the viewport to go down, not show empty spaces
		if m.chatVp.PastBottom() {
			m.chatVp.GotoBottom()
		} else {
			m.chatVp.LineDown(max(0, prevLineCount-currLineCount))
		}

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

	return m, tea.Batch(m.handleChatViewportUpdate(msg), m.handleMsgDialogViewportUpdate(msg))
}

func (m ChatViewportModel) View() string {
	// show dialog box once a msg is selected for info details
	if m.selMsgId != nil {
		v := m.msgDialogVp.View()
		// mark it so once out of its bounds we hide it
		v = zone.Mark(infoDialogBox, v)
		return v
	} else {
		return m.chatVp.View()
	}
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
		// chat bubble
		cb := cb.
			Align(align).
			Render(m.renderBubbleWithStatusInfo(msg))
		sb.WriteString("\n")
		sb.WriteString(cb)
	}
	return sb.String()
}

func (m ChatViewportModel) renderMsgDialogViewport() string {
	// get the message
	infoMsg := m.getSelMsgFromMsgSlice()
	if infoMsg == nil {
		return ""
	}
	var head, body, btnContainer, foot string
	headTxt := "YOU"
	if infoMsg.SenderID == selUserID {
		headTxt = selUsername
	}
	head = msgInfoHeaderStyle.Render(headTxt)
	body = msgInfoBodyStyle.
		Width(chatWidth() - msgInfoBodyStyle.GetHorizontalFrameSize()).
		Render(infoMsg.Body)

	copyBtn := zone.Mark(infoDialogCopyBtn, renderCopyBtn(m.selMsgDialogBtn))
	delBtn := zone.Mark(infoDialogDelForMeBtn, renderDeleteBtn(false, "DELETE"))
	delForMeFocus := false
	delForEveryoneFocus := false
	if m.selMsgDialogBtn == 1 {
		delForMeFocus = true
	}
	if m.selMsgDialogBtn == 2 {
		delForEveryoneFocus = true
	}
	delForMeBtn := zone.Mark(infoDialogDelForMeBtn, renderDeleteBtn(delForMeFocus, "DELETE FOR ME"))
	delForEveryoneBtn := zone.Mark(infoDialogDelForEveryoneBtn, renderDeleteBtn(delForEveryoneFocus, "DELETE FOR EVERYONE"))

	if m.selMsgDialogBtn == 0 {
		btnContainer = msgInfoContainerBtn.Render(copyBtn, delBtn)
	} else {
		btnContainer = msgInfoContainerBtn.Render(copyBtn, delForMeBtn, delForEveryoneBtn)
		if infoMsg.SenderID != m.client.CurrentUsr.ID {
			btnContainer = msgInfoContainerBtn.Render(copyBtn, delForMeBtn)
		}
	}

	status := renderInfoMsgStatus(infoMsg)
	foot = msgInfoFooterStyle.Render(status)

	return head + body + btnContainer + foot
}

func renderInfoMsgStatus(msg *domain.Message) string {
	l, err := time.LoadLocation("Local")
	if err != nil {
		slog.Error(err.Error())
	}
	f := "02-Jan-2006 | 3:04 PM"
	var sb strings.Builder
	if msg.SentAt != nil {
		sb.WriteString(fmt.Sprintf("✓     %v", msg.SentAt.In(l).Format(f)))
		sb.WriteString("\n\n")
	}
	if msg.DeliveredAt != nil {
		sb.WriteString(fmt.Sprintf("✓✓    %v", msg.DeliveredAt.In(l).Format(f)))
		sb.WriteString("\n\n")
	}
	if msg.ReadAt != nil {
		sb.WriteString(fmt.Sprintf("✓✓✓   %v", msg.ReadAt.In(l).Format(f)))
	}
	return sb.String()
}

func renderCopyBtn(selBtnIdx int) string {
	bg := primaryColor
	fg := primaryContrastColor
	if selBtnIdx != 0 {
		bg = darkGreyColor
		fg = lightGreyColor
	}
	return msgInfoBtnStyle.
		Background(bg).
		Foreground(fg).
		Padding(0, 3).
		Render("COPY")
}

func renderDeleteBtn(focus bool, btnTxt string) string {
	bg := dangerColor
	fg := whiteColor
	if !focus {
		bg = darkGreyColor
		fg = lightGreyColor
	}
	return msgInfoBtnStyle.
		Background(bg).
		Foreground(fg).
		Render(btnTxt)
}

// getSelectedMsgFromMsgSlice
func (m *ChatViewportModel) getSelMsgFromMsgSlice() *domain.Message {
	for _, msg := range m.msgs {
		if m.selMsgId != nil && msg.ID == *m.selMsgId {
			return msg
		}
	}
	return nil
}

func (m *ChatViewportModel) renderBubbleWithStatusInfo(msg *domain.Message) string {
	txtWidth := min(chatWidth()-20, lipgloss.Width(msg.Body)+2)
	bubbleStyle := chatBubbleLStyle.Width(txtWidth)
	// if prev msg sender is the same as current msg sender, do not show the protruding edge
	if m.prevRenderedMsg != nil &&
		m.prevRenderedMsg.SenderID == msg.SenderID &&
		m.prevRenderedMsg.SentAt.Day() == msg.SentAt.Day() {
		bubbleStyle = bubbleStyle.BorderStyle(lipgloss.RoundedBorder())
	}
	bubble := bubbleStyle.Render(msg.Body)
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
		bubbleStyle = chatBubbleRStyle.Width(txtWidth)
		if m.prevRenderedMsg != nil &&
			m.prevRenderedMsg.SenderID == msg.SenderID &&
			m.prevRenderedMsg.SentAt.Day() == msg.SentAt.Day() {
			bubbleStyle = bubbleStyle.BorderStyle(lipgloss.RoundedBorder())
		}
		// set the prevRenderedMsg to current msg senderId, once we are done with its use
		m.prevRenderedMsg = msg
		bubble = bubbleStyle.Render(msg.Body)
		// mark the msg with zone on the right side so we can pick these up using mouse clicks
		bubble = zone.Mark(msg.ID, bubble)
		sentAt = sentAt.Foreground(primaryColor)
		return lipgloss.JoinHorizontal(lipgloss.Center, status, " ", sentAt.Render(), " ", bubble)
	}
	// set the prevRenderedMsg to current msg senderId, once we are done with its use
	m.prevRenderedMsg = msg
	// mark the msg with zone on the left side so we can pick these up using mouse clicks
	bubble = zone.Mark(msg.ID, bubble)
	return lipgloss.JoinHorizontal(lipgloss.Center, bubble, " ", sentAt.Render())
}

func (m *ChatViewportModel) updateDimensions() {
	w := chatWidth()
	h := chatHeight() - (chatHeaderHeight + chatTextareaHeight)
	m.chatVp.Width = w
	m.msgDialogVp.Width = w
	m.chatVp.Height = h
	m.msgDialogVp.Height = h
}

func (m *ChatViewportModel) handleChatViewportUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.chatVp, cmd = m.chatVp.Update(msg)
	return cmd
}

func (m *ChatViewportModel) handleMsgDialogViewportUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.msgDialogVp, cmd = m.msgDialogVp.Update(msg)
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

// once the message is deleted, this deletes it from the msg slice, if not exists -> NOOP
func (m *ChatViewportModel) deleteMsgInMsgs(msgId string) {
	for i, mesg := range m.msgs {
		if mesg.ID == msgId {
			// [inclusive:exclusive]
			m.msgs = append(m.msgs[:i], m.msgs[i+1:]...)
			break
		}
	}
}

func (m ChatViewportModel) deleteForMe(msgId string) tea.Cmd {
	return func() tea.Msg {
		if err := m.client.DeleteMsgForMe(msgId); err != nil {
			return &errMsg{
				err:  "Unable to delete this message",
				code: 0,
			}
		}
		return deleteMsgSuccess(msgId)
	}
}

func (m ChatViewportModel) deleteForEveryone(msgId string) tea.Cmd {
	t := time.Now()
	delMsg := &domain.Message{
		ID:           msgId,
		SenderID:     m.client.CurrentUsr.ID,
		ReceiverID:   selUserID,
		SentAt:       &t,
		Operation:    domain.DeleteMsg,
		Confirmation: 0,
	}
	return func() tea.Msg {
		if err := m.client.DeleteMsgForEveryone(delMsg); err != nil {
			return &errMsg{
				err:  "Unable to delete this message from the receiver",
				code: 0,
			}
		}
		return deleteMsgSuccess(msgId)
	}
}
