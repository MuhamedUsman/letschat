package tui

import (
	"errors"
	"fmt"
	"github.com/MuhamedUsman/letschat/internal/client"
	"github.com/MuhamedUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/exp/maps"
	"strings"
	"time"
)

const timeout = 15 * time.Second

type OtpModel struct {
	otp         textinput.Model
	timer       timer.Model
	placeholder string
	sent        bool
	tabIdx      int
	dangerState bool
	userEmail   string
	errMsg      errMsg
	ev          *domain.ErrValidation
	client      *client.Client
}

func InitialOTPModel(email string) OtpModel {
	i := textinput.New()
	i.CharLimit = 6
	i.Prompt = ""
	i.Placeholder = "$$$$$$"
	i.PlaceholderStyle = lipgloss.NewStyle().Foreground(darkGreyColor)
	i.TextStyle = lipgloss.NewStyle().Foreground(primaryColor)
	i.Focus()
	i.Cursor = cursor.New()
	i.Cursor.SetMode(cursor.CursorHide)

	return OtpModel{
		otp:         i,
		timer:       timer.New(timeout),
		placeholder: "$$$$$$",
		userEmail:   email,
		ev:          domain.NewErrValidation(),
		client:      client.Get(),
	}
}

func (m OtpModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.timer.Init())
}

func (m OtpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		terminalWidth = msg.Width
		terminalHeight = msg.Height
	case tea.KeyMsg:
		m.dangerState = false // reset the dangerState once there is a key press
		m.otp.Placeholder = m.placeholder
		m.errMsg.err = ""
		m.otp.PlaceholderStyle = lipgloss.NewStyle().Foreground(darkGreyColor)
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.tabIdx == 0 {
				if err := m.validateOtp(); err != nil {
					m.populateErr("Invalid!")
					return m, nil
				}
				m.sent = true
				return m, m.activateUser()
			}
			if m.tabIdx == 1 {
				if m.timer.Timedout() {
					m.sent = false
					m.timer.Timeout = timeout
					return m, tea.Batch(m.timer.Init(), m.resendOtp())
				}
			}
		case "tab":
			if m.tabIdx == 1 {
				m.otp.Focus()
				m.tabIdx = 0
			} else {
				m.otp.Blur()
				m.tabIdx++
			}
		case "shift+tab":
			if m.tabIdx == 0 {
				m.otp.Blur()
				m.tabIdx = 1
			} else {
				m.otp.Focus()
				m.tabIdx--
			}
		}
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case errMsg:
		if msg.err == "Expired!" {
			m.populateErr(msg.String())
		} else {
			m.errMsg = msg
		}
		m.dangerState = true
		m.sent = false
		return m, nil

	case doneMsg:
		loginModel := InitialLoginModel()
		return loginModel, loginModel.Init()
	}

	var cmd tea.Cmd
	m.otp, cmd = m.otp.Update(msg)
	return m, cmd
}

func (m OtpModel) View() string {
	var sb strings.Builder
	sb.WriteString(letschatLogo)
	c := formContainer
	otpStyle := otpInputStyle

	if m.dangerState && m.errMsg.err != "" {
		c = c.BorderForeground(dangerColor)
		otpStyle = otpInputStyle.BorderForeground(dangerColor)
		m.otp.TextStyle = m.otp.TextStyle.Foreground(dangerColor)
		e := ansi.Wordwrap(m.errMsg.String(), 60, " ")
		sb.WriteString(infoTxtStyle.Foreground(dangerColor).Render(e))
	} else {
		sb.WriteString(infoTxtStyle.Render("We've sent you some random digits, paste them here & hit enter"))
	}

	if m.tabIdx == 0 {
		otpStyle = otpStyle.BorderForeground(primaryColor)
	}
	sb.WriteString(otpStyle.Render(m.otp.View()))
	btnStyle := buttonStyle
	if m.tabIdx == 1 {
		if m.timer.Timedout() {
			btnStyle = buttonStyle.Background(primaryColor).Foreground(primaryContrastColor)
		} else {
			btnStyle = btnStyle.Background(dangerColor).Foreground(whiteColor)
		}
	}
	var timeStr string
	if !m.timer.Timedout() {
		timeStr = " in " + m.timer.View()
		btnStyle = btnStyle.Width(15)
	}
	sb.WriteString(btnInputStyle.Align(lipgloss.Center).Render(btnStyle.Render(fmt.Sprintf("Resend%v", timeStr))))
	otpContainedView := c.Render(sb.String())
	return formContainerCentered(otpContainedView)
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func (m OtpModel) validateOtp() error {
	domain.ValidateOTP(m.otp.Value(), m.ev)
	if m.ev.HasErrors() {
		return ErrValidation
	}
	return nil
}

func (m *OtpModel) populateErr(err string) {
	m.otp.Reset()
	m.otp.Placeholder = err
	m.otp.PlaceholderStyle = lipgloss.NewStyle().Foreground(dangerColor)
	m.dangerState = true
	maps.Clear(m.ev.Errors)
}

func (m OtpModel) activateUser() tea.Cmd {
	return func() tea.Msg {
		if err := m.client.ActivateUser(m.otp.Value()); err != nil {
			if errors.Is(err, client.ErrExpiredOTP) {
				return errMsg{err: "Expired!"}
			} else {
				return errMsg{err: err.Error()}
			}
		} else {
			return doneMsg{}
		}
	}
}

func (m OtpModel) resendOtp() tea.Cmd {
	return func() tea.Msg {
		if err := m.client.ResendOtp(m.userEmail); err != nil {
			return errMsg{err: err.Error()}
		}
		return nil
	}
}
