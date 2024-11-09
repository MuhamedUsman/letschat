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

type LoginModel struct {
	txtInputs    []textinput.Model
	placeholders []string
	spinner      spinner.Model
	spin         bool
	activeBtn    int  // -1 -> none, 0 -> Continue 1 -> Signup
	tabIdx       int  // 0 - 1 -> txtInputs | 2 - 3 -> Continue & Signup btns
	dangerState  bool // we turn the form to dangerColor
	errMsg       errMsg
	ev           *domain.ErrValidation
	client       *client.Client
}

type InActiveUser struct{}

func InitialLoginModel() LoginModel {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(whiteColor)
	s.Spinner = spinner.Monkey

	m := LoginModel{
		txtInputs: make([]textinput.Model, 2),
		placeholders: []string{
			"your email goes here...",
			"and here goes the password...",
		},
		spinner:   s,
		activeBtn: -1,
		ev:        domain.NewErrValidation(),
		client:    client.Get(),
	}
	for i := range m.txtInputs {
		ti := textinput.New()
		ti.Prompt = ""
		ti.CharLimit = 64
		ti.TextStyle = lipgloss.NewStyle().Foreground(primaryColor)
		ti.Cursor = cursor.New()
		ti.Cursor.SetMode(cursor.CursorHide)
		switch i {
		case 0:
			ti.Placeholder = m.placeholders[i]
		case 1:
			ti.Placeholder = m.placeholders[i]
			ti.EchoCharacter = '*'
			ti.EchoMode = textinput.EchoPassword
		}
		m.txtInputs[i] = ti
	}
	return m
}

func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			s := msg.String()
			if s == "enter" {
				if m.tabIdx == 2 && !m.spin {
					if err := m.validateLoginModel(); err != nil {
						return m, nil
					}
					m.spin = true
					return m, tea.Batch(m.spinner.Tick, m.login())
				} else if m.tabIdx == 3 {
					registerModel := InitialUserRegisterModel()
					return registerModel, registerModel.Init()
				} else {
					if m.tabIdx != 2 {
						m.tabIdx++
					}
				}
			}
		case "tab":
			if m.tabIdx == 3 {
				m.tabIdx = 0
			} else {
				m.tabIdx++
			}
		case "shift+tab":
			if m.tabIdx == 0 {
				m.tabIdx = 3
			} else {
				m.tabIdx--
			}
		case "right":
			if m.tabIdx == 2 {
				m.activeBtn = 1
				m.tabIdx++
			}
		case "left":
			if m.tabIdx == 3 {
				m.activeBtn = 0
				m.tabIdx--
			}
		}

		{ // Updating btns
			if m.tabIdx == 2 {
				m.activeBtn = 0
			} else if m.tabIdx == 3 {
				m.activeBtn = 1
			} else {
				m.activeBtn = -1
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case InActiveUser:
		m.spin = false
		m.dangerState = true
		m.errMsg.err = "initiating account activation"
		otpModel := InitialOTPModel(m.txtInputs[0].Value())
		return otpModel, tea.Sequence(m.resendOtp(), otpModel.Init())

	case errMsg:
		m.spin = false
		m.dangerState = true
		m.errMsg = msg
		return m, nil

	case doneMsg:
		m.spin = false
		mainModel := InitialTabContainerModel()
		return mainModel, mainModel.Init()
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

func (m LoginModel) View() string {
	var sb strings.Builder
	sb.WriteString(letschatLogo)
	if m.errMsg.err != "" && m.dangerState {
		e := ansi.Wordwrap(m.errMsg.String(), 60, " ")
		sb.WriteString(infoTxtStyle.Foreground(dangerColor).Render(e))
	} else {
		sb.WriteString(infoTxtStyle.Render("Login to your account"))
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
	signupBtn := buttonStyle.Render("Register")
	if m.tabIdx >= len(m.txtInputs) {
		if m.activeBtn == 0 {
			var continueBtnTxt string
			if m.spin {
				continueBtnTxt = m.spinner.View()
			} else {
				continueBtnTxt = "Continue"
			}
			continueBtn = activeButtonStyleWithColor(primaryContrastColor, primaryColor).Render(continueBtnTxt)
			sb.WriteString(activeBtnInputStyle.Render(continueBtn, signupBtn))
		} else if m.activeBtn == 1 {
			signupBtn = activeButtonStyleWithColor(primaryContrastColor, primaryColor).Render("Register")
			sb.WriteString(activeBtnInputStyle.Render(continueBtn, signupBtn))
		}
	} else {
		sb.WriteString(btnInputStyle.Render(continueBtn, signupBtn))
	}
	c := formContainer
	if m.dangerState {
		c = c.BorderForeground(dangerColor)
	}
	return formContainerCentered(c.Render(sb.String()))
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

// validateLoginModel validates the input form then adds the errors to ev
func (m *LoginModel) validateLoginModel() error {
	// clear the previous errors
	maps.Clear(m.ev.Errors)
	domain.ValidateEmail(m.txtInputs[0].Value(), m.ev)
	m.ev.Evaluate(m.txtInputs[1].Value() != "", "password", "must be provided")
	populateErr := func(idx int, err string) {
		m.txtInputs[idx].Reset()
		m.txtInputs[idx].Placeholder = err
		m.txtInputs[idx].PlaceholderStyle = lipgloss.NewStyle().Foreground(dangerColor)
	}

	if m.ev.HasErrors() {
		m.txtInputs[1].Reset() // Reset password field for any error
		m.dangerState = true
		if err, ok := m.ev.Errors["email"]; ok {
			populateErr(0, err)
		}
		if err, ok := m.ev.Errors["password"]; ok {
			populateErr(1, err)
		}
		return errors.New("validation errors")
	}
	maps.Clear(m.ev.Errors)
	return nil
}

func (m *LoginModel) handleActiveTabIdxElement() {
	for i := range m.txtInputs {
		if i == m.tabIdx {
			m.txtInputs[i].Focus()
		} else {
			m.txtInputs[i].Blur()
		}
	}
	// Changes at tabIdx 3 - 4 only affects the view (btns) so the logic will reside in the View method
}

func (m LoginModel) handleTxtInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.txtInputs))
	for i := range m.txtInputs {
		if m.tabIdx == i {
			m.txtInputs[i], cmds[i] = m.txtInputs[i].Update(msg)
		}
	}
	return tea.Batch(cmds...)
}

func (m LoginModel) login() tea.Cmd {
	return func() tea.Msg {
		u := domain.UserAuth{
			Email:    m.txtInputs[0].Value(),
			Password: m.txtInputs[1].Value(),
		}
		if err := m.client.Login(u); err != nil {
			if errors.Is(err, client.ErrNonActiveUser) {
				return InActiveUser{}
			}
			return errMsg{err: err.Error()}
		} else {
			return doneMsg{}
		}
	}
}

func (m LoginModel) resendOtp() tea.Cmd {
	return func() tea.Msg {
		if err := m.client.ResendOtp(m.txtInputs[0].Value()); err != nil {
			return errMsg{err: err.Error()}
		}
		return nil
	}
}
