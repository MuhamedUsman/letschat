package tui

import (
	"github.com/M0hammadUsman/letschat/internal/tui/embed"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type UsageViewportModel struct {
	vp    viewport.Model
	usage string
	focus bool
}

func NewUsageViewportModel() UsageViewportModel {
	usageFiles := embed.EmbeddedFilesInstance()
	g, _ := glamour.NewTermRenderer(glamour.WithStylesFromJSONBytes(usageFiles.UsageTheme), glamour.WithEmoji())
	usage, _ := g.Render(string(usageFiles.UsageFile))
	vp := viewport.New(0, 0)
	vp.SetContent(usage)

	return UsageViewportModel{
		vp:    vp,
		usage: usage,
	}
}

func (m UsageViewportModel) Init() tea.Cmd {
	return nil
}

func (m UsageViewportModel) Update(msg tea.Msg) (UsageViewportModel, tea.Cmd) {
	m.vp.MouseWheelEnabled = m.focus
	m.vp.KeyMap = viewport.KeyMap{}
	if m.focus {
		m.vp.KeyMap = viewport.DefaultKeyMap()
	} else {
		m.vp.GotoTop()
	}
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.vp.Width = usageWidth()
		m.vp.Height = conversationHeight() - 1
		m.vp.SetContent(m.usage)
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m UsageViewportModel) View() string {
	return m.vp.View()
}
