package tui

import (
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/exp/maps"
	"strings"
)

type LoginModel struct {
	txtInputs    []textinput.Model
	placeholders []string
	activeBtn    int  // -1 -> none, 0 -> Continue 1 -> Signup
	tabIdx       int  // 0 - 1 -> txtInputs | 2 - 3 -> Continue & Signup btns
	dangerState  bool // we turn the form to secondaryColor
	ev           *domain.ErrValidation
}

// validateLoginModel validates the input form then adds the errors to ev
func (m *LoginModel) validateLoginModel() {
	// clear the previous errors
	maps.Clear(m.ev.Errors)
	domain.ValidateEmail(m.txtInputs[0].Value(), m.ev)
	m.ev.Evaluate(m.txtInputs[1].Value() != "", "password", "must be provided")
	populateErr := func(idx int, err string) {
		m.txtInputs[idx].Reset()
		m.txtInputs[idx].Placeholder = err
		m.txtInputs[idx].PlaceholderStyle = lipgloss.NewStyle().Foreground(secondaryColor)
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
	}
}

func InitialLoginModel() LoginModel {
	m := LoginModel{
		txtInputs: make([]textinput.Model, 2),
		placeholders: []string{
			"your email goes here...",
			"and here goes the password...",
		},
		activeBtn: -1,
		ev:        domain.NewErrValidation(),
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
		// must be after handling the active tab indices method
		m.dangerState = false // once there is a keypress remove the danger state
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			s := msg.String()
			if s == "enter" {
				if m.tabIdx == 2 {
					m.validateLoginModel()
					// TODO: Handle the login backend call
					// if the user is not activated send a new token and take him to otp page
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
		case "left":
			if m.tabIdx == 3 {
				m.activeBtn = 0
				m.tabIdx--
			}
		case "right":
			if m.tabIdx == 2 {
				m.activeBtn = 1
				m.tabIdx++
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

func (m LoginModel) View() string {
	var sb strings.Builder
	sb.WriteString(letschatLogo)
	sb.WriteString(infoTxtStyle.Render("Login to your account"))
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
	if m.activeBtn == 0 {
		continueBtn = activeButtonStyleWithColor(whiteColor, primaryColor).Render("Continue")
	} else if m.activeBtn == 1 {
		signupBtn = activeButtonStyleWithColor(whiteColor, primaryColor).Render("Register")
	}
	if m.tabIdx >= len(m.txtInputs) {
		sb.WriteString(activeBtnInputStyle.Render(continueBtn, signupBtn))
	} else {
		sb.WriteString(btnInputStyle.Render(continueBtn, signupBtn))
	}
	c := container
	if m.dangerState {
		c = c.BorderForeground(secondaryColor)
	}
	return containerCentered(c.Render(sb.String()))
}
