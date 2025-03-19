package middleware

import (
	"net/http"

	"github.com/michaelhass/cpaw/mux"
	"github.com/michaelhass/cpaw/service"
)

func AuthProtected(authService *service.AuthService) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("cpaw_session")
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

			_, err = authService.VerifyToken(r.Context(), c.Value)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
