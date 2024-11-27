package tui

import (
	"github.com/M0hammadUsman/letschat/internal/tui/embed"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type UsageViewportModel struct {
	vp    viewport.Model
	focus bool
}

func NewUsageViewportModel() UsageViewportModel {
	usageFiles := embed.EmbeddedFilesInstance()
	g, _ := glamour.NewTermRenderer(glamour.WithStylesFromJSONBytes(usageFiles.UsageTheme))
	usage, _ := g.Render(string(usageFiles.UsageFile))

	vp := viewport.New(0, 0)
	vp.SetContent(usage)

	return UsageViewportModel{
		vp: vp,
	}
}

func (m UsageViewportModel) Init() tea.Cmd {
	return nil
}

func (m UsageViewportModel) Update(msg tea.Msg) (UsageViewportModel, tea.Cmd) {
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.updateVPDimensions()
		m.vp.GotoBottom()
	}
	switch msg := msg.(type) {
	case tea.MouseButton:
		if msg == tea.MouseButtonWheelUp {
			m.vp.LineUp(5)
		}
		if msg == tea.MouseButtonWheelDown {
			m.vp.LineDown(5)
		}
	}
	return m, nil
}

func (m UsageViewportModel) View() string {
	return m.vp.View()
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func (m *UsageViewportModel) updateVPDimensions() {
	m.vp.Width = usageWidth()
	m.vp.Height = conversationHeight() - 1
}
