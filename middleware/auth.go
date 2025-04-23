package middleware

import (
	"net/http"

	"github.com/michaelhass/cpaw/ctx"
	"github.com/michaelhass/cpaw/models"
	"github.com/michaelhass/cpaw/mux"
	"github.com/michaelhass/cpaw/service"
)

func getValidSessionFromCookie(authService *service.AuthService, r *http.Request, cookieName string) (models.Session, error) {
	var session models.Session
	c, err := r.Cookie(cookieName)
	if err != nil {
		return session, err
	}

	err = c.Valid()
	if err != nil {
		return session, err
	}

	session, err = authService.VerifyToken(r.Context(), c.Value)
	return session, err
}

func AuthProtected(authService *service.AuthService, cookieName string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := getValidSessionFromCookie(authService, r, cookieName)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := ctx.WithUserId(r.Context(), session.UserId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthProtectedRedirect(
	authService *service.AuthService,
	cookieName string,
	redirectTo string,
) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := getValidSessionFromCookie(authService, r, cookieName)
			if err != nil {
				http.Redirect(w, r, redirectTo, http.StatusSeeOther)
				return
			}

			ctx := ctx.WithUserId(r.Context(), session.UserId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SetAuthenticatedUserCtx(authService *service.AuthService, cookieName string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := getValidSessionFromCookie(authService, r, cookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := authService.GetUserById(r.Context(), session.UserId)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := ctx.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
