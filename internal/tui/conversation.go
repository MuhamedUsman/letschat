package tui

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"net/http"
	"strconv"
	"time"
)

const (
	conversationSearchBar = "conversationSearchBar"
	conversationContainer = "conversationContainer"
)

type ConversationModel struct {
	conversationList list.Model
	// there is no built-in functionality for list focus as far as I scanned the docs, also see
	// getConversationListKeyMap, this will still update the model but make it look out of focus
	focus          bool
	fetchingConvos bool
	client         *client.Client
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type conversationItem struct{ id, title, desc string }

func (i conversationItem) Title() string       { return zone.Mark(i.id, i.title) }
func (i conversationItem) Description() string { return i.desc }
func (i conversationItem) FilterValue() string { return zone.Mark(i.id, i.title) }

func InitialConversationModel(c *client.Client) ConversationModel {
	m := list.New(nil, getDelegateWithCustomStyling(), 0, 0)
	m = applyCustomConversationListStyling(m)
	m.FilterInput = newConversationTxtInput("Filter by name...")
	m.KeyMap = getConversationListKeyMap(true)
	m.SetStatusBarItemName("Conversation", "Conversations")
	m.DisableQuitKeybindings()
	m.SetShowFilter(false)
	m.SetShowHelp(false)
	m.SetShowTitle(false)
	m.SetShowPagination(false)
	m.StatusMessageLifetime = 2 * time.Second
	return ConversationModel{
		conversationList: m,
		focus:            true,
		client:           c,
	}
}

func (m ConversationModel) Init() tea.Cmd {
	return nil
}

func (m ConversationModel) Update(msg tea.Msg) (ConversationModel, tea.Cmd) {

	if len(m.conversationList.Items()) == 0 && !m.fetchingConvos {
		m.fetchingConvos = true
		ioStatus = "Getting conversations"
		return m, tea.Batch(m.getConversations(), spinnerSpinCmd)
	}

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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateConversationWindowSize()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+f":
			return m, tea.Batch(m.conversationList.FilterInput.Focus(), m.handleConversationListUpdate(msg))
		case "ctrl+t":
			m.conversationList.FilterInput.Blur()
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
		m.fetchingConvos = false
		return m, tea.Batch(m.conversationList.SetItems(msg), spinnerResetCmd)
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
	if !enabled {
		kb := key.NewBinding() // disable keybindings when out of focus
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

func (m *ConversationModel) handleConversationSearchTxtInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.conversationList.FilterInput, cmd = m.conversationList.FilterInput.Update(msg)
	return cmd
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
		time.Sleep(3 * time.Second)
		convos, err, code := m.client.GetConversations()
		if code == http.StatusUnauthorized {
			return requireAuthMsg{}
		}
		if err != nil {
			return errMsg{err: err.Error(), code: code}
		}
		c := make([]list.Item, 0)
		for i, convo := range convos {
			id := "item_" + strconv.Itoa(i)
			item := conversationItem{id, convo.DisplayName, convo.LastOnline.Format(time.Kitchen)}
			c = append(c, list.Item(item))
		}
		return c
	}
}
