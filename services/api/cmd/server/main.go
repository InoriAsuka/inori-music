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
	address := os.Getenv("INORI_HTTP_ADDR")
	if address == "" {
		address = "127.0.0.1:8080"
	}
	adminToken := os.Getenv("INORI_ADMIN_TOKEN")
	if adminToken == "" {
		log.Print("INORI_ADMIN_TOKEN is not set; /api/v1/admin routes will return 503")
	}

	storageService := storage.NewService(storage.NewMemoryRepository())
	server := &http.Server{
		Addr:              address,
		Handler:           httpapi.NewHandler(storageService, httpapi.WithAdminToken(adminToken)).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("inori-music api server listening on %s", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
