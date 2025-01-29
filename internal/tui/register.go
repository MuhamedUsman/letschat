package tui

import (
	"errors"
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/exp/maps"
	"strings"
)

type UserRegisterModel struct {
	txtInputs    []textinput.Model
	spinner      spinner.Model
	spin         bool
	placeholders []string
	activeBtn    int // -1 -> none, 0 -> Continue 1 -> Login
	tabIdx       int // 0 - 2 -> txtInputs | 3 - 4 -> Continue & Login btns
	dangerState  bool
	errMsg       errMsg
	ev           *domain.ErrValidation
	client       *client.Client
}

func InitialUserRegisterModel() UserRegisterModel {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(primaryContrastColor)
	s.Spinner = spinner.Meter

	m := UserRegisterModel{
		txtInputs: make([]textinput.Model, 3),
		spinner:   s,
		ev:        domain.NewErrValidation(),
		placeholders: []string{
			"What should we call you, probably your name",
			"How should we contact you, probably your email",
			"How should we authenticate you, most probably your ex's name",
		},
		client: client.Get(),
	}

	for i := range m.txtInputs {

		crsr := cursor.New()
		crsr.SetMode(cursor.CursorHide)

		input := textinput.New()
		input.Prompt = ""
		input.Cursor = crsr
		input.TextStyle = activeInputStyle
		input.CharLimit = 64

		switch i {
		case 0:
			input.Placeholder = m.placeholders[i]
			input.Focus()
		case 1:
			input.Placeholder = m.placeholders[i]
		case 2:
			input.Placeholder = m.placeholders[i]
			input.EchoMode = textinput.EchoPassword
			input.EchoCharacter = '*'
		}
		m.txtInputs[i] = input
	}
	return m
}

func (m UserRegisterModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m UserRegisterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		terminalWidth = msg.Width
		terminalHeight = msg.Height

	case tea.KeyMsg:
		m.handleActiveTabIdxElement()
		// must be after handling the activeTab tab indices method
		m.dangerState = false // once there is a keypress remove the danger state
		m.errMsg.err = ""
		switch msg.String() {

		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			// user hit continue btn
			if m.tabIdx == 3 {
				if !m.spin {
					m.spin = true
					if err := m.validateUserRegisterModel(); err == nil {
						return m, tea.Batch(m.spinner.Tick, m.registerUser())
					}
				}
				return m, nil
			}
			if m.tabIdx == 4 {
				loginModel := InitialLoginModel()
				return loginModel, loginModel.Init()
			}
			m.tabIdx++
			if m.tabIdx == 3 {
				m.activeBtn = 0
			}

		case "tab":
			if m.tabIdx == 4 {
				m.tabIdx = 0
			} else {
				m.tabIdx++
			}
		case "shift+tab":
			if m.tabIdx == 0 {
				m.tabIdx = 4
			} else {
				m.tabIdx--
			}
		case "right":
			if m.tabIdx == 3 {
				m.activeBtn = 1
				m.tabIdx++
			}
		case "left":
			if m.tabIdx == 4 {
				m.activeBtn = 0
				m.tabIdx--
			}
		}

		{ // Updating btns
			if m.tabIdx == 3 {
				m.activeBtn = 0
			} else if m.tabIdx == 4 {
				m.activeBtn = 1
			} else {
				m.activeBtn = -1
			}
		}

	case errMsg:
		m.errMsg = msg
		m.dangerState = true
		m.spin = false

	case *domain.UserRegister: // server validation error resp populated in *domain.UserRegister
		if msg.Name != m.txtInputs[0].Value() {
			m.populateErr(0, msg.Name)
		}
		if msg.Email != m.txtInputs[1].Value() {
			m.populateErr(1, msg.Email)
		}
		if msg.Password != m.txtInputs[2].Value() {
			m.populateErr(2, msg.Password)
		}
		m.txtInputs[2].Reset()
		return m, nil

	case doneMsg:
		otpModel := InitialOTPModel(m.txtInputs[1].Value())
		return otpModel, otpModel.Init()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	}
	// as the user focuses the input fields we reset the placeholders to defaults
	for i := range m.txtInputs {
		if m.txtInputs[i].Focused() {
			m.txtInputs[i].Placeholder = m.placeholders[i]
			m.txtInputs[i].PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
		}
	}
	return m, m.handleTxtInputs(msg)
}

