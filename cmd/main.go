package main

import (
	"log"

	"github.com/yodzafar/url-shortener-app/internal/config"
	appWire "github.com/yodzafar/url-shortener-app/internal/wire"

	_ "github.com/yodzafar/url-shortener-app/docs"
)

//	@title			URL Shortener API
//	@version		1.0
//	@description	REST API for the URL shortener service.

//	@BasePath	/

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and the JWT access token.

func main() {
	cfg := config.Mustload()

	srv, cleanup, err := appWire.InitializerServer(cfg)

	if err != nil {
		log.Fatalf("wire: %v", err)
	}

	defer cleanup()

	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
