package handler

import (
	"io"
	"net/http"

	"github.com/Zhukek/loyalty/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(logger logger.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Oke")
	})

	return router
}
