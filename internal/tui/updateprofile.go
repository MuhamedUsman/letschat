package tui

import (
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"golang.org/x/exp/maps"
	"strings"
)

type inputStyles struct {
	header lipgloss.Style
	field  lipgloss.Style
}

type UpdateProfileModel struct {
	inputTitles      []string
	errFieldTitles   []string
	inputFieldStyles []inputStyles
	txtInputs        []textinput.Model
	placeholders     []string
	tabIdx           int
	spinner          spinner.Model
	spin             bool
	dangerState      bool
	includePass      bool
	ev               *domain.ErrValidation
	client           *client.Client
}

func NewUpdateProfileModel(c *client.Client) UpdateProfileModel {
	up := UpdateProfileModel{
		inputTitles:      []string{"Name", "Email", "Previous Password", "New Password", "Confirm Password"},
		errFieldTitles:   []string{"name", "email", "prevPass", "newPass", "confirmPass"},
		inputFieldStyles: make([]inputStyles, 5),
		txtInputs:        make([]textinput.Model, 5),
		tabIdx:           -1,
		includePass:      true,
		spinner:          spinner.New(),
		ev:               domain.NewErrValidation(),
		client:           c,
	}
	for i := range up.txtInputs {
		crsr := cursor.New()
		crsr.Style = lipgloss.NewStyle().Foreground(primaryColor)
		crsr.TextStyle = crsr.Style

		t := textinput.New()
		t.Prompt = ""
		t.PlaceholderStyle = lipgloss.NewStyle().Foreground(primarySubtleDarkColor)
		t.TextStyle = lipgloss.NewStyle().Foreground(primaryColor)
		t.Cursor = crsr
		t.CharLimit = 64

		switch i {
		case 2, 3, 4:
			t.EchoCharacter = '*'
			t.EchoMode = textinput.EchoPassword
		}

		up.txtInputs[i] = t
	}
	return up
}

func (m UpdateProfileModel) Init() tea.Cmd {
	return nil
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
			m.focusTxtInputsAccordingly()

		case "esc":
			m.tabIdx = -1
			m.removeAllErrors()
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
				if err := m.validateTxtInputs(); err == nil {
					return m, nil
				}
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
		if msg.Button == tea.MouseButtonLeft {
			for i := range 7 {
				if zone.Get(fmt.Sprint("formItem", i)).InBounds(msg) {
					m.tabIdx = i
					m.focusTxtInputsAccordingly()
				}
			}
		}
	}

	m.removeErrAccordingToTabIdx()
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

func (m *UpdateProfileModel) populatePlaceholders() {
	if m.client.CurrentUsr != nil {
		m.txtInputs[0].Placeholder = m.client.CurrentUsr.Name
		m.txtInputs[1].Placeholder = m.client.CurrentUsr.Email
	}
}

func (m UpdateProfileModel) renderForm() string {
	m.manageInputStylesAccordingly()
	var sb strings.Builder
	for i, t := range m.inputTitles {
		if i == 2 && !m.includePass {
			// do not include password fields
			break
		}
		title := m.inputFieldStyles[i].header.Render(t)
		field := m.inputFieldStyles[i].field.Render(m.txtInputs[i].View())
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

func (m *UpdateProfileModel) manageInputStylesAccordingly() {
	for i := range m.inputTitles {
		if i == 2 && !m.includePass {
			// do not include password fields
			break
		}
		m.inputFieldStyles[i].header = updateProfileInputHeaderStyle
		m.inputFieldStyles[i].field = updateProfileInputFieldStyle.Width(updateProfileWidth() - 8)
		if i == m.tabIdx {
			m.inputFieldStyles[i].header = updateProfileInputHeaderStyle.Italic(true).Foreground(primaryColor)
			m.inputFieldStyles[i].field = updateProfileInputFieldStyle.Width(updateProfileWidth() - 8).BorderForeground(primaryColor)
		}
	}
	if m.ev.HasErrors() {
		for j, et := range m.errFieldTitles { // et -> errorTitle
			if _, ok := m.ev.Errors[et]; ok {
				m.inputFieldStyles[j].header = updateProfileInputHeaderDangerStyle
				m.inputFieldStyles[j].field = updateProfileInputFieldDangerStyle
			}
		}
	}
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

// validateUserRegisterModel validates the input form then adds the errors to ev
func (m *UpdateProfileModel) validateTxtInputs() error {
	// clear the previous errors
	maps.Clear(m.ev.Errors)
	for i := range m.txtInputs {
		if !m.includePass && i == 2 {
			break
		}
		switch i {
		case 0:
			domain.ValidateName(m.txtInputs[i].Value(), m.ev)
		case 1:
			domain.ValidateEmail(m.txtInputs[i].Value(), m.ev)
		case 2, 3, 4:
			domain.ValidPlainPasswordWithKey(m.txtInputs[i].Value(), m.ev, m.errFieldTitles[i])
		}
	}
	// if passwords do not match
	if m.txtInputs[4].Value() != m.txtInputs[3].Value() {
		m.txtInputs[4].Reset()
		m.ev.AddError(m.errFieldTitles[4], "must match the new password")
	}
	if m.ev.HasErrors() {
		for i, et := range m.errFieldTitles { // et -> errorTitle
			if err, ok := m.ev.Errors[et]; ok {
				m.populateErr(i, err)
			}
		}
		return ErrValidation
	}
	return nil
}

func (m *UpdateProfileModel) populateErr(idx int, err string) {
	m.txtInputs[idx].Reset()
	m.txtInputs[idx].Placeholder = err
	m.txtInputs[idx].PlaceholderStyle = lipgloss.NewStyle().Foreground(dangerColor)
	m.spin = false
}

func (m *UpdateProfileModel) removeErrAccordingToTabIdx() {
	// Remove errors as field gets into focus
	if m.ev.HasErrors() {
		for i, et := range m.errFieldTitles {
			if m.txtInputs[i].Focused() {
				m.txtInputs[i].Placeholder = ""
				delete(m.ev.Errors, et)
				break
			}
		}
	}
}

func (m *UpdateProfileModel) removeAllErrors() {
	maps.Clear(m.ev.Errors)
	for i := range m.txtInputs {
		m.txtInputs[i].Placeholder = ""
	}
}
