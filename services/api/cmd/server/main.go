package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/httpapi"
	"inori-music/services/api/internal/storage"
	pgstore "inori-music/services/api/internal/storage/postgres"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Open a single shared pool if INORI_DATABASE_URL is set.
	pool, err := openDatabasePool(ctx)
	if err != nil {
		log.Fatal(err)
	}

	repository, err := storageRepository(ctx, pool)
	if err != nil {
		log.Fatal(err)
	}
	storageService := storage.NewService(repository)
	mediaObjectRepository, err := mediaObjectRepository(ctx, pool)
	if err != nil {
		log.Fatal(err)
	}
	mediaObjectService := storage.NewMediaObjectService(repository, mediaObjectRepository)

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
		Addr: address,
		Handler: httpapi.NewHandler(
			storageService,
			httpapi.WithAdminToken(adminToken),
			httpapi.WithMediaObjectService(mediaObjectService),
			httpapi.WithServiceInfo(httpapi.ServiceInfo{Name: "inori-api", Version: version, Commit: commit, BuildTime: buildTime}),
		).Routes(),
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

// openDatabasePool opens a pgxpool if INORI_DATABASE_URL is set, or returns nil.
func openDatabasePool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("INORI_DATABASE_URL")
	if dsn == "" {
		return nil, nil
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}
	defer conn.Release()
	if err := pgstore.Migrate(ctx, conn.Conn()); err != nil {
		pool.Close()
		return nil, err
	}
	log.Print("storage: using PostgreSQL persistence")
	return pool, nil
}

func storageRepository(ctx context.Context, pool *pgxpool.Pool) (storage.Repository, error) {
	if pool != nil {
		return pgstore.NewBackendRepository(pool), nil
	}
	path := os.Getenv("INORI_STORAGE_REPOSITORY_FILE")
	if path == "" {
		return storage.NewMemoryRepository(), nil
	}
	log.Printf("storage repository file enabled at %s", path)
	return storage.NewFileRepository(path)
}

func mediaObjectRepository(ctx context.Context, pool *pgxpool.Pool) (storage.MediaObjectRepository, error) {
	if pool != nil {
		return pgstore.NewMediaObjectRepository(pool), nil
	}
	path := os.Getenv("INORI_MEDIA_OBJECT_REPOSITORY_FILE")
	if path == "" {
		return storage.NewMemoryMediaObjectRepository(), nil
	}
	log.Printf("media object repository file enabled at %s", path)
	return storage.NewFileMediaObjectRepository(path)
}
