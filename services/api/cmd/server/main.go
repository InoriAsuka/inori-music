package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"inori-music/services/api/internal/httpapi"
	"inori-music/services/api/internal/storage"
)

func main() {
	config, err := loadServerConfig(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	if config.InsecureDevAuth && config.AdminToken == "" {
		log.Print("warning: INORI_INSECURE_DEV_AUTH=1 disables admin bearer authentication; use only for local development")
	}

	storageService := storage.NewService(storage.NewMemoryRepository())
	options := []httpapi.Option{}
	if config.AdminToken != "" {
		options = append(options, httpapi.WithAdminToken(config.AdminToken))
	}
	if config.InsecureDevAuth && config.AdminToken == "" {
		options = append(options, httpapi.WithInsecureAdminAuth())
	}

	server := &http.Server{
		Addr:              config.Address,
		Handler:           httpapi.NewHandler(storageService, options...).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("inori-music api server listening on %s", config.Address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
