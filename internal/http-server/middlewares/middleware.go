package middlewares

import (
	"github.com/ilyabukanov123/api-mail/internal/config"
	"net/http"
)

func StaticAuthMiddleware(app *config.App, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != app.Config.AuthorizationToken {
			http.Error(w, "You need to log in to get a unique link. Authorization failed ", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
