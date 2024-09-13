package utility

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"net/http"
)

type ctxKey string

const UserCtxKey = ctxKey("USER")

func ContextSetUser(r *http.Request, user *domain.User) *http.Request {
	ctx := context.WithValue(r.Context(), UserCtxKey, user)
	return r.WithContext(ctx)
}

func ContextGetUser(ctx context.Context) *domain.User {
	user, ok := ctx.Value(UserCtxKey).(*domain.User)
	if !ok {
		panic("missing user in request context")
	}
	return user
}
