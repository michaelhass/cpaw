package ctx

import (
	"context"
)

const keyUserCtx = "keyUserCtx"

func WithUserId(parent context.Context, userId string) context.Context {
	return context.WithValue(parent, keyUserCtx, userId)
}

func GetUserId(c context.Context) (string, bool) {
	user, ok := c.Value(keyUserCtx).(string)
	return user, ok
}
