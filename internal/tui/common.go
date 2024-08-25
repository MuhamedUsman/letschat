package tui

import "errors"

type errMsg string

func (e errMsg) String() string {
	return string(e)
}

type doneMsg struct{}

var (
	ErrValidation = errors.New("validation error")
)
