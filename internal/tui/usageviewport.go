package tui

import (
	"github.com/MuhamedUsman/letschat/internal/tui/embed"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type UsageViewportModel struct {
	vp    viewport.Model
	usage string
	focus bool
}

func NewUsageViewportModel() UsageViewportModel {
	vp := viewport.New(0, 0)
	usageFiles := embed.EmbeddedFilesInstance()
	g, _ := glamour.NewTermRenderer(glamour.WithStylesFromJSONBytes(usageFiles.UsageTheme), glamour.WithEmoji())
	usage, _ := g.Render(string(usageFiles.UsageFile))
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
		m.vp.SetContent(m.renderViewport())
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m UsageViewportModel) View() string {
	return m.vp.View()
}

func (m UsageViewportModel) renderViewport() string {
	title := sectionTitleStyle.Render("Letschat Usage")
	title = lipgloss.PlaceHorizontal(usageWidth(), lipgloss.Center, title)
	return lipgloss.JoinVertical(lipgloss.Left, title, m.usage)
}
