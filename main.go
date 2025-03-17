package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/michaelhass/cpaw/db"
	"github.com/michaelhass/cpaw/db/repository"
	"github.com/michaelhass/cpaw/handler"
	"github.com/michaelhass/cpaw/middleware"
	"github.com/michaelhass/cpaw/mux"
	"github.com/michaelhass/cpaw/service"
	"github.com/michaelhass/cpaw/views"
)

func main() {
	db, err := db.NewSqlite(
		db.WithDbName("cpaw"),
		db.WithDbPath("cpaw.db"),
	)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()
	if err := db.MigrateUp(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
		return
	}
	if err := db.SetUp(); err != nil {
		log.Fatal(err)
		return
	}

	userRepository := repository.NewUserRepository(db.DB)
	sessionRespository := repository.NewSessionRespository(db.DB)
	authService := service.NewAuthService(sessionRespository, userRepository)

	initialCredentials, err := authService.SetUp(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if len(initialCredentials.Id) > 0 {
		const logMsg string = "--- Created initial user. ---\nid: %s, name: %s, pw: %s\n"
		log.Printf(
			logMsg,
			initialCredentials.Id,
			initialCredentials.UserName,
			initialCredentials.Password,
		)
	}
	log.Println("Auth service is ready.")

	mainMux := mux.NewDefaultMux()
	mainMux.Use(middleware.Logger)

	mainMux.Handle("/assets/css/", http.StripPrefix("/assets/css/", http.FileServer(http.Dir("views/assets/css"))))

	mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loggedIn := false
		if !loggedIn {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		component := views.WithDefaultPage(views.Index())
		component.Render(r.Context(), w)
	})

	mainMux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		component := views.WithDefaultPage(views.Login())
		component.Render(r.Context(), w)
	})

	mainMux.HandleFunc("GET /signup", func(w http.ResponseWriter, r *http.Request) {
		component := views.WithDefaultPage(views.SignUp())
		component.Render(r.Context(), w)
	})

	mainMux.Group("/api/v1", func(api *mux.Mux) {
		apiHandler := handler.NewApiHandler(authService)
		apiHandler.RegisterRoutes(api)

	})

	const port string = ":3000"
	log.Println("Starting server at port", port)
	if err := http.ListenAndServe(port, mainMux); err != nil {
		log.Fatal(err)
	}
}
