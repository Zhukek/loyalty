package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Zhukek/loyalty/internal/logger"
	"github.com/Zhukek/loyalty/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func ping(w http.ResponseWriter, logger logger.Logger, rep repository.Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := rep.Ping(ctx)
	if err != nil {
		logger.LogErr(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func NewRouter(logger logger.Logger, repository repository.Repository) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ping(w, logger, repository)
	})

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {

		})
		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {

		})
		r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {

		})
		r.Get("/orders", func(w http.ResponseWriter, r *http.Request) {

		})
		r.Get("/withdrawals", func(w http.ResponseWriter, r *http.Request) {

		})
		r.Route("/balance", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {

			})
			r.Post("/withdraw", func(w http.ResponseWriter, r *http.Request) {

			})
		})
	})

	return router
}
