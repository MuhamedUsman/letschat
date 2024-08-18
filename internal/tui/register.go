package tui

import "C"
import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	FormBlock = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Width(100).
			Height(50).
			PaddingLeft(15).
			Margin(10).
			AlignVertical(lipgloss.Center)

	Logo = LetschatLogo.Margin(0, 0, 2, 1).Render("Letschat")
)

type UserRegisterModel struct {
	textInputs     []textinput.Model
	selectedButton int
	focus          int
	width          int
	height         int
}

func InitialModel() UserRegisterModel {
	m := UserRegisterModel{
		textInputs: make([]textinput.Model, 3),
	}
	var t textinput.Model
	for i := range m.textInputs {
		t = textinput.New()
		switch i {
		case 0:
			t.Placeholder = "What should we call you, probably your name..."
			t.Width = 60
			t.CharLimit = 30
			t.Focus()
			t.TextStyle = lipgloss.NewStyle().Foreground(primaryColor)
		case 1:
			t.Placeholder = "How should we contact you, probably your email..."
			t.Width = 60
			t.CharLimit = 30
		case 2:
			t.Placeholder = "How should we authenticate you, most probably your ex's name..."
			t.Width = 60
			t.CharLimit = 30
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '*'
		}
		m.textInputs[i] = t
	}
	return m
}

func (m UserRegisterModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m UserRegisterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.focus < len(m.textInputs)+2 {
				m.focus++
			} else if m.focus == 3 {
				m.selectedButton = 1
			} else if m.focus == 4 {
				m.selectedButton = 2
			}
		case "tab":
			if m.focus < len(m.textInputs)+2 {
				m.focus++
			} else {
				m.focus = 0
			}
		}
	}
	for _, t := range m.textInputs {
		t.Focus()
	}
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m UserRegisterModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))
	for i := range m.textInputs {
		m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m UserRegisterModel) View() string {
	var b strings.Builder
	b.WriteString(Logo)
	for i := range m.textInputs {
		border := InputInactiveBorder
		if i == m.focus {
			border = InputActiveBorder
		}
		b.WriteString(border.Render(m.textInputs[i].View()))
	}

	continueStyle := InputInactiveButton
	quitStyle := InputInactiveButton
	if m.selectedButton == 1 {
		continueStyle = ContinueButton
	} else if m.selectedButton == 2 {
		quitStyle = QuitButton
	}

	buttons := InputActiveBorder.Render(
		continueStyle.Render("Continue"),
		quitStyle.Render("Quit"),
	)
	b.WriteString(buttons)

	return FormBlock.Height(m.height / 2).Width(m.width / 2).Render(b.String())
}
