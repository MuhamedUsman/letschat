package tui

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PreferencesModel struct {
	up      UpdateProfileModel
	usageVp viewport.Model
	client  *client.Client
}

func NewPreferencesModel(c *client.Client) PreferencesModel {
	return PreferencesModel{
		up:      NewUpdateProfileModel(c),
		usageVp: viewport.New(0, 0),
		client:  c,
	}
}

func (m PreferencesModel) Init() tea.Cmd {
	return nil
}

func (m PreferencesModel) Update(msg tea.Msg) (PreferencesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		}
	}
	return m, tea.Batch(m.handleUsageViewportUpdate(msg), m.handleUpdateProfileModelUpdate(msg))
}

func (m PreferencesModel) View() string {
	d := verticalDivider.Height(conversationHeight()).Render()
	return lipgloss.JoinHorizontal(lipgloss.Left, m.up.View(), d, m.usageVp.View())
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
