package handler

import (
	"net/http"

	cmux "github.com/michaelhass/cpaw/mux"
	"github.com/michaelhass/cpaw/service"
	"github.com/michaelhass/cpaw/views"
)

type TemplateHandler struct {
	authService *service.AuthService
	itemService *service.ItemService
}

func NewTemplateHandler(
	authService *service.AuthService,
	itemService *service.ItemService,
) *ApiHandler {
	return &ApiHandler{
		authService: authService,
		itemService: itemService,
	}
}

func (th *TemplateHandler) RegisterRoutes(mux *cmux.Mux) {
	mux.HandleFunc("GET /login/", func(w http.ResponseWriter, r *http.Request) {
		component := views.WithDefaultPage(views.Login())
		component.Render(r.Context(), w)
	})

	mux.HandleFunc("GET /signup/", func(w http.ResponseWriter, r *http.Request) {
		component := views.WithDefaultPage(views.SignUp())
		component.Render(r.Context(), w)
	})
}
