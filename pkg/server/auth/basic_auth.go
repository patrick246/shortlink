package auth

import (
	"crypto/subtle"
	"github.com/patrick246/shortlink/pkg/server"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

func BasicAuth(username, passwordHash, securedPrefix string) server.MiddlewareFactory {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if !strings.HasPrefix(request.URL.Path, securedPrefix) {
				next.ServeHTTP(writer, request)
				return
			}

			reqUser, reqPassword, ok := request.BasicAuth()
			if !ok {
				writer.Header().Set("www-authenticate", `Basic realm="/admin"`)
				writer.WriteHeader(401)
				return
			}

			passwordCorrect := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(reqPassword)) == nil
			usernameCorrect := subtle.ConstantTimeCompare([]byte(username), []byte(reqUser)) == 1

			if !usernameCorrect || !passwordCorrect {
				writer.Header().Set("www-authenticate", `Basic realm="/admin"`)
				writer.WriteHeader(401)
				return
			}

			next.ServeHTTP(writer, request)
		})
	}

}
