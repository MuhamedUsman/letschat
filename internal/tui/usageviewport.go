package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type UsageViewportModel struct {
	vp    viewport.Model
	focus bool
}

func NewUsageViewportModel() UsageViewportModel {
	return UsageViewportModel{
		vp: viewport.New(0, 0),
	}
}

func (m UsageViewportModel) Init() tea.Cmd {
	return nil
}

func (m UsageViewportModel) Update(msg tea.Msg) (UsageViewportModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		}
	}
	return m, nil
}

func (m UsageViewportModel) View() string {
	return m.vp.View()
}
