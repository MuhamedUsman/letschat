package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"github.com/MuhamedUsman/letschat/internal/domain"
	"math/big"
	"time"
)

var _ domain.TokenService = (*TokenService)(nil)

type TokenService struct {
	tokenRepo domain.TokenRepository
}

func NewTokenService(tokenRepo domain.TokenRepository) *TokenService {
	return &TokenService{tokenRepo: tokenRepo}
}

// GenerateToken generates OTP if scope is ScopeActivation & AuthenticationToken if scope is ScopeAuthentication
func (s *TokenService) GenerateToken(ctx context.Context, userID string, scope string) (string, error) {
	token := new(domain.Token)
	var err error
	switch scope {
	case domain.ScopeActivation:
		token, err = generateOTP(userID, scope, domain.ScopeActivationTTL)
	case domain.ScopeAuthentication:
		token, err = generateAuthToken(userID, scope, domain.ScopeAuthenticationTTL)
	default:
		panic("invalid token scope")
	}
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}
	if err = s.tokenRepo.Insert(ctx, token); err != nil {
		return "", fmt.Errorf("error inserting token: %w", err)
	}
	return token.PlainText, nil
}

func (s *TokenService) DeleteAllForUser(ctx context.Context, userID string, scope string) error {
	return s.tokenRepo.DeleteAllForUser(ctx, userID, scope)
}

func generateOTP(userID, scope string, ttl time.Duration) (*domain.Token, error) {
	token := &domain.Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}
	otp, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return nil, err
	}
	token.PlainText = fmt.Sprintf("%06d", otp)
	hashArray := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hashArray[:]
	return token, err
}

func generateAuthToken(userID, scope string, ttl time.Duration) (*domain.Token, error) {
	token := &domain.Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, err
	}
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randBytes)
	hashArray := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hashArray[:]
	return token, nil
}
