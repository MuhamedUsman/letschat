package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"log/slog"
)

type UsageViewportModel struct {
	vp    viewport.Model
	focus bool
}

func NewUsageViewportModel() UsageViewportModel {
	vp := viewport.New(50, 32)
	vp.MouseWheelEnabled = true
	return UsageViewportModel{
		vp: vp,
	}
}

func (m UsageViewportModel) Init() tea.Cmd {
	return nil
}

func (m UsageViewportModel) Update(msg tea.Msg) (UsageViewportModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
		if err != nil {
			slog.Error(err.Error())
		}
		md, err := g.Render(usage())
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

func usage() string {
	return "# 🔎 DISCOVER TAB\n### SEARCH BAR\n- FOCUS        ⇒  `CTRL+F` OR `LEFT CLICK`\n### RESULT TABLE\n- UP           ⇒  `↑` OR `k` OR `SCROLL UP`\n- DOWN         ⇒  `↓` OR `j` OR `SCROLL DOWN`\n- SELECT       ⇒  `ENTER`\n---\n# 💭 CONVERSATIONS TAB\n### CONVERSATIONS LIST\n- FILTER       ⇒  `CTRL+F` OR `LEFT CLICK`\n- UP           ⇒  `↑` OR `K` OR `SCROLL UP`\n- DOWN         ⇒  `↓` OR `J` OR `SCROLL DOWN`\n- SELECT       ⇒  `ENTER` OR `LEFT CLICK ON NAME`\n### CHATTING WINDOW\n- FOCUS TYPING ⇒  `CTRL+T` OR `HOVER`\n- SEND MSG     ⇒  `CTRL+S`\n- ⇏ DEL LINE   ⇒  `CTRL+K`\n- ⇍ DEL LINE   ⇒  `CTRL+U`\n- CHAT OPTIONS ⇒  `CTRL+O` OR `LEFT CLICK ⚙️`\n- MESSAGE INFO ⇒  `RIGHT CLICK ON MESSAGE[^1]`\n- UP           ⇒  `↑` OR `K` OR `SCROLL UP`\n- PAGE UP      ⇒  `B` OR `PGUP`\n- ½ PG UP      ⇒  `U` OR `CTRL+U`\n- DOWN         ⇒  `↓` OR `J` OR `SCROLL DOWN`\n- PAGE DOWN    ⇒  `F` OR `PGDN`\n- ½ PG DOWN    ⇒  `D` OR `CTRL+D`\n---\n# ⚙️ PREFERENCES TAB\n### ACCOUNT SETTINGS FORM\n- MOVE FR-WARD  ⇒ `TAB`\n- MOVE BK-WARD  ⇒ `SHIFT + TAB`\n- SELECT FIELD  ⇒ `LEFT CLICK`\n- MOVE IN BTNS  ⇒ `↑` `←` `→` `↓`\n---\n**NOTE:** _To press a button, hit_ `ENTER`\n\n[^1]: Message must be completely in the viewport.\n"
}
