package client

import (
	"errors"
)

var (
	ErrServerValidation = errors.New("server validation error")
	ErrExpiredOTP       = errors.New("expired otp")
	ErrNonActiveUser    = errors.New("not activated")
	ErrUnauthorized     = errors.New("invalid credentials")
	// ErrApplication code is 0
	ErrApplication = errors.New("Your side of application have encountered an error, if the error persists you may report this issue to the developer at https://github.com/M0hammadUsman/letschat.")
)

func getMostNestedError(err error) error {
	for err != nil {
		next := errors.Unwrap(err)
		if next == nil {
			return err
		}
		err = next
	}
	return nil
}
