package main

import (
	"log"
	"net/http"

	"github.com/michaelhass/cpaw/handlers"
	"github.com/michaelhass/cpaw/mux"
	"github.com/michaelhass/cpaw/views"
)

func main() {
	mainMux := mux.NewDefaultMux()
	mainMux.Use(handlers.Logger)

	mainMux.Handle("/assets/css/", http.StripPrefix("/assets/css/", http.FileServer(http.Dir("views/assets/css"))))

	mainMux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		component := views.WithDefaultPage(views.Login())
		component.Render(r.Context(), w)
	})

	mainMux.HandleFunc("GET /signup", func(w http.ResponseWriter, r *http.Request) {
		component := views.WithDefaultPage(views.SignUp())
		component.Render(r.Context(), w)
	})

	mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loggedIn := false
		if !loggedIn {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		component := views.WithDefaultPage(views.Index())
		component.Render(r.Context(), w)
	})

	// mainMux.Group("/api/v1", func(api *mux.Mux) {
	// 	api.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
	// 		w.Write([]byte("user"))
	// 	})

	// 	api.Group("/items", func(items *mux.Mux) {
	// 		items.Use(handler.AuthHandler)
	// 		items.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
	// 			w.Write([]byte("items"))
	// 		})
	// 		items.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
	// 			w.Write([]byte("items with id"))
	// 		})
	// 	})
	// })

	if err := http.ListenAndServe(":3000", mainMux); err != nil {
		log.Fatal(err)
	}
}
