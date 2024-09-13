package facade

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/api/mailer"
	"github.com/M0hammadUsman/letschat/internal/api/service"
	"github.com/M0hammadUsman/letschat/internal/common"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"log/slog"
)

type UserFacade struct {
	service   *service.Service
	txManager TXManager
	mailer    *mailer.Mailer
	bgTask    *common.BackgroundTask
}

func NewUserFacade(service *service.Service,
	txMan TXManager,
	mailer *mailer.Mailer,
	bgTask *common.BackgroundTask) *UserFacade {
	return &UserFacade{
		service:   service,
		txManager: txMan,
		mailer:    mailer,
		bgTask:    bgTask,
	}
}

func (f *UserFacade) RegisterUser(ctx context.Context, u *domain.UserRegister) error {
	var otp string
	// Registering User & Generating OTP in a transaction
	if err := f.txManager.RunInTX(ctx, func(ctx context.Context) error {
		userID, err := f.service.UserService.RegisterUser(ctx, u)
		if err != nil {
			return err
		}
		otp, err = f.service.GenerateToken(ctx, userID, domain.ScopeActivation)
		return err
	}); err != nil {
		return err
	}
	// Sending Activation Email
	f.bgTask.Run(func(context.Context) {
		data := map[string]any{
			"name":  u.Name,
			"token": otp,
		}
		if err := f.mailer.Send(u.Email, "email.tmpl.html", data); err != nil {
			slog.Error(err.Error())
		}
	})
	return nil
}

func (f *UserFacade) GetByUniqueField(ctx context.Context, fieldValue string) (*domain.User, error) {
	return f.service.GetByUniqueField(ctx, fieldValue)
}

func (f *UserFacade) UpdateUser(ctx context.Context, u *domain.UserUpdate) error {
	return f.service.UpdateUser(ctx, u)
}

func (f *UserFacade) UpdateUserOnlineStatus(ctx context.Context, u *domain.User, online bool) error {
	return f.service.UpdateUserOnlineStatus(ctx, u, online)
}

func (f *UserFacade) ActivateUser(ctx context.Context, plainToken string) error {
	return f.txManager.RunInTX(ctx, func(ctx context.Context) error {
		usr, err := f.service.GetForToken(ctx, domain.ScopeActivation, plainToken)
		if err != nil {
			return err
		}
		if err = f.service.ActivateUser(ctx, usr); err != nil {
			return err
		}
		return f.service.DeleteAllForUser(ctx, usr.ID, domain.ScopeActivation)
	})
}

func (f *UserFacade) SearchUser(
	ctx context.Context,
	queryParam string,
	filter domain.Filter,
) ([]*domain.User, *domain.Metadata, error) {
	return f.service.GetByQuery(ctx, queryParam, filter)
}
