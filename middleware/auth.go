package middleware

import (
	"net/http"

	"github.com/michaelhass/cpaw/ctx"
	"github.com/michaelhass/cpaw/mux"
	"github.com/michaelhass/cpaw/service"
)

func AuthProtected(authService *service.AuthService, sessionCookieName string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(sessionCookieName)
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				return
			} else if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = c.Valid()
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			session, err := authService.VerifyToken(r.Context(), c.Value)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := ctx.WithUserId(r.Context(), session.UserId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
