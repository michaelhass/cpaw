package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/michaelhass/cpaw/ctx"
	"github.com/michaelhass/cpaw/middleware"
	"github.com/michaelhass/cpaw/models"
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
	indexRedirectPath := "/"
	authProtectedRedirect := middleware.AuthProtectedRedirect(th.authService, sessionCookieName, indexRedirectPath)
	mux.Use(middleware.SetAuthenticatedUserCtx(th.authService, sessionCookieName))
	mux.HandleFunc("/", th.handleIndexPage)

	mux.HandleFunc("POST /signin/", th.handleSignIn(indexRedirectPath))
	mux.HandleFunc("POST /signout/", th.handleSignOut(indexRedirectPath))

	mux.Group("/items", func(items *cmux.Mux) {
		items.Use(authProtectedRedirect)
		items.HandleFunc("GET /", th.handleGetItems)
		items.HandleFunc("POST /", th.handleCreateItem)
		items.HandleFunc("DELETE /{itemId}/", th.handleDeleteItem)
	})

	mux.Group("/settings", func(settings *cmux.Mux) {
		settings.Use(authProtectedRedirect)
		settings.HandleFunc("GET /", th.handleSettingsPage)
		settings.HandleFunc("PUT /auth/password/", th.handleUpdateUserPassword)
		settings.HandleFunc("GET /auth/users/", th.handleGetUsers)
		settings.HandleFunc("POST /auth/users/", th.handleCreateUser)
		settings.HandleFunc("DELETE /auth/users/{userId}/", th.handleDeleteUserById)
	})
}

func (th *TemplateHandler) handleIndexPage(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	user, _ := ctx.GetUser(context)
	viewData := views.IndexPageData{
		User: user,
	}
	indexPage := views.IndexPage(viewData)
	indexPage.Render(context, w)
}

func (th *TemplateHandler) handleSignIn(onSuccesRedirect string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			user     = r.FormValue("username")
			password = r.FormValue("password")
		)

		authResult, err := th.authService.SignIn(r.Context(), user, password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprintf("Invalid credentials")))
			return
		}
		cookie := &http.Cookie{}
		cookie.Name = sessionCookieName
		cookie.Value = authResult.Session.Token
		cookie.Expires = time.Unix(authResult.Session.ExpiresAt, 0)
		cookie.Path = "/"
		http.SetCookie(w, cookie)

		http.Redirect(w, r, onSuccesRedirect, http.StatusSeeOther)
	}
}

func (th *TemplateHandler) handleSignOut(redirectTo string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)

		if err == nil {
			token := cookie.Value
			th.authService.SignOut(r.Context(), token)
		}

		http.SetCookie(w, &http.Cookie{
			Name:    sessionCookieName,
			Value:   "",
			Expires: time.Now(),
		})
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	}
}

func (th *TemplateHandler) handleGetItems(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	userId, ok := ctx.GetUserId(context)
	if !ok || len(userId) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	items, _ := th.itemService.ListItemsForUser(context, userId)
	itemsList := views.ItemList(items)
	itemsList.Render(r.Context(), w)
}

func (th *TemplateHandler) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	userId, ok := ctx.GetUserId(context)
	if !ok || len(userId) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	content := r.FormValue("content")
	item, err := th.itemService.CreateItem(context, service.CreateItemsParams{
		Content: content,
		UserId:  userId,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	views.Item(item).Render(context, w)
}

func (th *TemplateHandler) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	userId, ok := ctx.GetUserId(context)
	if !ok || len(userId) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	itemId := r.PathValue("itemId")
	err := th.itemService.DeleteItemForUser(context, service.DeleteUserItemParams{
		ItemId: itemId,
		UserId: userId,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (th *TemplateHandler) handleSettingsPage(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	user, _ := ctx.GetUser(context)
	viewData := views.SettingsPageData{
		User: user,
	}
	settingsPage := views.SettingsPage(viewData)
	settingsPage.Render(context, w)
}

func (th *TemplateHandler) handleUpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	userId, ok := ctx.GetUserId(r.Context())
	if !ok || len(userId) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	password := r.FormValue("password")
	err := th.authService.UpdatePassword(r.Context(), service.UpdatePasswordParams{
		UserId:   userId,
		Password: password,
	})

	if errors.Is(err, service.ErrMinPasswordLength) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Password updated"))
}

func (th *TemplateHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	currentUserId, ok := ctx.GetUserId(r.Context())
	if !ok || len(currentUserId) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	role := models.Role(r.FormValue("role"))

	user, err := th.authService.CreateUser(r.Context(), service.CreateUserParams{
		UserName: username,
		Password: password,
		Role:     role,
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	userRow := views.SettingsUserRow(newSettingsUserRowData(user, currentUserId))
	userRow.Render(r.Context(), w)
}

func (th *TemplateHandler) handleDeleteUserById(w http.ResponseWriter, r *http.Request) {
	err := th.authService.DeleteUserById(r.Context(), r.PathValue("userId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (th *TemplateHandler) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := th.authService.ListUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currentUserId, _ := ctx.GetUserId(r.Context())

	var viewData []views.SettingsUserRowData
	for _, user := range users {
		viewData = append(viewData, newSettingsUserRowData(user, currentUserId))
	}
	rows := views.SettingsUserRows(viewData)
	rows.Render(r.Context(), w)
}

func newSettingsUserRowData(user models.User, currentUserId string) views.SettingsUserRowData {
	return views.SettingsUserRowData{
		User:        user,
		IsDeletable: currentUserId != user.Id,
	}
}
