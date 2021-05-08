package main

import (
	"context"
	"flag"
	"github.com/patrick246/shortlink/pkg/observability/logging"
	"github.com/patrick246/shortlink/pkg/persistence"
	"github.com/patrick246/shortlink/pkg/persistence/badger"
	"github.com/patrick246/shortlink/pkg/persistence/mongodb"
	"github.com/patrick246/shortlink/pkg/server"
	"github.com/patrick246/shortlink/pkg/server/auth"
	"os"
	"os/signal"
	"syscall"
)

var listenAddr = flag.String("addr", ":8080", "Address and port to listen on")
var storageType = flag.String("storage.type", "mongodb", "Used storage type. Possible values: mongodb, local")
var mongodbUrl = flag.String("storage.mongodb.uri", "mongodb://localhost:27017/shortlink", "MongoDB URI to connect to when using MongoDB storage")
var storagePath = flag.String("storage.local.path", "./storage", "Storage path when using local storage")
var authType = flag.String("auth.type", "none", "Used authentication for admin area. Possible values: none, basic, oidc")
var basicAuthUser = flag.String("auth.basic.user", "admin", "Username for basic authentication")
var basicAuthPassword = flag.String("auth.basic.password", "$2y$12$K7yP/8CraK8RB0yxvv2H4OI6jrC4ym.Xmzx9KQSvqSw3r.3gvtkRu", "Bcrypt password hash for basic authentication")
var oidcIssuer = flag.String("auth.oidc.issuer", "https://idp.example.com", "OpenID Connect issuer used for autodiscovery")
var oidcClientId = flag.String("auth.oidc.client-id", "client", "OpenID Connect Client ID")
var oidcClientSecret = flag.String("auth.oidc.client-secret", "secret", "OpenID Connect Client secret")
var oidcRedirectUri = flag.String("auth.oidc.redirect-uri", "https://shortlink.example.com/oauth2/callback", "Full redirect URI registered at the auth server, path has to be /oauth2/callback")

var log = logging.CreateLogger("main")

func main() {
	flag.Parse()

	var repo persistence.Repository
	switch *storageType {
	case "mongodb":
		dbConn, err := mongodb.NewConnection(*mongodbUrl)
		if err != nil {
			log.Fatalw("db connection error", "uri", *mongodbUrl, "error", err)
		}

		repo, err = mongodb.New(dbConn)
		if err != nil {
			log.Fatalw("repo error", "error", err)
		}

	case "local":
		var err error
		repo, err = badger.New(*storagePath)
		if err != nil {
			log.Fatalw("local storage error", "path", *storagePath, "error", err)
		}
	default:
		log.Fatalw("unknown storage type", "type", *storageType)
	}
	defer repo.Close()

	var authMiddleware server.MiddlewareFactory
	log.Infow("setting up authentication", "type", *authType)

	switch *authType {
	case "none":
		authMiddleware = auth.Noop()
	case "basic":
		authMiddleware = auth.BasicAuth(*basicAuthUser, *basicAuthPassword, "/admin/shortlinks")
	case "oidc":
		var err error
		authMiddleware, err = auth.OpenIDConnect(auth.OidcConfig{
			Issuer:       *oidcIssuer,
			ClientId:     *oidcClientId,
			ClientSecret: *oidcClientSecret,
			RedirectUri:  *oidcRedirectUri,
		}, "/admin/shortlinks")
		if err != nil {
			log.Fatalw("oidc error", "issuer", *oidcIssuer, "clientId", *oidcClientId, "redirectUri", *oidcRedirectUri, "error", err)
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, os.Interrupt)

	runCtx, cancel := context.WithCancel(context.Background())
	go func() {
		<-signalChan
		cancel()
	}()

	shortlinkServer := server.New(*listenAddr, repo, authMiddleware)
	err := shortlinkServer.ListenAndServe(runCtx)
	if err != nil {
		log.Fatalw("server error", "addr", *listenAddr, "error", err)
	}
}
