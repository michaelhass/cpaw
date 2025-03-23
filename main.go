package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/michaelhass/cpaw/db"
	"github.com/michaelhass/cpaw/db/repository"
	"github.com/michaelhass/cpaw/handler"
	"github.com/michaelhass/cpaw/middleware"
	"github.com/michaelhass/cpaw/mux"
	"github.com/michaelhass/cpaw/service"
	"github.com/michaelhass/cpaw/views"
	"golang.org/x/sync/errgroup"
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

	defer func() {
		db.Close()
		log.Println("DB closed")
	}()

	if err := db.SetUp(); err != nil {
		log.Fatal(err)
		return
	}

	userRepository := repository.NewUserRepository(db.DB)
	sessionRespository := repository.NewSessionRespository(db.DB)
	itemRepository := repository.NewItemRepository(db.DB)

	authService := service.NewAuthService(sessionRespository, userRepository)
	itemService := service.NewItemService(itemRepository)

	initialCredentials, err := authService.SetUp(context.Background())
	if err != nil {
		log.Fatal("Error setting up auth services", err)
		return
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
	log.Println("Services are ready.")

	mainMux := mux.NewDefaultMux()
	mainMux.Use(middleware.Logger)
	mainMux.Use(middleware.Recover)

	mainMux.Handle("/assets/css/", http.StripPrefix("/assets/css/", http.FileServer(http.Dir("views/assets/css"))))

	mainMux.HandleFunc("GET /login/", func(w http.ResponseWriter, r *http.Request) {
		component := views.WithDefaultPage(views.Login())
		component.Render(r.Context(), w)
	})

	mainMux.HandleFunc("GET /signup/", func(w http.ResponseWriter, r *http.Request) {
		component := views.WithDefaultPage(views.SignUp())
		component.Render(r.Context(), w)
	})

	mainMux.Group("/api/v1", func(apiMux *mux.Mux) {
		apiHandler := handler.NewApiHandler(authService, itemService)
		apiHandler.RegisterRoutes(apiMux)
	})

	const addr string = ":3000"
	listenAndServe(addr, mainMux)
}

func listenAndServe(addr string, mux *mux.Mux) {
	log.Println("Starting server at addr", addr)

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return mainCtx
		},
	}

	errGroup, groupCtx := errgroup.WithContext(mainCtx)

	errGroup.Go(func() error {
		return server.ListenAndServe()
	})

	errGroup.Go(func() error {
		<-groupCtx.Done()
		return server.Shutdown(context.Background())
	})

	if err := errGroup.Wait(); err != nil {
		log.Println("Exit:", err)
	}
}