func (m UserRegisterModel) View() string {
	var sb strings.Builder
	sb.WriteString(letschatLogo)
	if m.errMsg.err != "" && m.dangerState {
		e := ansi.Wordwrap(m.errMsg.String(), 60, " ")
		sb.WriteString(infoTxtStyle.Foreground(dangerColor).Render(e))
	} else {
		sb.WriteString(infoTxtStyle.Render("Signup for a new account"))
	}

	// Rendering txt input fields
	for i := range m.txtInputs {
		if i == m.tabIdx {
			sb.WriteString(activeInputStyle.Render(m.txtInputs[i].View()))
		} else {
			sb.WriteString(inputStyle.Render(m.txtInputs[i].View()))
		}
	}

	// Rendering btns
	continueBtn := buttonStyle.Render("Continue")
	loginBtn := buttonStyle.Render("Login→")
	if m.tabIdx >= len(m.txtInputs) {
		if m.activeBtn == 0 {
			var continueBtnTxt string
			if m.spin {
				continueBtnTxt = m.spinner.View()
			} else {
				continueBtnTxt = "Continue"
			}
			continueBtn = activeButtonStyleWithColor(primaryContrastColor, primaryColor).Render(continueBtnTxt)
			sb.WriteString(activeBtnInputStyle.Render(continueBtn, loginBtn))
		} else if m.activeBtn == 1 {
			loginBtn = activeButtonStyleWithColor(primaryContrastColor, primaryColor).Render("Login↗")
			sb.WriteString(activeBtnInputStyle.Render(continueBtn, loginBtn))
		}
	} else {
		sb.WriteString(btnInputStyle.Render(continueBtn, loginBtn))
	}
	c := formContainer
	if m.dangerState {
		c = c.BorderForeground(dangerColor)
	}
	return formContainerCentered(c.Render(sb.String()))
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

// validateUserRegisterModel validates the input form then adds the errors to ev
func (m *UserRegisterModel) validateUserRegisterModel() error {
	// clear the previous errors
	maps.Clear(m.ev.Errors)
	domain.ValidateName(m.txtInputs[0].Value(), m.ev)
	domain.ValidateEmail(m.txtInputs[1].Value(), m.ev)
	domain.ValidPlainPassword(m.txtInputs[2].Value(), m.ev)

	if m.ev.HasErrors() {
		m.txtInputs[2].Reset() // Reset password field for any error
		m.dangerState = true
		if err, ok := m.ev.Errors["name"]; ok {
			m.populateErr(0, err)
		}
		if err, ok := m.ev.Errors["email"]; ok {
			m.populateErr(1, err)
		}
		if err, ok := m.ev.Errors["password"]; ok {
			m.populateErr(2, err)
		}
		return ErrValidation
	}
	return nil
}

func (m *UserRegisterModel) populateErr(idx int, err string) {
	m.txtInputs[idx].Reset()
	m.txtInputs[idx].Placeholder = err
	m.txtInputs[idx].PlaceholderStyle = lipgloss.NewStyle().Foreground(dangerColor)
	m.spin = false
}

func (m *UserRegisterModel) handleActiveTabIdxElement() {
	for i := range m.txtInputs {
		if i == m.tabIdx {
			m.txtInputs[i].Focus()
		} else {
			m.txtInputs[i].Blur()
		}
	}
	// Changes at tabIdx 3 - 4 only affects the view (btns) so the logic will reside in the View method
}

func (m UserRegisterModel) handleTxtInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.txtInputs))
	for i := range m.txtInputs {
		if i == m.tabIdx {
			m.txtInputs[i], cmds[i] = m.txtInputs[i].Update(msg)
		}
	}
	return tea.Batch(cmds...)
}

func (m *UserRegisterModel) registerUser() tea.Cmd {
	return func() tea.Msg {
		u := &domain.UserRegister{
			Name:     m.txtInputs[0].Value(),
			Email:    m.txtInputs[1].Value(),
			Password: m.txtInputs[2].Value(),
		}
		if err := m.client.Register(u); err != nil {
			if errors.Is(err, client.ErrServerValidation) {
				return u
			}
			return errMsg{err: err.Error()}
		}
		return doneMsg{}
	}
}
