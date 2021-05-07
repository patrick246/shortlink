# Shortlink
Go shortlink server. Add shortlink codes in the admin UI, afterwards, the server will redirect when someone navigates to /:code.

Admin-Path: /admin/shortlinks

## Usage
 - Decide between storage backends: Local storage using badger, or MongoDB storage 
 - Decide between admin authentication methods: None, Basic Auth or OpenID Connect 
```
Usage of ./shortlink:
  -addr string
        Address and port to listen on (default ":8080")
  -auth.basic.password string
        Bcrypt password hash for basic authentication (default "$2y$12$K7yP/8CraK8RB0yxvv2H4OI6jrC4ym.Xmzx9KQSvqSw3r.3gvtkRu")
  -auth.basic.user string
        Username for basic authentication (default "admin")
  -auth.oidc.client-id string
        OpenID Connect Client ID (default "client")
  -auth.oidc.client-secret string
        OpenID Connect Client secret (default "secret")
  -auth.oidc.issuer string
        OpenID Connect issuer used for autodiscovery (default "https://idp.example.com")
  -auth.oidc.redirect-uri string
        Full redirect URI registered at the auth server, path has to be /oauth2/callback (default "https://shortlink.example.com/oauth2/callback")
  -auth.type string
        Used authentication for admin area. Possible values: none, basic, oidc (default "none")
  -storage.local.path string
        Storage path when using local storage (default "./storage")
  -storage.mongodb.uri string
        MongoDB URI to connect to when using MongoDB storage (default "mongodb://localhost:27017/shortlink")
  -storage.type string
        Used storage type. Possible values: mongodb, local (default "mongodb")
```
