package main

import (
	"flag"
	"os"
)

type config struct {
	ListenAddr  string
	StorageType string

	// MongoDB Storage
	MongoDbUri string

	// Badger storage
	StoragePath string

	AuthType string

	// Basic auth
	BasicAuthUser     string
	BasicAuthPassword string

	// OpenId Connect Auth
	OidcIssuer       string
	OidcClientId     string
	OidcClientSecret string
	OidcRedirectUri  string
}

func getConfig() config {
	listenAddrFlag := flag.String("addr", ":8080", "Address and port to listen on")
	storageTypeFlag := flag.String("storage.type", "mongodb", "Used storage type. Possible values: mongodb, local")
	mongodbUrlFlag := flag.String("storage.mongodb.uri", "mongodb://localhost:27017/shortlink", "MongoDB URI to connect to when using MongoDB storage")
	storagePathFlag := flag.String("storage.local.path", "./storage", "Storage path when using local storage")
	authTypeFlag := flag.String("auth.type", "none", "Used authentication for admin area. Possible values: none, basic, oidc")
	basicAuthUserFlag := flag.String("auth.basic.user", "admin", "Username for basic authentication")
	basicAuthPasswordFlag := flag.String("auth.basic.password", "$2y$12$K7yP/8CraK8RB0yxvv2H4OI6jrC4ym.Xmzx9KQSvqSw3r.3gvtkRu", "Bcrypt password hash for basic authentication")
	oidcIssuerFlag := flag.String("auth.oidc.issuer", "https://idp.example.com", "OpenID Connect issuer used for autodiscovery")
	oidcClientIdFlag := flag.String("auth.oidc.client-id", "client", "OpenID Connect Client ID")
	oidcClientSecretFlag := flag.String("auth.oidc.client-secret", "secret", "OpenID Connect Client secret")
	oidcRedirectUriFlag := flag.String("auth.oidc.redirect-uri", "https://shortlink.example.com/oauth2/callback", "Full redirect URI registered at the auth server, path has to be /oauth2/callback")
	flag.Parse()

	listenAddrEnv := os.Getenv("LISTEN_ADDR")
	storageTypeEnv := os.Getenv("STORAGE_TYPE")
	mongodbUrlEnv := os.Getenv("STORAGE_MONGODB_URI")
	storagePathEnv := os.Getenv("STORAGE_LOCAL_PATH")
	authTypeEnv := os.Getenv("AUTH_TYPE")
	basicAuthUserEnv := os.Getenv("AUTH_BASIC_USER")
	basicAuthPasswordEnv := os.Getenv("AUTH_BASIC_PASSWORD")
	oidcIssuerEnv := os.Getenv("AUTH_OIDC_ISSUER")
	oidcClientIdEnv := os.Getenv("AUTH_OIDC_CLIENTID")
	oidcClientSecretEnv := os.Getenv("AUTH_OIDC_CLIENTSECRET")
	oidcRedirectUriEnv := os.Getenv("AUTH_OIDC_REDIRECTURI")

	return config{
		ListenAddr:        flagOrEnv(*listenAddrFlag, listenAddrEnv, ":8080"),
		StorageType:       flagOrEnv(*storageTypeFlag, storageTypeEnv, "mongodb"),
		MongoDbUri:        flagOrEnv(*mongodbUrlFlag, mongodbUrlEnv, "mongodb://localhost:27017/shortlink"),
		StoragePath:       flagOrEnv(*storagePathFlag, storagePathEnv, "./storage"),
		AuthType:          flagOrEnv(*authTypeFlag, authTypeEnv, "none"),
		BasicAuthUser:     flagOrEnv(*basicAuthUserFlag, basicAuthUserEnv, "admin"),
		BasicAuthPassword: flagOrEnv(*basicAuthPasswordFlag, basicAuthPasswordEnv, "$2y$12$K7yP/8CraK8RB0yxvv2H4OI6jrC4ym.Xmzx9KQSvqSw3r.3gvtkRu"),
		OidcIssuer:        flagOrEnv(*oidcIssuerFlag, oidcIssuerEnv, "https://idp.example.com"),
		OidcClientId:      flagOrEnv(*oidcClientIdFlag, oidcClientIdEnv, "client"),
		OidcClientSecret:  flagOrEnv(*oidcClientSecretFlag, oidcClientSecretEnv, "secret"),
		OidcRedirectUri:   flagOrEnv(*oidcRedirectUriFlag, oidcRedirectUriEnv, "https://shortlink.example.com/oauth2/callback"),
	}
}

func flagOrEnv(flag, env, defaultValue string) string {
	if flag == defaultValue && env != "" {
		return env
	}
	return flag
}
