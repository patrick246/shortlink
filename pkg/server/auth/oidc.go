package auth

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/patrick246/shortlink/pkg/server"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
	"time"
)

type OidcConfig struct {
	Issuer       string
	ClientId     string
	ClientSecret string
	RedirectUri  string
}

const authCookieName = "__Host-Authentication"
const stateCookieName = "__Host-State"

func OpenIDConnect(config OidcConfig, securedPrefix string) (server.MiddlewareFactory, error) {
	setupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	provider, err := oidc.NewProvider(setupCtx, config.Issuer)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.ClientId,
	})

	oauth2Config := oauth2.Config{
		ClientID:     config.ClientId,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectUri,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path == "/oauth2/callback" {
				stateCookie, err := request.Cookie(stateCookieName)
				if err != nil {
					http.Error(writer, "State cookie not present", http.StatusBadRequest)
					return
				}

				urlState := request.URL.Query().Get("state")
				if stateCookie.Value != urlState {
					http.Error(writer, "Mismatching state", http.StatusBadRequest)
					return
				}

				oauth2Token, err := oauth2Config.Exchange(request.Context(), request.URL.Query().Get("code"))
				if err != nil {
					http.Error(writer, "Error exchanging code", http.StatusInternalServerError)
					return
				}

				rawIDToken, ok := oauth2Token.Extra("id_token").(string)
				if !ok {
					http.Error(writer, "Error getting ID token", http.StatusInternalServerError)
					return
				}

				idToken, err := verifier.Verify(request.Context(), rawIDToken)
				if err != nil {
					http.Error(writer, "Got invalid token", http.StatusInternalServerError)
					return
				}

				http.SetCookie(writer, &http.Cookie{
					Name:     authCookieName,
					Value:    rawIDToken,
					Expires:  idToken.Expiry,
					Path:     "/",
					Secure:   true,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
				http.Redirect(writer, request, "/admin/shortlinks", http.StatusFound)
				return
			}

			if !strings.HasPrefix(request.URL.Path, securedPrefix) {
				next.ServeHTTP(writer, request)
				return
			}

			authenticated := oidcCheckAuthenticated(request, verifier)
			if !authenticated {
				state := uuid.New().String()
				http.SetCookie(writer, &http.Cookie{
					Name:     stateCookieName,
					Value:    state,
					MaxAge:   int(time.Minute.Seconds()),
					Path:     "/",
					Secure:   true,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
				http.Redirect(writer, request, oauth2Config.AuthCodeURL(state), http.StatusFound)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}, nil
}

func oidcCheckAuthenticated(request *http.Request, verifier *oidc.IDTokenVerifier) bool {
	authCookie, err := request.Cookie(authCookieName)
	if err != nil {
		return false
	}

	_, err = verifier.Verify(request.Context(), authCookie.Value)
	if err != nil {
		return false
	}
	return true
}
