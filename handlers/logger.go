package handlers

import (
	"log"
	"net/http"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusWriter := newStatusResponseWriter(w)
		next.ServeHTTP(statusWriter, r)
		log.Println(r.Method, r.URL.Path, statusWriter.statusCode)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusResponseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

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
