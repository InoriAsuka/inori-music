package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	repository, err := storageRepository()
	if err != nil {
		log.Fatal(err)
	}
	storageService := storage.NewService(repository)
	mediaObjectService := storage.NewMediaObjectService(repository, storage.NewMemoryMediaObjectRepository())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if interval := storageRefreshInterval(); interval > 0 {
		log.Printf("storage refresh scheduler enabled with interval %s", interval)
		scheduler := storage.NewRefreshScheduler(storageService, interval, func(report storage.RefreshReport, err error) {
			if err != nil {
				log.Printf("storage refresh failed: %v", err)
				return
			}
			log.Printf("storage refresh completed for %d backends", len(report.Results))
		})
		go scheduler.Run(ctx)
	}

	server := &http.Server{
		Addr:              address,
		Handler:           httpapi.NewHandler(storageService, httpapi.WithAdminToken(adminToken), httpapi.WithMediaObjectService(mediaObjectService)).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("http server shutdown failed: %v", err)
		}
	}()

	log.Printf("inori-music api server listening on %s", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func storageRefreshInterval() time.Duration {
	value := os.Getenv("INORI_STORAGE_REFRESH_INTERVAL")
	if value == "" {
		return 0
	}
	interval, err := time.ParseDuration(value)
	if err != nil || interval <= 0 {
		log.Printf("ignoring invalid INORI_STORAGE_REFRESH_INTERVAL %q", value)
		return 0
	}
	return interval
}

func storageRepository() (storage.Repository, error) {
	path := os.Getenv("INORI_STORAGE_REPOSITORY_FILE")
	if path == "" {
		return storage.NewMemoryRepository(), nil
	}
	log.Printf("storage repository file enabled at %s", path)
	return storage.NewFileRepository(path)
}
