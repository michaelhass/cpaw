package handlers

import "net/http"

func AuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("cpaw_session")
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// TODO: check session
		err = c.Valid()
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// token := c.Value
		// find token
		next.ServeHTTP(w, r)
	})
}
