//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	appi18n "github.com/yodzafar/url-shortener-app/i18n"
	"github.com/yodzafar/url-shortener-app/internal/config"
	"github.com/yodzafar/url-shortener-app/internal/handler"
	appMiddleware "github.com/yodzafar/url-shortener-app/internal/middleware"
	"github.com/yodzafar/url-shortener-app/internal/pkg/db"
	"github.com/yodzafar/url-shortener-app/internal/pkg/token"
	"github.com/yodzafar/url-shortener-app/internal/pkg/validation"
	"github.com/yodzafar/url-shortener-app/internal/repository"
	"github.com/yodzafar/url-shortener-app/internal/server"
	"github.com/yodzafar/url-shortener-app/internal/usecase"
)

var DBSet = wire.NewSet(db.NewPostgresDb)

var RepositorySet = wire.NewSet(
	repository.NewUserRepository,
)

var I18nSet = wire.NewSet(
	provideTranslator,
)

var TokenSet = wire.NewSet(
	provideTokenManager,
)

var ValidationSet = wire.NewSet(
	validation.New,
)

var UsecaseSet = wire.NewSet(
	usecase.NewAuthUsecase,
	usecase.NewUserUsecase,
)

var HandlerSet = wire.NewSet(
	handler.NewAuthHandler,
	handler.NewUserHandler,
	provideHandlers,
)

var MiddlewareSet = wire.NewSet(
	appMiddleware.NewAuthMiddleware,
	appMiddleware.NewLangMiddleware,
	appMiddleware.NewErrorHandler,
)

func InitializerServer(cfg *config.Config) (*server.Server, func(), error) {
	wire.Build(
		DBSet,
		RepositorySet,
		I18nSet,
		TokenSet,
		ValidationSet,
		UsecaseSet,
		HandlerSet,
		MiddlewareSet,
		server.New,
	)

	return nil, nil, nil
}

func provideTokenManager(cfg *config.Config) *token.Manager {
	return token.NewManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
}

func provideTranslator(cfg *config.Config) (*appi18n.Translator, error) {
	return appi18n.New(cfg.App.LocalesDir)
}

func provideHandlers(auth *handler.AuthHandler, user *handler.UserHandler) server.Handlers {
	return server.Handlers{
		Auth: auth,
		User: user,
	}
}
