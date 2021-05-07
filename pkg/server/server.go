package server

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"github.com/patrick246/shortlink/pkg/observability/logging"
	"github.com/patrick246/shortlink/pkg/persistence"
	"net/http"
	"time"
)

var shutdownTimeout = 5 * time.Second

var log = logging.CreateLogger("server")

type Server struct {
	router *httprouter.Router
	server http.Server
	repo   persistence.Repository
}

type MiddlewareFactory func(next http.Handler) http.Handler

func New(addr string, repo persistence.Repository, authMiddleware MiddlewareFactory) *Server {
	router := httprouter.New()

	server := &Server{
		repo:   repo,
		router: router,
		server: http.Server{
			Addr:         addr,
			Handler:      authMiddleware(router),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

	router.Handler(http.MethodGet, "/static/*filepath", http.FileServer(http.FS(staticContent)))
	router.GET("/admin/shortlinks", server.listShortlinks)
	router.POST("/admin/shortlinks", server.createOrEdit)
	router.GET("/admin/shortlinks/:code", server.editShortlink)
	router.POST("/admin/shortlinks/:code", server.createOrEdit)
	router.POST("/admin/shortlinks/:code/delete", server.deleteShortlink)

	router.NotFound = http.HandlerFunc(server.handleCodeRequests)

	return server
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		log.Infow("shutting down server", "timeout", shutdownTimeout)
		_ = s.server.Shutdown(shutdownCtx)
	}()

	log.Infow("listening", "addr", s.server.Addr)
	err := s.server.ListenAndServe()
	if err == http.ErrServerClosed {
		time.Sleep(shutdownTimeout)
	} else if err != nil {
		return err
	}
	return nil
}
