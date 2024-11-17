package tui

import (
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"strings"
)

type UpdateProfileModel struct {
	inputTitles  []string
	txtInputs    []textinput.Model
	placeholders []string
	tabIdx       int
	spinner      spinner.Model
	spin         bool
	dangerState  bool
	includePass  bool
	ev           *domain.ErrValidation
}

func NewUpdateProfileModel() UpdateProfileModel {
	up := UpdateProfileModel{
		inputTitles: []string{"Name", "Email", "Previous Password", "New Password", "Confirm Password"},
		txtInputs:   make([]textinput.Model, 5),
		tabIdx:      -1,
		includePass: true,
		spinner:     spinner.New(),
	}
	for i := range up.txtInputs {
		crsr := cursor.New()
		crsr.Style = lipgloss.NewStyle().Foreground(primaryColor)

		t := textinput.New()
		t.Prompt = ""
		t.PlaceholderStyle = lipgloss.NewStyle().Foreground(primarySubtleDarkColor)
		t.TextStyle = lipgloss.NewStyle().Foreground(primaryColor)
		t.Cursor = crsr
		t.CharLimit = 64

		switch i {
		case 0:
			t.Placeholder = "Muhammad Usman"
		case 1:
			t.Placeholder = "usmannadeem3344@gmail.com"
		case 2, 3, 4:
			t.EchoCharacter = '*'
			t.EchoMode = textinput.EchoPassword
		}

		up.txtInputs[i] = t
	}
	return up
}

func (m UpdateProfileModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m UpdateProfileModel) Update(msg tea.Msg) (UpdateProfileModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.setTxtInputWidthAccordingly()

	case tea.KeyMsg:
		switch msg.String() {

		case "tab":
			if !m.includePass && m.tabIdx == 1 {
				m.tabIdx = 4
			}
			m.tabIdx = (m.tabIdx + 1) % (len(m.inputTitles) + 2)
			return m, m.focusTxtInputsAccordingly()

		case "esc":
			m.tabIdx = -1
			return m, m.focusTxtInputsAccordingly()

		case "enter":
			switch m.tabIdx {
			case 0, 1, 2, 3, 4:
				if !m.includePass && m.tabIdx == 1 {
					m.tabIdx = 5
				} else {
					m.tabIdx++
				}
				m.focusTxtInputsAccordingly()
			case 5:
				m.includePass = !m.includePass
				// clear the associated fields
				for i := 2; i <= 4; i++ {
					m.txtInputs[i].Reset()
				}
				if m.includePass {
					m.tabIdx = 2
					m.focusTxtInputsAccordingly()
				}
			case 6:

			}

		case "up", "left":
			if m.tabIdx == 6 {
				m.tabIdx = 5
			}

		case "down", "right":
			if m.tabIdx == 5 {
				m.tabIdx = 6
			}
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease {
			for i := range 7 {
				if zone.Get(fmt.Sprint("formItem", i)).InBounds(msg) {
					m.tabIdx = i
					m.focusTxtInputsAccordingly()
				}
			}

		}
	}

	return m, tea.Batch(m.handleTxtInputUpdate(msg))
}

func (m UpdateProfileModel) View() string {
	title := sectionTitleStyle.Render("Account Settings")
	title = lipgloss.PlaceHorizontal(updateProfileWidth(), lipgloss.Center, title)
	form := updateProfileFormStyle.Render(m.renderForm())
	c := lipgloss.NewStyle().Width(updateProfileWidth())
	return c.Render(title, form)
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func (m UpdateProfileModel) renderForm() string {
	var sb strings.Builder
	for i, t := range m.inputTitles {
		// do not include password fields
		if i == 2 && !m.includePass {
			break
		}
		title := updateProfileInputHeaderStyle.Render(t)
		field := updateProfileInputFieldStyle.Width(updateProfileWidth() - 8).Render(m.txtInputs[i].View())
		if i == m.tabIdx {
			title = updateProfileInputHeaderStyle.Italic(true).Foreground(primaryColor).Render(t)
			field = updateProfileInputFieldStyle.Width(updateProfileWidth() - 8).
				BorderForeground(primaryColor).
				Render(m.txtInputs[i].View())
		}
		sb.WriteString(title)
		sb.WriteString("\n")
		field = zone.Mark(fmt.Sprint("formItem", i), field)
		sb.WriteString(field)
		sb.WriteString("\n")
	}
	sb.WriteString(m.renderFormBtns())
	return sb.String()
}

func (m UpdateProfileModel) renderFormBtns() string {
	s1 := "INCLUDE PASSWORD"
	s2 := "EXCLUDE PASSWORD"
	s3 := "UPDATE ACCOUNT"
	btn1 := updateProfileFromBlurBtnStyle.Render(s1)
	btn2 := updateProfileFromBlurBtnStyle.Padding(0, 3).Render(s3)
	if m.includePass {
		btn1 = updateProfileFromBlurBtnStyle.Render(s2)
	}
	switch m.tabIdx {
	case len(m.txtInputs):
		btn1 = updateProfileFormActiveBtnStyle.Render(s1)
		if m.includePass {
			btn1 = updateProfileFormDangerBtnStyle.Render(s2)
		}
	case len(m.txtInputs) + 1:
		btn2 = updateProfileFormActiveBtnStyle.Padding(0, 3).Render(s3)
	}
	btn1 = zone.Mark("formItem5", btn1)
	btn2 = zone.Mark("formItem6", btn2)
	btns := lipgloss.JoinHorizontal(lipgloss.Bottom, btn1, "  ", btn2)
	if updateProfileWidth() < 50 {
		btns = lipgloss.JoinVertical(lipgloss.Center, btn1, btn2)
	}
	return lipgloss.PlaceHorizontal(updateProfileWidth()-6, lipgloss.Center, btns)
}

func (m *UpdateProfileModel) handleTxtInputUpdate(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.txtInputs))
	for i := range m.txtInputs {
		m.txtInputs[i], cmds[i] = m.txtInputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m *UpdateProfileModel) focusTxtInputsAccordingly() tea.Cmd {
	var cmd tea.Cmd
	for i := range m.txtInputs {
		m.txtInputs[i].Blur()
		if m.tabIdx == i {
			cmd = m.txtInputs[i].Focus()
		}
	}
	return cmd
}

func (m *UpdateProfileModel) setTxtInputWidthAccordingly() {
	for i := range m.txtInputs {
		m.txtInputs[i].Width = updateProfileWidth() - 11
	}
}
