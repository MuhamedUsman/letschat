package client

import "errors"

var (
	ErrServerDown       = errors.New("no connection could be made because the target machine actively refused it")
	ErrServerValidation = errors.New("server validation error")
	ErrExpiredOTP       = errors.New("expired otp")
	ErrNonActiveUser    = errors.New("not activated")
	ErrUnauthorized     = errors.New("invalid credentials")
)
