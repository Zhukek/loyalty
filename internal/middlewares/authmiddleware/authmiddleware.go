package authmiddleware

import (
	"context"
	"net/http"

	"github.com/Zhukek/loyalty/internal/middlewares"
	"github.com/Zhukek/loyalty/internal/utils"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("jwt")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		user, err := utils.GetTokenData(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), middlewares.UserKey, user)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
