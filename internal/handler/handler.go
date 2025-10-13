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
	"github.com/Zhukek/loyalty/internal/middlewares/authmiddleware"
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
		logger.LogErr("ping database", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func create(w http.ResponseWriter, req *http.Request, logger logger.Logger, rep repository.Repository) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	user := models.User{}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.LogErr("read body", err)

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, &user); err != nil {
		logger.LogErr("json unmarshal", err)

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user.Pass, err = utils.HashPass(user.Pass)

	if err != nil {
		logger.LogErr("hash pass", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	createdUser, err := rep.CreateUser(user.Log, user.Pass, context.Background())

	if errors.Is(err, errs.ErrUsernameTaken) {
		w.WriteHeader(http.StatusConflict)
		return
	} else if err != nil {
		logger.LogErr("create user", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwt, err := utils.GenerateJWT(createdUser)

	if err != nil {
		logger.LogErr("generate jwt", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := utils.GenerateJWTCookie(jwt)

	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
}

func auth(w http.ResponseWriter, req *http.Request, logger logger.Logger, rep repository.Repository) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	user := models.User{}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.LogErr("read body", err)

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, &user); err != nil {
		logger.LogErr("json unmarshal", err)

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	foundUser, err := rep.GetUserByName(user.Log, context.Background())
	if err != nil {
		logger.LogInfo("get user", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := utils.CheckPass(foundUser.Pass, user.Pass); err != nil {
		logger.LogInfo("check pass", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	jwt, err := utils.GenerateJWT(&foundUser.UserPublic)
	if err != nil {
		logger.LogErr("generate jwt", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := utils.GenerateJWTCookie(jwt)
	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
}

func newOrder(w http.ResponseWriter, req *http.Request, logger logger.Logger, rep repository.Repository) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

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
			create(w, r, logger, repository)
		})
		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			auth(w, r, logger, repository)
		})

		r.Group(func(r chi.Router) {
			r.Use(authmiddleware.AuthMiddleware)

			r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
				newOrder(w, r, logger, repository)
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
	})

	return router
}
