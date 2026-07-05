package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/audioanalysis"
	"inori-music/services/api/internal/auth"
	authpg "inori-music/services/api/internal/auth/postgres"
	"inori-music/services/api/internal/catalog"
	catalogpg "inori-music/services/api/internal/catalog/postgres"
	"inori-music/services/api/internal/favorites"
	favoritespg "inori-music/services/api/internal/favorites/postgres"
	"inori-music/services/api/internal/history"
	historypg "inori-music/services/api/internal/history/postgres"
	"inori-music/services/api/internal/httpapi"
	"inori-music/services/api/internal/search"
	"inori-music/services/api/internal/storage"
	pgstore "inori-music/services/api/internal/storage/postgres"
	"inori-music/services/api/internal/userplaylist"
	userplaylistpg "inori-music/services/api/internal/userplaylist/postgres"
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

	// Auth service — only available when PostgreSQL is configured.
	var authService *auth.Service
	if pool != nil {
		authService = auth.NewService(
			authpg.NewUserRepository(pool),
			authpg.NewSessionRepository(pool),
			auth.ServiceConfig{SessionTTL: sessionTTL()},
		)
		if err := authService.EnsureInitialAdmin(
			ctx,
			os.Getenv("INORI_INITIAL_ADMIN_USER"),
			os.Getenv("INORI_INITIAL_ADMIN_PASSWORD"),
		); err != nil {
			log.Printf("initial admin setup: %v", err)
		}
	}

	// Catalog service — PostgreSQL when pool is available, in-memory otherwise.
	catalogRepo := catalogRepository(pool)
	catalogService := catalog.NewService(catalogRepo)
	catalogService.WithAudioAnalyzer(audioanalysis.New(mediaObjectService, storageService, catalogService))

	// Meilisearch search — optional, falls back to PG if not configured.
	var searchSvc search.Service
	if meiliHost := os.Getenv("MEILI_HOST"); meiliHost != "" {
		ms, err := search.NewMeilisearch(meiliHost, os.Getenv("MEILI_SEARCH_KEY"))
		if err != nil {
			log.Printf("meilisearch init: %v", err)
		}
		if ms != nil {
			searchSvc = ms
			catalogService.WithSearchService(searchSvc)
			log.Printf("meilisearch enabled at %s", meiliHost)
		}
	}

	// History service — PostgreSQL when pool is available, in-memory otherwise.
	historyRepo := historyRepository(pool)
	historyService := history.NewService(historyRepo)

	// Favorites service — PostgreSQL when pool is available, in-memory otherwise.
	favoritesRepo := favoritesRepository(pool)
	favoritesService := favorites.NewService(favoritesRepo)

	// User playlist service — PostgreSQL when pool is available, in-memory otherwise.
	userPlaylistRepo := userPlaylistRepository(ctx, pool)
	userPlaylistService := userplaylist.NewService(userPlaylistRepo)

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

	handlerOpts := []httpapi.HandlerOption{
		httpapi.WithAdminToken(adminToken),
		httpapi.WithMediaObjectService(mediaObjectService),
		httpapi.WithCatalogService(catalogService),
		httpapi.WithHistoryService(historyService),
		httpapi.WithFavoritesService(favoritesService),
		httpapi.WithUserPlaylistService(userPlaylistService),
		httpapi.WithCORSOrigins(corsOrigins()),
		httpapi.WithServiceInfo(httpapi.ServiceInfo{Name: "inori-api", Version: version, Commit: commit, BuildTime: buildTime}),
	}
	if authService != nil {
		handlerOpts = append(handlerOpts, httpapi.WithAuthService(authService))
	}
	if searchSvc != nil {
		handlerOpts = append(handlerOpts, httpapi.WithSearchService(searchSvc))
	}

	server := &http.Server{
		Addr:              address,
		Handler:           httpapi.NewHandler(storageService, handlerOpts...).Routes(),
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

func sessionTTL() time.Duration {
	value := os.Getenv("INORI_SESSION_TTL")
	if value == "" {
		return 24 * time.Hour
	}
	ttl, err := time.ParseDuration(value)
	if err != nil || ttl <= 0 {
		log.Printf("ignoring invalid INORI_SESSION_TTL %q, using 24h", value)
		return 24 * time.Hour
	}
	return ttl
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

// catalogRepository returns a PostgreSQL-backed catalog repository when a pool is
// available, falling back to an in-memory repository for development and testing.
func catalogRepository(pool *pgxpool.Pool) catalog.Repository {
	if pool != nil {
		return catalogpg.NewRepository(pool)
	}
	return catalog.NewMemoryRepository()
}

// historyRepository returns a PostgreSQL-backed history repository when a pool is
// available, falling back to an in-memory repository for development and testing.
func historyRepository(pool *pgxpool.Pool) history.Repository {
	if pool != nil {
		return historypg.NewRepository(pool)
	}
	return history.NewMemoryRepository()
}

// favoritesRepository returns a PostgreSQL-backed favorites repository when a pool is
// available, falling back to an in-memory repository for development and testing.
func favoritesRepository(pool *pgxpool.Pool) favorites.Repository {
	if pool != nil {
		return favoritespg.NewRepository(pool)
	}
	return favorites.NewMemoryRepository()
}

// userPlaylistRepository returns a PostgreSQL-backed user playlist repository when a pool
// is available, falling back to an in-memory repository for development and testing.
func userPlaylistRepository(ctx context.Context, pool *pgxpool.Pool) userplaylist.Repository {
	if pool != nil {
		if err := userplaylistpg.Migrate(ctx, pool); err != nil {
			log.Printf("user playlist migration: %v", err)
		}
		return userplaylistpg.NewRepository(pool)
	}
	return userplaylist.NewMemoryRepository()
}

// corsOrigins parses INORI_CORS_ORIGINS (comma-separated list of allowed origins).
// Returns an empty slice when the env var is unset, which enables permissive mode
// in the CORS middleware (any origin is reflected — suitable for local development).
func corsOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("INORI_CORS_ORIGINS"))
	if raw == "" {
		log.Print("INORI_CORS_ORIGINS is not set; CORS middleware running in permissive mode (any origin allowed)")
		return nil
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if o := strings.TrimSpace(p); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}
