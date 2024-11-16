package tui

import (
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	conversationSearchBar = "conversationSearchBar"
	conversationContainer = "conversationContainer"
)

type convosBroadcast struct {
	ch    <-chan client.Convos
	token int
}

type ConversationModel struct {
	conversationList list.Model
	// if there is a redirect from the discover tab then we add the selected user here,
	// so if there is dynamic changes to the conversation list, we first check if the selDiscUser is not nil
	// then we add that user in the top of the list
	selDiscUserConvo *domain.Conversation
	// there is no built-in functionality for list focus as far as I scanned the docs, also see
	// getConversationListKeyMap, this will still update the model but make it look out of focus
	focus  bool
	client *client.Client
	cb     convosBroadcast
}

type conversationItem struct{ id, selConvoUsrId, title, status, latestMsg string }

func (i conversationItem) Title() string { return zone.Mark(i.id, fmt.Sprint(i.title, i.status)) }
func (i conversationItem) FilterValue() string {
	return zone.Mark(i.id, fmt.Sprintf("%v|%v", i.title, i.selConvoUsrId))
}
func (i conversationItem) ConvoID() string     { return i.selConvoUsrId }
func (i conversationItem) Description() string { return i.latestMsg }

func InitialConversationModel(c *client.Client) ConversationModel {
	m := list.New(nil, getDelegateWithCustomStyling(), 0, 0)
	m = applyCustomConversationListStyling(m)
	m.FilterInput = newConversationTxtInput("Filter by name...")
	m.KeyMap = getConversationListKeyMap(true)
	m.SetStatusBarItemName("Conversation", "Conversations")
	m.SetShowFilter(false)
	m.SetShowHelp(false)
	m.SetShowTitle(false)
	m.SetShowPagination(false)
	m.StatusMessageLifetime = 2 * time.Second

	token, ch := c.Conversations.Subscribe()
	return ConversationModel{
		conversationList: m,
		focus:            true,
		client:           c,
		cb: convosBroadcast{
			ch:    ch,
			token: token,
		},
	}
}

func (m ConversationModel) Init() tea.Cmd {
	return m.getConversations()
}

