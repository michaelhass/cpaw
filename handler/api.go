package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/michaelhass/cpaw/constant"
	"github.com/michaelhass/cpaw/ctx"
	"github.com/michaelhass/cpaw/db/repository"
	"github.com/michaelhass/cpaw/middleware"
	cmux "github.com/michaelhass/cpaw/mux"
	"github.com/michaelhass/cpaw/service"
)

type ApiHandler struct {
	authService *service.AuthService
	itemService *service.ItemService
}

func NewApiHandler(
	authService *service.AuthService,
	itemService *service.ItemService,
) *ApiHandler {
	return &ApiHandler{
		authService: authService,
		itemService: itemService,
	}
}

func (api *ApiHandler) RegisterRoutes(mux *cmux.Mux) {
	mux.HandleFunc("GET /signin", api.handleSignIn)
	mux.HandleFunc("GET /signout", api.handleSignOut)

	mux.Group("/items", func(m *cmux.Mux) {
		m.Use(middleware.AuthProtected(api.authService))
		m.HandleFunc("GET /", api.handleListUserItems)
		m.HandleFunc("POST /", api.handleCreateItemForUser)
		m.HandleFunc("GET /{itemId}/", api.handleGetUserItem)
		m.HandleFunc("DELETE /{itemId}/", api.handleDeleteUserItemById)
	})
}

type signInRequest struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

func (api *ApiHandler) handleSignIn(w http.ResponseWriter, r *http.Request) {
	var signInRequest signInRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&signInRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	authResult, err := api.authService.SignIn(r.Context(), signInRequest.UserName, signInRequest.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}
	cookie := &http.Cookie{}
	cookie.Name = constant.SessionCookieName
	cookie.Value = authResult.Session.Token
	cookie.Expires = time.Unix(authResult.Session.ExpiresAt, 0)
	http.SetCookie(w, cookie)

	writeJSONResponse(w, authResult.User, http.StatusAccepted)
}

func (api *ApiHandler) handleSignOut(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(constant.SessionCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		log.Println("no cookie")
		w.WriteHeader(http.StatusOK)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := cookie.Value
	api.authService.SignOut(r.Context(), token)
	http.SetCookie(w, &http.Cookie{
		Name:    constant.SessionCookieName,
		Value:   "",
		Expires: time.Now(),
	})
	w.WriteHeader(http.StatusOK)
}

func (api *ApiHandler) handleGetUserItem(w http.ResponseWriter, r *http.Request) {
	userId, ok := ctx.GetUserId(r.Context())
	if !ok && len(userId) > 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	itemId := r.PathValue("itemId")
	if len(itemId) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	item, err := api.itemService.GetItemForUser(r.Context(), service.GetItemForUserParams{
		ItemId: itemId,
		UserId: userId,
	})

	if errors.Is(err, repository.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, item, http.StatusOK)
}

func (api *ApiHandler) handleListUserItems(w http.ResponseWriter, r *http.Request) {
	userId, ok := ctx.GetUserId(r.Context())
	if !ok && len(userId) > 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	items, err := api.itemService.ListItemsForUser(r.Context(), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, items, http.StatusOK)
}

type createItemRequestBody struct {
	Content string `json:"content"`
}

func (api *ApiHandler) handleCreateItemForUser(w http.ResponseWriter, r *http.Request) {
	userId, ok := ctx.GetUserId(r.Context())
	if !ok && len(userId) > 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var body createItemRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	item, err := api.itemService.CreateItem(r.Context(), repository.CreateItemParams{
		Content: body.Content,
		UserId:  userId,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, item, http.StatusCreated)
}

func (api *ApiHandler) handleDeleteUserItemById(w http.ResponseWriter, r *http.Request) {
	userId, ok := ctx.GetUserId(r.Context())
	if !ok && len(userId) > 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	itemId := r.PathValue("itemId")
	if len(itemId) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := api.itemService.DeleteItemForUser(r.Context(), service.DeleteUserItemParams{
		ItemId: itemId,
		UserId: userId,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}
