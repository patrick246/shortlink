package auth

import (
	"github.com/patrick246/shortlink/pkg/server"
	"net/http"
)

func Noop() server.MiddlewareFactory {
	return func(next http.Handler) http.Handler {
		return next
	}
}
