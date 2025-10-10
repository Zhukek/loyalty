package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Zhukek/loyalty/internal/errs"
	"github.com/Zhukek/loyalty/internal/logger"
	"github.com/Zhukek/loyalty/internal/models"
	"github.com/Zhukek/loyalty/internal/repository"
	"github.com/Zhukek/loyalty/internal/utils"
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

func create(res http.ResponseWriter, req *http.Request, logger logger.Logger, rep repository.Repository) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")

	user := models.User{}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.LogErr(
			"target", "read body",
			"error", err,
		)

		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, &user); err != nil {
		logger.LogErr(
			"target", "json unmarshal",
			"error", err,
		)

		res.WriteHeader(http.StatusBadRequest)
		return
	}

	user.Pass, err = utils.HashPass(user.Pass)

	if err != nil {
		logger.LogErr(
			"target", "hash pass",
			"error", err,
		)

		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	createdUser, err := rep.CreateUser(user.Log, user.Pass, context.Background())

	if errors.Is(err, errs.ErrUsernameTaken) {
		res.WriteHeader(http.StatusConflict)
		return
	} else if err != nil {
		logger.LogErr(
			"target", "create user",
			"error", err,
		)

		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Добавить установку JWT

	res.WriteHeader(http.StatusOK)
}

func NewRouter(logger logger.Logger, repository repository.Repository) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ping(w, logger, repository)
	})

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			create(w, r, logger, repository)
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
