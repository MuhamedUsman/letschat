package tui

import (
	"errors"
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
	"net/http"
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
	tabIdx           int
	spinner          spinner.Model
	ev               *domain.ErrValidation
	client           *client.Client
	// booleans for state management
	spin, includePass, showSuccess, populatePlaceholders, focus bool
	// to detect changes to currentUser name & email
	prevName, prevEmail string
}

func NewUpdateProfileModel(c *client.Client) UpdateProfileModel {
	up := UpdateProfileModel{
		inputTitles:          []string{"Name", "Email", "Previous Password", "New Password", "Confirm Password"},
		errFieldTitles:       []string{"name", "email", "prevPass", "newPass", "confirmPass"},
		inputFieldStyles:     make([]inputStyles, 5),
		txtInputs:            make([]textinput.Model, 5),
		tabIdx:               -1,
		populatePlaceholders: true,
		spinner:              newSpinner(),
		ev:                   domain.NewErrValidation(),
		client:               c,
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

	if m.populatePlaceholders {
		m.populateDefaultPlaceholders()
		m.populatePlaceholders = false
	}

	if !m.focus {
		m.resetForm()
		m.focusTxtInputsAccordingly()
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.setTxtInputWidthAccordingly()

	case tea.KeyMsg:
		switch msg.String() {

		case "tab":
			if m.focus {
				// if pass is not included, then after email field, goto first button
				if !m.includePass && m.tabIdx == 1 {
					m.tabIdx = 4
				}
				m.tabIdx = (m.tabIdx + 1) % (len(m.inputTitles) + 2)
				m.focusTxtInputsAccordingly()
			}

		case "shift+tab":
			if m.focus {
				l := len(m.inputTitles) + 2
				m.tabIdx = (m.tabIdx - 1 + l) % l
				if !m.includePass && m.tabIdx == 4 {
					m.tabIdx = 1
				}
				m.focusTxtInputsAccordingly()
			}

		case "esc":
			m.resetForm()
			m.focusTxtInputsAccordingly()

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
				if !m.spin {
					m.spin = true
					if err := m.validateTxtInputs(); err == nil {
						return m, tea.Batch(m.spinner.Tick, m.updateUser())
					}
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

	case spinner.TickMsg:
		if msg.ID == m.spinner.ID() {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case doneMsg:
		m.spin = false
		m.spinner = newSpinner() // reset spinner
		m.resetAllfields()
		m.includePass = false
		m.showSuccess = true
		m.tabIdx = -1
		m.populateDefaultPlaceholders()
		return m, countdownShowSuccessCmd()

	case hideSuccessMsg:
		m.showSuccess = false

	case *domain.ErrValidation:
		m.spin = false
		m.spinner = newSpinner()
		if msg.HasErrors() {
			m.populateServerErr(msg)
		}

	case errMsg:
		m.spin = false
		m.spinner = newSpinner()
		return m, func() tea.Msg { return &msg } // TabContainerModel will show this

	}

	m.removeErrAccordingToTabIdx()
	return m, tea.Batch(m.handleTxtInputUpdate(msg))
}

func (m UpdateProfileModel) View() string {
	title := sectionTitleStyle.Render("Account Settings")
	title = lipgloss.PlaceHorizontal(updateProfileWidth(), lipgloss.Center, title)
	form := updateProfileFormStyle.Render(m.renderForm())
	c := lipgloss.NewStyle().Width(updateProfileWidth()).Height(conversationHeight())
	return c.Render(title, form)
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func newSpinner() spinner.Model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)
	s.Spinner = spinner.Points
	return s
}

func (m *UpdateProfileModel) populateDefaultPlaceholders() {
	for m.client.CurrentUsr != nil { // the loop max runs for 2 iterations, tested it
		m.txtInputs[0].Placeholder = m.client.CurrentUsr.Name
		m.prevName = m.client.CurrentUsr.Name
		m.txtInputs[1].Placeholder = m.client.CurrentUsr.Email
		m.prevEmail = m.client.CurrentUsr.Email
		if m.client.CurrentUsr != nil {
			break
		}
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
	if m.showSuccess {
		successMsg := updateProfileFormSuccessStyle.Width(updateProfileWidth() - 10).Render()
		sb.WriteString(lipgloss.PlaceHorizontal(updateProfileWidth()-6, lipgloss.Center, successMsg))
	}
	return sb.String()
}

func (m UpdateProfileModel) renderFormBtns() string {
	s1 := "INCLUDE PASSWORD"
	s2 := "EXCLUDE PASSWORD"
	s3 := "UPDATE ACCOUNT"
	btn2Style := updateProfileFromBlurBtnStyle.Padding(0, 3)
	btn1 := updateProfileFromBlurBtnStyle.Render(s1)
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
		btn2Style = updateProfileFormActiveBtnStyle.Padding(0, 3)
	}
	if m.spin {
		s3 = m.spinner.View()
		btn2Style = updateProfileFromBlurBtnStyle.Padding(0, 8).Background(primaryContrastColor)
	}
	btn1 = zone.Mark("formItem5", btn1)
	btn2 := btn2Style.Render(s3)
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
				m.inputFieldStyles[j].field = updateProfileInputFieldDangerStyle.Width(updateProfileWidth() - 8)
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
		if m.tabIdx == i && m.focus {
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
	validateEmptyField := true
	for i := range m.txtInputs {
		if !m.includePass && i == 2 {
			break
		}
		// if password fields are not included, and only one of name/email is set, then if other is empty,
		// validate its placeholder instead
		if !m.includePass && (m.txtInputs[0].Value() != "" || m.txtInputs[1].Value() != "") {
			validateEmptyField = false
		}
		switch i {
		case 0, 1:
			// if only password fields are included, and name/email are empty -> validate their placeholders instead
			// if password fields are not included, and only one of name/email is set, then if other is empty,
			// validate its placeholder instead
			toValidate := m.txtInputs[i].Value()
			if (m.txtInputs[i].Value() == "" && m.includePass) || (!m.includePass && !validateEmptyField) {
				toValidate = m.txtInputs[i].Placeholder
			}
			if i == 0 {
				domain.ValidateName(toValidate, m.ev)
			} else {
				domain.ValidateEmail(toValidate, m.ev)
			}
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
				if i == 0 {
					m.txtInputs[i].Placeholder = m.client.CurrentUsr.Name
				} else if i == 1 {
					m.txtInputs[i].Placeholder = m.client.CurrentUsr.Email
				} else {
					m.txtInputs[i].Placeholder = ""
				}
				m.txtInputs[i].PlaceholderStyle = lipgloss.NewStyle().Foreground(primarySubtleDarkColor)
				delete(m.ev.Errors, et)
				break
			}
		}
	}
}

func (m *UpdateProfileModel) resetForm() {
	m.tabIdx = -1
	m.includePass = false
	m.removeAllErrors()
	m.populateDefaultPlaceholders()
}

func (m *UpdateProfileModel) removeAllErrors() {
	maps.Clear(m.ev.Errors)
	for i := range m.txtInputs {
		m.txtInputs[i].Placeholder = ""
		m.txtInputs[i].PlaceholderStyle = lipgloss.NewStyle().Foreground(primarySubtleDarkColor)
	}
}

func (m *UpdateProfileModel) resetAllfields() {
	for i := range m.txtInputs {
		m.txtInputs[i].Reset()
	}
}

func (m *UpdateProfileModel) populateServerErr(msg *domain.ErrValidation) {
	if err, ok := msg.Errors["name"]; ok {
		m.ev.AddError(m.errFieldTitles[0], err)
		m.populateErr(0, err)
	}
	if err, ok := msg.Errors["email"]; ok {
		m.ev.AddError(m.errFieldTitles[1], err)
		m.populateErr(1, err)
	}
	if err, ok := msg.Errors["currentPassword"]; ok {
		m.ev.AddError(m.errFieldTitles[2], err)
		m.populateErr(2, err)
	}
}

func (m *UpdateProfileModel) updateUser() tea.Cmd {
	return func() tea.Msg {
		u := domain.UserUpdate{
			ID:    m.client.CurrentUsr.ID,
			Name:  m.txtInputs[0].Value(),
			Email: m.txtInputs[1].Value(),
		}
		if u.Name == "" {
			u.Name = m.client.CurrentUsr.Name
		}
		if u.Email == "" {
			u.Email = m.client.CurrentUsr.Email
		}
		curPass := m.txtInputs[2].Value()
		newPass := m.txtInputs[4].Value()
		if m.includePass {
			u.CurrentPassword = &curPass
			u.NewPassword = &newPass
		}
		ev, code, err := m.client.UpdateUser(u)
		if code == http.StatusUnauthorized {
			return requireAuthMsg{}
		}
		if code == http.StatusInternalServerError {
			return errMsg{
				err:  "the server is overwhelmed",
				code: code,
			}
		}
		if err != nil {
			if errors.Is(err, client.ErrServerValidation) {
				return ev
			}
			return errMsg{
				err:  err.Error(),
				code: code,
			}
		}
		return doneMsg{}
	}
}