func (m ConversationModel) Update(msg tea.Msg) (ConversationModel, tea.Cmd) {
	if m.focus {
		m.conversationList.KeyMap = getConversationListKeyMap(true)
	} else {
		m.conversationList.KeyMap = getConversationListKeyMap(false)
	}

	if len(m.conversationList.Items()) > 0 || m.conversationList.FilterState() == list.Filtering {
		m.conversationList.SetShowStatusBar(true)
	} else {
		m.conversationList.SetShowStatusBar(false)
	}

	// Remove the selDiscUserConvo as the user changed the convo selection
	// if there is a message sent on this convo the dynamic fetch "getConversations()" will show this convo
	if m.selDiscUserConvo != nil && m.selDiscUserConvo.UserID != selUserID {
		m.selDiscUserConvo = nil
		m.conversationList.RemoveItem(0)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateConversationWindowSize()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.focus {
				selUserID = m.getSelConvoUsrID()
				selUsername = m.getSelConvoUsername()
			}
			return m, nil
		case "ctrl+f":
			return m, tea.Batch(m.conversationList.FilterInput.Focus(), m.handleConversationListUpdate(msg))
		case "ctrl+t":
			m.conversationList.FilterInput.Blur()
		case "ctrl+s": // create the convo if the selDiscUserConvo does not exist
			if m.selDiscUserConvo != nil {
				return m, m.createConvoIfNotExist()
			}
		case "esc":
			m.conversationList.FilterInput.Blur()
		}

	case tea.MouseMsg:
		if zone.Get(conversationContainer).InBounds(msg) {
			m.focus = true
		} else {
			m.focus = false
		}
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			if zone.Get(conversationContainer).InBounds(msg) {
				m.conversationList.CursorDown()
			}
		case tea.MouseButtonWheelUp:
			if zone.Get(conversationContainer).InBounds(msg) {
				m.conversationList.CursorUp()
			}
		case tea.MouseButtonLeft:
			for i, listItem := range m.conversationList.VisibleItems() {
				v, _ := listItem.(conversationItem)
				// Check each item to see if it's in bounds.
				if zone.Get(v.id).InBounds(msg) {
					// If so, select it in the list.
					m.conversationList.Select(i)
					selUserID = m.getSelConvoUsrID()
					selUsername = m.getSelConvoUsername()
					break
				}
			}
			if zone.Get(conversationSearchBar).InBounds(msg) {
				return m, m.handleConversationListUpdate(tea.KeyMsg{Type: tea.KeyCtrlF})
			} else {
				m.conversationList.FilterInput.Blur()
				return m, m.handleConversationListUpdate(tea.KeyMsg{Type: tea.KeyEsc})
			}
		default:
		}

	case []list.Item:
		// m.getConversations() so we can read again for updated conversations
		return m, tea.Batch(
			m.conversationList.SetItems(msg),
			spinnerResetCmd,
			m.getConversations(),
			m.conversationList.NewStatusMessage("Updated Conversations"),
		)

	case selDiscUserMsg:
		items := m.conversationList.Items()
		// if the selected user is already in the convo list
		for i, item := range items {
			usrId := extractUsrId(item.FilterValue())
			if usrId == msg.id {
				m.conversationList.Select(i)
				selUserID = m.getSelConvoUsrID()
				selUsername = m.getSelConvoUsername()
				return m, nil
			}
		}
		t := time.Now()
		convo := &domain.Conversation{
			UserID:     msg.id,
			Username:   msg.name,
			UserEmail:  msg.email,
			LastOnline: &t,
		}
		m.selDiscUserConvo = convo
		selUserID = msg.id
		selUsername = msg.name
		renderState := false
		if m.client.WsConnState.Get() == client.Connected {
			renderState = true
		}
		cmd := m.conversationList.InsertItem(0, populateConvoItem(0, convo, renderState))
		return m, cmd

	case createdConvoMsg:
		renderState := false
		if m.client.WsConnState.Get() == client.Connected {
			renderState = true
		}
		// after the convo is created in the db, we remove it from here, now the getConversation() will show it
		m.selDiscUserConvo = nil
		cmd := m.conversationList.InsertItem(0, populateConvoItem(0, msg, renderState))
		return m, cmd
	}

	return m, tea.Batch(m.handleConversationListUpdate(msg))
}

func (m ConversationModel) View() string {
	searchBarStyle := conversationSearchBarStyle.Width(conversationWidth() - 4)
	if m.conversationList.FilterInput.Focused() {
		searchBarStyle = conversationActiveSearchBarStyle.Width(conversationWidth() - 4)
	}
	s := searchBarStyle.Render(m.conversationList.FilterInput.View())
	s = zone.Mark(conversationSearchBar, s)
	searchAndList := lipgloss.JoinVertical(lipgloss.Left, s, m.conversationList.View())
	convos := conversationContainerStyle.Width(conversationWidth()).Height(conversationHeight()).Render(searchAndList)
	return zone.Mark(conversationContainer, convos)
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func newConversationTxtInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.TextStyle = lipgloss.NewStyle().Foreground(primaryColor)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(primaryColor)
	ti.CharLimit = 64
	ti.Prompt = ""
	ti.Placeholder = placeholder
	return ti
}

func getDelegateWithCustomStyling() list.ItemDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(primaryColor).
		BorderForeground(primaryColor).
		BorderStyle(lipgloss.ThickBorder())

	d.Styles.NormalDesc = d.Styles.NormalDesc.
		BorderForeground(primaryColor)

	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(primarySubtleDarkColor).
		BorderForeground(primaryColor).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(primaryColor)

	return d
}

func applyCustomConversationListStyling(m list.Model) list.Model {
	m.Styles.StatusBar = m.Styles.StatusBar.
		Foreground(primaryColor).
		MarginTop(1)
	m.Styles.NoItems = m.Styles.NoItems.
		Margin(1, 1).
		Foreground(primarySubtleDarkColor).
		SetString("We've Found")
	return m
}

