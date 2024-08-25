package client

import "errors"

var (
	ErrServerValidation = errors.New("server validation error")
	ErrExpiredOTP       = errors.New("expired otp")
)
