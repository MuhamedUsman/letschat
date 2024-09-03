package domain

import "errors"

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrAlreadyActive  = errors.New("user already active")
	ErrInactive       = errors.New("user inactive")
)

type ErrValidation struct {
	Errors map[string]string
}

func NewErrValidation() *ErrValidation {
	return &ErrValidation{Errors: make(map[string]string)}
}

// implements error interface, so unwrap the error to get the validation errors
func (ErrValidation) Error() string {
	return "validation error"
}

func (e *ErrValidation) AddError(field, message string) {
	e.Errors[field] = message
}

func (e *ErrValidation) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *ErrValidation) Evaluate(ok bool, field, message string) {
	if !ok {
		e.AddError(field, message)
	}
}
