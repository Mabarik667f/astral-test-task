package handler

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type UserHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type DocHandler interface {
	Upload(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	DeleteByID(w http.ResponseWriter, r *http.Request)
}

func API(sessionManager *scs.SessionManager, userHandler UserHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"}, // TODO: to config
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.Logger)

	r.Use(sessionManager.LoadAndSave)

	r.Route("/api", func(r chi.Router) {
		UserHandlerRegister(r, userHandler)
	})

	return r
}

func UserHandlerRegister(r chi.Router, h UserHandler) {
	r.Post("/register", h.Register)
	r.Post("/auth", h.Login)
	r.Delete("/auth/{token}", h.Logout)
}

func DocHandlerRegister(r chi.Router, h DocHandler) {
	r.Post("/docs", h.Upload)
	r.Get("/docs", h.Get)
	r.Get("/docs/{id}", h.GetByID)
	r.Delete("/docs/{id}", h.DeleteByID)
}
