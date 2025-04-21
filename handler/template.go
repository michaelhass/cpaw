package handler

import (
	"net/http"
	"time"

	"github.com/a-h/templ"
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
) *TemplateHandler {
	return &TemplateHandler{
		authService: authService,
		itemService: itemService,
	}
}

func (th *TemplateHandler) RegisterRoutes(mux *cmux.Mux) {
	mux.Handle("/", th.indexPage())
	// mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	mux.Handle("GET /signin/", th.signInPage())
	mux.HandleFunc("POST /signin/", th.signInLogin)
	mux.HandleFunc("POST /signout/", th.signout)
}

func (th *TemplateHandler) indexPage() http.Handler {
	indexPage := views.WithDefaultPage(views.Index())
	return templ.Handler(indexPage)
}

func (th *TemplateHandler) signInPage() http.Handler {
	loginPage := views.WithDefaultPage(views.Login())
	return templ.Handler(loginPage)
}

func (th *TemplateHandler) signInLogin(w http.ResponseWriter, r *http.Request) {
	var (
		user     = r.FormValue("user_name")
		password = r.FormValue("password")
	)

	authResult, err := th.authService.SignIn(r.Context(), user, password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}
	cookie := &http.Cookie{}
	cookie.Name = sessionCookieName
	cookie.Value = authResult.Session.Token
	cookie.Expires = time.Unix(authResult.Session.ExpiresAt, 0)
	cookie.Path = "/"
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", 301)
}

func (th *TemplateHandler) signout(w http.ResponseWriter, r *http.Request) {

}
