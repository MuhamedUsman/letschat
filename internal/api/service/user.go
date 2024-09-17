package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

// ensures UserService implements letschat.UserService interface
var _ domain.UserService = (*UserService)(nil)

type UserService struct {
	userRepository domain.UserRepository
}

func NewUserService(userRepo domain.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepo,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, u *domain.UserRegister) (string, error) {
	ev := domain.NewErrValidation()
	domain.ValidateName(u.Name, ev)
	domain.ValidateEmail(u.Email, ev)
	domain.ValidPlainPassword(u.Password, ev)
	if ev.HasErrors() {
		return "", ev
	}
	exists, err := s.ExistsUser(ctx, u.Email)
	if err != nil {
		return "", err
	}
	if exists {
		ev.AddError("email", "already exists")
		return "", ev
	}
	passHash, err := generatePasswordHash(u.Password) // check if exists then hash, takes 200 ms approx.
	if err != nil {
		return "", fmt.Errorf("error generating password hash: %w", err)
	}
	usr := &domain.User{
		Name:     u.Name,
		Email:    u.Email,
		Password: passHash,
	}
	userID, err := s.userRepository.RegisterUser(ctx, usr)
	if errors.Is(err, domain.ErrDuplicateEmail) {
		ev.AddError("email", "already exists")
		return "", ev
	}
	return userID, nil
}

func (s *UserService) ExistsUser(ctx context.Context, email string) (bool, error) {
	return s.userRepository.ExistsUser(ctx, email)
}

func (s *UserService) GetByUniqueField(ctx context.Context, fieldValue string) (*domain.User, error) {
	var fieldName string
	if strings.Contains(fieldValue, "@") {
		fieldName = "email"
	} else {
		fieldName = "id"
		if uuid.Validate(fieldValue) != nil { // if err (invalid UUID) send back 404, do not send too much detail
			return nil, domain.ErrRecordNotFound
		}
	}
	user, err := s.userRepository.GetByUniqueField(ctx, fieldName, fieldValue)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, u *domain.UserUpdate) error {
	ev := domain.NewErrValidation()
	domain.ValidateName(u.Name, ev)
	domain.ValidateEmail(u.Email, ev)
	domain.ValidPlainPassword(u.NewPassword, ev)
	if ev.HasErrors() {
		return ev
	}
	usr, err := s.userRepository.GetByUniqueField(ctx, "email", u.Email)
	if err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			ev.AddError("email", "not exists")
			return ev
		}
		return err
	}
	if !comparePasswordHash(usr.Password, u.CurrentPassword) {
		ev.AddError("currentPassword", "does not match")
		return ev
	}
	newPassHash, err := generatePasswordHash(u.NewPassword)
	if err != nil {
		return err
	}
	usr.Name = u.Name
	usr.Email = u.Email
	usr.Password = newPassHash
	if err = s.userRepository.UpdateUser(ctx, usr); err != nil {
		return err
	}
	return nil
}

func (s *UserService) UpdateUserOnlineStatus(ctx context.Context, usr *domain.User, online bool) error {
	u, err := s.userRepository.GetByUniqueField(ctx, "id", usr.ID)
	if err != nil {
		return err
	}
	var lastOnline *time.Time
	if !online {
		now := time.Now()
		lastOnline = &now
	}
	u.LastOnline = lastOnline
	return s.userRepository.UpdateUser(ctx, u)
}

func (s *UserService) GetForToken(ctx context.Context, scope string, plainToken string) (*domain.User, error) {
	ev := domain.NewErrValidation()
	switch scope {
	case domain.ScopeActivation:
		domain.ValidateOTP(plainToken, ev)
	case domain.ScopeAuthentication:
		domain.ValidateAuthenticationToken(plainToken, ev)
	}
	if ev.HasErrors() {
		return nil, ev
	}
	tokenHash := sha256.Sum256([]byte(plainToken))
	usr, err := s.userRepository.GetForToken(ctx, scope, tokenHash[:]) // converting array to slice
	if err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			switch scope {
			case domain.ScopeActivation:
				ev.AddError("otp", "invalid")
			case domain.ScopeAuthentication:
				ev.AddError("token", "invalid")
			}
			return nil, ev
		}
		return nil, err
	}
	return usr, nil
}

func (s *UserService) ActivateUser(ctx context.Context, user *domain.User) error {
	if user.Activated {
		return domain.ErrAlreadyActive
	}
	return s.userRepository.ActivateUser(ctx, user)
}

func (s *UserService) AuthenticateUser(ctx context.Context, u *domain.UserAuth) (string, error) {
	ev := domain.NewErrValidation()
	domain.ValidateEmail(u.Email, ev)
	domain.ValidPlainPassword(u.Password, ev)
	if ev.HasErrors() {
		return "", ev
	}
	usr, err := s.userRepository.GetByUniqueField(ctx, "email", u.Email)
	if err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			ev.AddError("email", "not registered")
			return "", ev
		}
	}
	if !usr.Activated {
		ev.AddError("email", "not activated")
		return "", ev
	}
	if !comparePasswordHash(usr.Password, u.Password) {
		ev.AddError("password", "does not match")
		return "", ev
	}
	return usr.ID, nil
}

func (s *UserService) GetByQuery(
	ctx context.Context,
	queryParam string,
	filter domain.Filter,
) ([]*domain.User, *domain.Metadata, error) {
	var paramName string // name or email
	if strings.Contains(queryParam, "@") {
		paramName = "email"
	} else {
		paramName = "name"
	}
	return s.userRepository.GetByQuery(ctx, paramName, queryParam, filter)
}

func (s *UserService) SetOnlineUsersLastSeen(ctx context.Context, t time.Time) error {
	return s.userRepository.SetOnlineUsersLastSeen(ctx, t)
}

func generatePasswordHash(plainPassword string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), 12)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// comparePasswordHash returns true for password-hash match, false otherwise
func comparePasswordHash(hash []byte, plain string) bool {
	return bcrypt.CompareHashAndPassword(hash, []byte(plain)) == nil
}
