package tui

import (
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

var (
	ErrValidation = errors.New("validation error")
)

type errMsg struct {
	err  string
	code int
}

func (e errMsg) String() string {
	return e.err
}

type doneMsg struct{}

type requireAuthMsg struct{}

func requireAuthCmd() tea.Msg {
	return requireAuthMsg{}
}

type spinMsg struct{}

func spinnerSpinCmd() tea.Msg { return spinMsg{} }

type resetSpinnerMsg struct{}

func spinnerResetCmd() tea.Msg { return resetSpinnerMsg{} }

type selDiscUserMsg struct { // selected Discovered User Msg
	id, name, email string
}

type SentMsg *domain.Message

type echoTypingMsg struct{}

func echoTypingCmd() tea.Cmd {
	return func() tea.Msg {
		t := time.NewTimer(2 * time.Second)
		<-t.C
		return echoTypingMsg{}
	}
}

type deleteForMeSuccessMsg string // stores id of the deleted msg, remove msg with this id from the msgs slice

type deleteForEveryoneSuccessMsg struct{}
