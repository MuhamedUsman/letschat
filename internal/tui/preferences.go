package tui

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

const (
	updateProfile = "updateProfile"
	usageVp       = "usageVp"
)

type PreferencesModel struct {
	up      UpdateProfileModel
	usageVp UsageViewportModel
	focus   bool
	client  *client.Client
}

func NewPreferencesModel(c *client.Client) PreferencesModel {
	return PreferencesModel{
		up:      NewUpdateProfileModel(c),
		usageVp: NewUsageViewportModel(),
		client:  c,
	}
}

func (m PreferencesModel) Init() tea.Cmd {
	return tea.Batch(m.up.Init(), m.usageVp.Init())
}

func (m PreferencesModel) Update(msg tea.Msg) (PreferencesModel, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.up.focus = m.focus
	case tea.MouseMsg:
		m.up.focus = false
		m.usageVp.focus = false
		if zone.Get(updateProfile).InBounds(msg) {
			m.up.focus = true
		}
		if zone.Get(usageVp).InBounds(msg) {
			m.usageVp.focus = true
		}
	}
	return m, tea.Batch(m.handleUsageViewportUpdate(msg), m.handleUpdateProfileModelUpdate(msg))
}

func (m PreferencesModel) View() string {
	d := verticalDivider.Height(conversationHeight()).Render()
	upView := zone.Mark(updateProfile, m.up.View())
	usageVpView := zone.Mark(usageVp, m.usageVp.View())
	return lipgloss.JoinHorizontal(lipgloss.Left, upView, d, usageVpView)
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func (m *PreferencesModel) handleUpdateProfileModelUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.up, cmd = m.up.Update(msg)
	return cmd
}

func (m *PreferencesModel) handleUsageViewportUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.usageVp, cmd = m.usageVp.Update(msg)
	return cmd
}
