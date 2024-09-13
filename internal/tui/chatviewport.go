package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type ChatViewport struct {
	viewport viewport.Model
	messages []string
}

func (ChatViewport) Init() tea.Cmd {
	//TODO implement me
	panic("implement me")
}

func (ChatViewport) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//TODO implement me
	panic("implement me")
}

func (ChatViewport) View() string {
	//TODO implement me
	panic("implement me")
}
