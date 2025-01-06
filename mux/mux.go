package mux

import (
	"net/http"
)

type Mux struct {
	http.ServeMux
	middlewares []MiddlewareFunc
}

func NewDefaultMux() *Mux {
	return &Mux{
		ServeMux:    *http.NewServeMux(),
		middlewares: []MiddlewareFunc{},
	}
}

func (m *Mux) Group(prefix string, fn func(m *Mux)) *Mux {
	groupRouter := NewDefaultMux()
	fn(groupRouter)
	m.Handle(prefix+"/", http.StripPrefix(prefix, groupRouter))
	return groupRouter
}

func (m *Mux) Use(middlewares ...MiddlewareFunc) {
	m.middlewares = append(m.middlewares, middlewares...)
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	var handler http.Handler = &m.ServeMux

	for i := len(m.middlewares) - 1; i >= 0; i-- {
		next := m.middlewares[i]
		handler = next(handler)
	}

	handler.ServeHTTP(w, request)
}

type MiddlewareFunc func(next http.Handler) http.Handler
