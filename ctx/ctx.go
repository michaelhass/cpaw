package ctx

import (
	"context"

	"github.com/michaelhass/cpaw/models"
)

const keyUserIdCtx = "keyUserIdCtx"

func WithUserId(parent context.Context, userId string) context.Context {
	return context.WithValue(parent, keyUserIdCtx, userId)
}

func GetUserId(c context.Context) (string, bool) {
	user, ok := c.Value(keyUserIdCtx).(string)
	return user, ok
}

const keyUserCtx = "keyUserCtx"

func WithUser(parent context.Context, user models.User) context.Context {
	return context.WithValue(parent, keyUserCtx, user)
}

func GetUser(c context.Context) (models.User, bool) {
	user, ok := c.Value(keyUserCtx).(models.User)
	return user, ok
}
