package middleware

import (
	"net/http"
	"strings"
)

func AddTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL
		if !strings.HasSuffix(url.Path, "/") {
			url.Path += "/"
			r.URL = url
		}
		next.ServeHTTP(w, r)
	})
}
