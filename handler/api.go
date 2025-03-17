package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/michaelhass/cpaw/mux"
	"github.com/michaelhass/cpaw/service"
)

type ApiHandler struct {
	AuthService *service.AuthService
}

func NewApiHandler(authService *service.AuthService) *ApiHandler {
	return &ApiHandler{
		AuthService: authService,
	}
}

func (api *ApiHandler) RegisterRoutes(mux *mux.Mux) {
	mux.HandleFunc("GET /signin", api.handleSignIn)
	mux.HandleFunc("GET /signout", api.handleSignOut)

	// 	api.Group("/items", func(items *mux.Mux) {
	// 		items.Use(handler.AuthHandler)
	// 		items.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
	// 			w.Write([]byte("items"))
	// 		})
	// 		items.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
	// 			w.Write([]byte("items with id"))
	// 		})
	// 	})
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

	authResult, err := api.AuthService.SignIn(r.Context(), signInRequest.UserName, signInRequest.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}
	cookie := &http.Cookie{}
	cookie.Name = "cpaw_session"
	cookie.Value = authResult.Session.Token
	cookie.Expires = time.Unix(authResult.Session.ExpiresAt, 0)
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(authResult.User)
}

func (api *ApiHandler) handleSignOut(w http.ResponseWriter, r *http.Request) {

}
