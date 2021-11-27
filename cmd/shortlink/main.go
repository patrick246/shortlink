package main

import (
	"context"
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

var log = logging.CreateLogger("main")

func main() {
	conf := getConfig()

	var repo persistence.Repository
	switch conf.StorageType {
	case "mongodb":
		dbConn, err := mongodb.NewConnection(conf.MongoDbUri)
		if err != nil {
			log.Fatalw("db connection error", "uri", conf.MongoDbUri, "error", err)
		}

		repo, err = mongodb.New(dbConn)
		if err != nil {
			log.Fatalw("repo error", "error", err)
		}

	case "local":
		var err error
		repo, err = badger.New(conf.StoragePath)
		if err != nil {
			log.Fatalw("local storage error", "path", conf.StoragePath, "error", err)
		}
	default:
		log.Fatalw("unknown storage type", "type", conf.StorageType)
	}
	defer repo.Close()

	err := repo.Migrate(context.Background())
	if err != nil {
		log.Fatalw("migration error", "error", err)
	}

	var authMiddleware server.MiddlewareFactory
	log.Infow("setting up authentication", "type", conf.AuthType)

	switch conf.AuthType {
	case "none":
		authMiddleware = auth.Noop()
	case "basic":
		authMiddleware = auth.BasicAuth(conf.BasicAuthUser, conf.BasicAuthPassword, "/admin/shortlinks")
	case "oidc":
		var err error
		authMiddleware, err = auth.OpenIDConnect(auth.OidcConfig{
			Issuer:       conf.OidcIssuer,
			ClientId:     conf.OidcClientId,
			ClientSecret: conf.OidcClientSecret,
			RedirectUri:  conf.OidcRedirectUri,
		}, "/admin/shortlinks")
		if err != nil {
			log.Fatalw("oidc error", "issuer", conf.OidcIssuer, "clientId", conf.OidcClientId, "redirectUri", conf.OidcRedirectUri, "error", err)
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, os.Interrupt)

	runCtx, cancel := context.WithCancel(context.Background())
	go func() {
		<-signalChan
		cancel()
	}()

	shortlinkServer := server.New(conf.ListenAddr, repo, authMiddleware)
	err = shortlinkServer.ListenAndServe(runCtx)
	if err != nil {
		log.Fatalw("server error", "addr", conf.ListenAddr, "error", err)
	}
}
