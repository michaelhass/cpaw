package handler

import (
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/michaelhass/cpaw/ctx"
	"github.com/michaelhass/cpaw/middleware"
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
	mux.Handle("GET /signin/", th.signInPage())
	mux.HandleFunc("POST /signin/", th.signIn("/items/"))
	mux.HandleFunc("POST /signout/", th.signout)

	// mux.Handle("GET /items/", th.itemsPage("/signin/"))
	mux.Group("/items", func(items *cmux.Mux) {
		items.Use(middleware.AuthProtected(th.authService, sessionCookieName))
		items.Handle("GET /", th.itemsPage("/signin/"))
		items.HandleFunc("POST /", th.handleCreateItem)
	})
}

func (th *TemplateHandler) indexPage() http.Handler {
	indexPage := views.WithDefaultPage(views.IndexPage())
	return templ.Handler(indexPage)
}

func (th *TemplateHandler) signInPage() http.Handler {
	loginPage := views.WithDefaultPage(views.SignInPage())
	return templ.Handler(loginPage)
}

func (th *TemplateHandler) signIn(onSuccesRedirect string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		http.Redirect(w, r, onSuccesRedirect, http.StatusAccepted)
	}
}

func (th *TemplateHandler) signout(w http.ResponseWriter, r *http.Request) {

}

func (th *TemplateHandler) itemsPage(redirectTo string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context := r.Context()
		userId, ok := ctx.GetUserId(context)
		if !ok || len(userId) == 0 {
			http.Redirect(w, r, redirectTo, http.StatusUnauthorized)
			return
		}

		items, _ := th.itemService.ListItemsForUser(context, userId)
		itemsPage := views.WithDefaultPage(views.ItemsPage(items))
		itemsPage.Render(r.Context(), w)
	})
}

func (th *TemplateHandler) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	userId, ok := ctx.GetUserId(context)
	if !ok || len(userId) == 0 {
		http.Redirect(w, r, "/signin/", http.StatusUnauthorized)
		return
	}

	content := r.FormValue("content")
	_, err := th.itemService.CreateItem(context, service.CreateItemsParams{
		Content: content,
		UserId:  userId,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	items, err := th.itemService.ListItemsForUser(context, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	views.ItemList(items).Render(context, w)
}
