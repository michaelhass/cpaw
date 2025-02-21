package middleware

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
