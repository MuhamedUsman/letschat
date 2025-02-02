package tui

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"net/http"
	"strconv"
	"unicode/utf8"
)

const (
	discoverSearchBar = "discoverSearchBar"
	discoverTable     = "discoverTable"
)

type DiscoverModel struct {
	searchTxtInput textinput.Model
	table          table.Model
	tableUsrIDs    []string // user ids related to each row
	metadata       domain.Metadata
	focusIdx       int // 0 -> Search, 1 -> Table
	focus          bool
	placeholder    string
	client         *client.Client
}

func InitialDiscoverModel(c *client.Client) DiscoverModel {
	m := DiscoverModel{
		placeholder: "Bashbunni OR bashbunni@bunnibrain.letschat",
		table:       newDiscoverTable(),
		focus:       true,
		client:      c,
	}
	m.searchTxtInput = newDiscoverTxtInput(m.placeholder)
	m.searchTxtInput.Cursor = newDiscoverCursor()
	return m
}

func (m DiscoverModel) Init() tea.Cmd {
	return nil
}

func (m DiscoverModel) Update(msg tea.Msg) (DiscoverModel, tea.Cmd) {
	m.focusAccordingly()
	m.handleDiscoverTableHeight()
	// Fetching more records if the user is in the end of the table
	curPage := m.metadata.CurrentPage
	if m.table.Cursor() == len(m.table.Rows())-5 && curPage < m.metadata.LastPage && ioStatus == "" {
		ioStatus = "Fetching more"
		m.table.MoveDown(1)
		return m, tea.Batch(m.searchUser(m.searchTxtInput.Value(), curPage+1), spinnerSpinCmd)
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+f":
			m.focusIdx = 0
			return m, m.focusAccordingly()
		case "up", "down":
			m.focusIdx = 1
		case "enter":
			if m.focusIdx == 0 && m.focus {
				if utf8.RuneCountInString(m.searchTxtInput.Value()) > 0 {
					m.table.SetRows(nil) // clearing any previous records
					ioStatus = "Searching"
					return m, tea.Batch(spinnerSpinCmd, m.searchUser(m.searchTxtInput.Value(), 1))
				}
			}
			if m.focusIdx == 1 && m.focus {
				selRow := m.table.SelectedRow()
				selMsg := selDiscUserMsg{
					id:    m.tableUsrIDs[m.table.Cursor()-1],
					name:  selRow[1],
					email: selRow[2],
				}
				return m, func() tea.Msg { return selMsg } // cmd
			}
		}

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			if zone.Get(discoverSearchBar).InBounds(msg) {
				m.focusIdx = 0
				return m, m.focusAccordingly()
			}
			if zone.Get(discoverTable).InBounds(msg) {
				m.focusIdx = 1
				return m, m.focusAccordingly()
			}
		case tea.MouseButtonWheelDown:
			if zone.Get(discoverTable).InBounds(msg) {
				m.focusIdx = 1
				m.table.MoveDown(1)
				return m, m.focusAccordingly()
			}
		case tea.MouseButtonWheelUp:
			if zone.Get(discoverTable).InBounds(msg) {
				m.focusIdx = 1
				m.table.MoveUp(1)
				return m, m.focusAccordingly()
			}
		default:
		}

	case tableResp:
		m.table.SetRows(msg.rows)
		m.tableUsrIDs = msg.rowsIds
		m.metadata = msg.metadata
		if m.metadata.CurrentPage == 1 {
			m.table.SetCursor(0)
		}
		if len(m.table.Rows()) > 0 {
			m.focusIdx = 1
			return m, tea.Batch(m.focusAccordingly(), spinnerResetCmd)
		}
		m.focusIdx = 0
		return m, spinnerResetCmd
	}

	return m, tea.Batch(m.handleDiscoverSearchTxtInput(msg), m.handleDiscoverTableUpdate(msg))
}

func (m *DiscoverModel) View() string {
	bar := activeDiscoverBar.Render(m.searchTxtInput.View())
	bar = zone.Mark(discoverSearchBar, bar)
	var s string
	if len(m.table.Rows()) > 0 {
		s = discoverTableStyle.Render(m.table.View())
		s = zone.Mark(discoverTable, s)
	} else {
		s = bunny
		s = lipgloss.PlaceVertical(terminalHeight-10, lipgloss.Center, s)
	}
	s = lipgloss.JoinVertical(lipgloss.Center, bar, s)
	return lipgloss.PlaceHorizontal(terminalWidth-2, lipgloss.Center, s)
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func newDiscoverTxtInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.TextStyle = lipgloss.NewStyle().Foreground(primaryColor)
	ti.Focus()
	ti.CharLimit = 64
	ti.Prompt = ""
	ti.Placeholder = placeholder
	return ti
}

func newDiscoverCursor() cursor.Model {
	c := cursor.New()
	cStyle := lipgloss.NewStyle().Foreground(primaryColor)
	c.Style = cStyle
	c.TextStyle = cStyle
	return c
}

func newDiscoverTable() table.Model {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		Foreground(primaryColor).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(primaryContrastColor).
		Background(primaryColor).
		Bold(false)

	cols := []table.Column{
		{Title: "#", Width: 6},
		{Title: "Name", Width: 30},
		{Title: "Email", Width: 45},
		{Title: "Joined Since", Width: 20},
	}
	t := table.New(table.WithColumns(cols))
	t.SetStyles(s)
	return t
}

func (m *DiscoverModel) handleDiscoverSearchTxtInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.searchTxtInput, cmd = m.searchTxtInput.Update(msg)
	return cmd
}

func (m *DiscoverModel) handleDiscoverTableUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return cmd
}

func (m *DiscoverModel) handleDiscoverTableHeight() {
	h := terminalHeight - 12
	m.table.SetHeight(h)
}

func (m *DiscoverModel) focusAccordingly() tea.Cmd {
	var cmd tea.Cmd
	if m.focus {
		if m.focusIdx == 0 {
			cmd = m.searchTxtInput.Focus()
			m.table.Blur()
		} else if m.focusIdx == 1 {
			m.table.Focus()
			m.searchTxtInput.Blur()
		}
	} else {
		m.searchTxtInput.Blur()
		m.table.Blur()
		m.focusIdx = -1
	}
	return cmd
}

type tableResp struct {
	rows     []table.Row
	rowsIds  []string
	metadata domain.Metadata
}

func (m DiscoverModel) searchUser(query string, page int) tea.Cmd {
	return func() tea.Msg {
		resp, code, err := m.client.SearchUser(query, page)
		if code == http.StatusUnauthorized {
			return requireAuthMsg{}
		}
		if err != nil {
			return &errMsg{err: err.Error(), code: code}
		}
		rows := m.table.Rows()
		ids := m.tableUsrIDs
		l := len(rows)
		for _, u := range resp.Users {
			// do not show the current user in the results
			if u.ID == m.client.CurrentUsr.ID {
				continue
			}
			cell := table.Row{strconv.Itoa(l + 1), u.Name, u.Email, u.CreatedAt.Format("January 2006")}
			l++
			rows = append(rows, cell)
			ids = append(ids, u.ID)
		}
		m.table.SetRows(rows)
		return tableResp{
			rows:     rows,
			rowsIds:  ids,
			metadata: resp.Metadata,
		}
	}
}
