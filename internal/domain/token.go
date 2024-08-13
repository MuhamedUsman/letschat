package domain

import (
	"context"
	"regexp"
	"time"
)

const (
	ScopeActivation        = "activation"
	ScopeAuthentication    = "authentication"
	ScopeActivationTTL     = 15 * time.Minute
	ScopeAuthenticationTTL = 7 * 24 * time.Hour
)

var (
	RgxOtp = regexp.MustCompile("^[0-9]{6}$")
)

type Token struct {
	PlainText string    `json:"plainText"`
	Hash      []byte    `json:"-"`
	UserID    string    `json:"-" db:"user_id"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

type TokenService interface {
	GenerateToken(ctx context.Context, userID string, scope string) (string, error)
	DeleteAllForUser(ctx context.Context, userID string, scope string) error
}

type TokenRepository interface {
	Insert(ctx context.Context, token *Token) error
	DeleteAllForUser(ctx context.Context, userID, scope string) error
}

func ValidateOTP(otp string, ev *ErrValidation) {
	ev.Evaluate(RgxOtp.MatchString(otp), "otp", "invalid sequence")
}

func ValidateAuthenticationToken(token string, ev *ErrValidation) {
	ev.Evaluate(token != "", "token", "must be provided")
	ev.Evaluate(len(token) == 26, "token", "must be 26 bytes long")
}