func getConversationListKeyMap(enabled bool) list.KeyMap {
	km := list.DefaultKeyMap()
	km.Filter = key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "filter by name"))
	kb := key.NewBinding() // disable keybindings when out of focus
	km.Quit = kb           // default
	km.ForceQuit = kb      // default
	if !enabled {
		km.CursorUp = kb
		km.CursorDown = kb
		km.NextPage = kb
		km.PrevPage = kb
		km.GoToStart = kb
		km.GoToEnd = kb
		km.ShowFullHelp = kb
	}
	return km
}

func populateConvoItem(i int, convo *domain.Conversation, renderState bool) conversationItem {
	id := "item_" + strconv.Itoa(i)
	var latestMsg string
	if convo.LatestMsg != nil {
		latestMsg = *convo.LatestMsg
	} else {
		latestMsg = "..."
	}
	var s string
	if renderState {
		s = renderStateInfo(convo)
	}
	item := conversationItem{id, convo.UserID, convo.Username, s, latestMsg}
	return item
}

func renderStateInfo(convo *domain.Conversation) string {
	t := convo.LastOnline
	if t == nil {
		return conversationOnlineIndicator
	}
	// TODO: Show last seen instead of ❄️
	return ""
}

func containsSelConvo(convos []*domain.Conversation, convoUsrID string) bool {
	return slices.ContainsFunc(convos, func(conversation *domain.Conversation) bool {
		return convoUsrID == conversation.UserID
	})
}

func (m *ConversationModel) handleConversationListUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.conversationList, cmd = m.conversationList.Update(msg)
	return cmd
}

func (m *ConversationModel) updateConversationWindowSize() {
	m.conversationList.SetSize(tabGapLeftWidth-4, terminalHeight-7)
	m.conversationList.SetDelegate(getDelegateWithCustomStyling())
	m.conversationList.FilterInput.Width = tabGapLeftWidth - 9
}

func (m ConversationModel) getConversations() tea.Cmd {
	return func() tea.Msg {
		ctx := m.client.BT.GetShtdwnCtx()
		c := make([]list.Item, 0)
		convos := <-m.cb.ch
		// we check if the ctx is done after unblocking
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		for i, convo := range convos {
			renderState := false
			if m.client.WsConnState.Get() == client.Connected {
				renderState = true
			}
			/*if i == 0 { // add it to first idx
				if m.selDiscUserConvo != nil && !containsSelConvo(convos, m.selDiscUserConvo.UserID) {
					c = append(c, populateConvoItem(-1, m.selDiscUserConvo, renderState))
				}
			}*/
			item := populateConvoItem(i, convo, renderState)
			c = append(c, item)
		}
		return c
	}
}

func (m ConversationModel) getSelConvoUsrID() string {
	// We hold the actual "title|selectedConvoUsrID" in the filter value
	if m.conversationList.SelectedItem() == nil {
		return ""
	}
	fv := m.conversationList.SelectedItem().FilterValue()
	return extractUsrId(fv)
}

func extractUsrId(s string) string {
	idWithSomeStylingTxt := strings.Split(s, "|")[1] // 033d13fa-b6d8-43db-b288-34fe801570e6[1012z
	return idWithSomeStylingTxt[:36]                 // 033d13fa-b6d8-43db-b288-34fe801570e6
}

func (m ConversationModel) getSelConvoUsername() string {
	if m.conversationList.SelectedItem() == nil {
		return ""
	}
	fv := m.conversationList.SelectedItem().FilterValue()
	return strings.Split(fv, "|")[0]
}

func (m ConversationModel) createConvoIfNotExist() tea.Cmd {
	return func() tea.Msg {
		createdConvo, err := m.client.CreateConvoIfNotExist(m.selDiscUserConvo)
		if err != nil {
			return &errMsg{err: "Unable to create the conversation", code: 0}
		}
		if createdConvo != nil {
			return createdConvoMsg(createdConvo)
		}
		return nil
	}
}
