package tui

import (
	"errors"
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"strings"
	"time"
)

const timeout = 5 * time.Second

var (
	otpInputStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), false, false, true, false).
		BorderForeground(greyColor).
		Padding(0, 1, 0, 1).
		Margin(1, 0, 1, 0).
		Width(10).
		Align(lipgloss.Center)
)

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
	i.PlaceholderStyle = lipgloss.NewStyle().Foreground(greyColor)
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
		m.errMsg = ""
		m.otp.PlaceholderStyle = lipgloss.NewStyle().Foreground(greyColor)
		switch msg.String() {
		case "ctrl+c", "q":
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
		if msg == "Expired!" {
			m.populateErr(msg.String())
		} else {
			m.errMsg = msg
		}
		m.dangerState = true
		m.sent = false
		return m, nil

	case doneMsg:
		return InitialLoginModel(), nil
	}

	var cmd tea.Cmd
	m.otp, cmd = m.otp.Update(msg)
	return m, cmd
}

func (m OtpModel) View() string {
	var sb strings.Builder
	sb.WriteString(letschatLogo)
	c := container
	otpStyle := otpInputStyle

	if m.dangerState && m.errMsg != "" {
		c = c.BorderForeground(secondaryColor)
		otpStyle = otpInputStyle.BorderForeground(secondaryColor)
		m.otp.TextStyle = m.otp.TextStyle.Foreground(secondaryColor)
		e := wrap.String(wordwrap.String(m.errMsg.String(), 60), 60)
		sb.WriteString(infoTxtStyle.Foreground(secondaryColor).Render(e))
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
			btnStyle = buttonStyle.Background(primaryColor).Foreground(whiteColor)
		} else {
			btnStyle = btnStyle.Background(secondaryColor).Foreground(whiteColor)
		}
	}
	var timeStr string
	if !m.timer.Timedout() {
		timeStr = " in " + m.timer.View()
		btnStyle = btnStyle.Width(15)
	}
	sb.WriteString(btnInputStyle.Render(btnStyle.Render(fmt.Sprintf("Resend%v", timeStr))))
	otpContainedView := c.Render(sb.String())
	return containerCentered(otpContainedView)
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
	m.otp.PlaceholderStyle = lipgloss.NewStyle().Foreground(secondaryColor)
	m.dangerState = true
	m.ev = domain.NewErrValidation() // so we don't have errors when calling ev.HasErrors() next time
}

func (m OtpModel) activateUser() tea.Cmd {
	return func() tea.Msg {
		if err := m.client.ActivateUser(m.otp.Value()); err != nil {
			if errors.Is(err, client.ErrExpiredOTP) {
				return errMsg("Expired!")
			} else {
				return errMsg(err.Error())
			}
		} else {
			return doneMsg{}
		}
	}
}

func (m OtpModel) resendOtp() tea.Cmd {
	return func() tea.Msg {
		if err := m.client.ResendOtp(m.userEmail); err != nil {
			return errMsg(err.Error())
		}
		return nil
	}
}
