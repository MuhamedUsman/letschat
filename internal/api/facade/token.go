package facade

import (
	"context"
	"errors"
	"github.com/MuhamedUsman/letschat/internal/api/mailer"
	"github.com/MuhamedUsman/letschat/internal/api/service"
	"github.com/MuhamedUsman/letschat/internal/common"
	"github.com/MuhamedUsman/letschat/internal/domain"
	"log/slog"
)

type TokenFacade struct {
	service   *service.Service
	txManager TXManager
	mailer    *mailer.Mailer
	bgTask    *common.BackgroundTask
}

func NewTokenFacade(service *service.Service,
	txMan TXManager,
	mailer *mailer.Mailer,
	bgTask *common.BackgroundTask) *TokenFacade {
	return &TokenFacade{
		service:   service,
		txManager: txMan,
		mailer:    mailer,
		bgTask:    bgTask,
	}
}

func (t *TokenFacade) GenerateOTP(ctx context.Context, email string) error {
	usr, err := t.service.GetByUniqueField(ctx, email)
	ev := domain.NewErrValidation()
	if err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			ev.AddError("email", "not registered")
			return ev
		}
		return err
	}
	if usr.Activated {
		return domain.ErrAlreadyActive
	}
	otp, err := t.service.GenerateToken(ctx, usr.ID, domain.ScopeActivation)
	if err != nil {
		return err
	}
	t.bgTask.Run(func(context.Context) {
		data := map[string]string{
			"name":  usr.Name,
			"token": otp,
		}
		if err = t.mailer.Send(email, "email.tmpl.html", data); err != nil {
			slog.Error(err.Error())
		}
	})
	return nil
}

func (t *TokenFacade) GenerateAuthToken(ctx context.Context, u *domain.UserAuth) (string, error) {
	usrID, err := t.service.AuthenticateUser(ctx, u)
	if err != nil {
		return "", err
	}
	var otp string
	if err = t.txManager.RunInTX(ctx, func(ctx context.Context) error {
		// idempotent if domain.ScopeAuthentication tokens do not exist for the user
		if err = t.service.DeleteAllForUser(ctx, usrID, domain.ScopeAuthentication); err != nil {
			return err
		}
		otp, err = t.service.GenerateToken(ctx, usrID, domain.ScopeAuthentication)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return "", err
	}
	return otp, nil
}

func (t *TokenFacade) VerifyAuthToken(ctx context.Context, token string) (*domain.User, error) {
	usr, err := t.service.GetForToken(ctx, domain.ScopeAuthentication, token)
	if err != nil {
		return nil, err
	}
	return usr, nil
}
