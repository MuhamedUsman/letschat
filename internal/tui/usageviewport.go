package tui

import (
	"github.com/M0hammadUsman/letschat/internal/tui/embed"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"log/slog"
)

type UsageViewportModel struct {
	vp         viewport.Model
	usageFiles *embed.EmbeddedFiles
	focus      bool
}

func NewUsageViewportModel() UsageViewportModel {
	vp := viewport.New(50, 30)
	vp.MouseWheelEnabled = true
	return UsageViewportModel{
		vp:         vp,
		usageFiles: embed.EmbeddedFilesInstance(),
	}
}

func (m UsageViewportModel) Init() tea.Cmd {
	return nil
}

func (m UsageViewportModel) Update(msg tea.Msg) (UsageViewportModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g, err := glamour.NewTermRenderer(glamour.WithStylesFromJSONBytes(m.usageFiles.UsageTheme))
		if err != nil {
			slog.Error(err.Error())
		}
		md, err := g.Render(string(m.usageFiles.UsageFile))
		if err != nil {
			slog.Error(err.Error())
		}
		m.vp.SetContent(md)

	case tea.KeyMsg:
		switch msg.String() {

		}
	}
	return m, nil
}

func (m UsageViewportModel) View() string {
	return m.vp.View()
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

/*func (m *UsageViewportModel) updateVPDimensions() {
	m.msgVP.Width = lipgloss.Width(s) + 4
	m.msgVP.Height = terminalHeight - 4
}*/
