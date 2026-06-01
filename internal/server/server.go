package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/yodzafar/url-shortener-app/internal/config"
	"github.com/yodzafar/url-shortener-app/internal/handler"
	appMiddleware "github.com/yodzafar/url-shortener-app/internal/middleware"
)

type Server struct {
	echo *echo.Echo
	cfg  *config.Config
}

type Handlers struct {
	Auth *handler.AuthHandler
}

func New(
	cfg *config.Config,
	h Handlers,
	authMW *appMiddleware.AuthMiddleware,
	langMW *appMiddleware.LangMW,
	errH *appMiddleware.ErrorHandler,
) *Server {
	e := echo.New()

	e.HTTPErrorHandler = errH.Handle

	e.Use(middleware.Recover())
	e.Use(middleware.Secure())

	if cfg.IsDevelopment() {
		e.Use(middleware.RequestLogger())
	}

	e.Use(langMW.Handle)

	registerRoutes(e, h, authMW)

	return &Server{echo: e, cfg: cfg}
}

// Handler exposes the underlying HTTP handler (useful for in-process testing).
func (s *Server) Handler() http.Handler {
	return s.echo
}

func (s *Server) Start() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.cfg.Server.Port),
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
		Handler:      s.echo,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)

	go func() {
		log.Printf("→ [%s] http://localhost:%s", s.cfg.App.Env, s.cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server: %w", err)
	case sig := <-quit:
		log.Printf("→ signal %v, shutting down ...", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown: %v", err)
	}

	log.Println("→ server stopped")

	return nil
}
