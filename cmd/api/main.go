package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/year-recap/gen/go/recap/v1/recapv1connect"
	"github.com/year-recap/internal/bootstrap"
	"github.com/year-recap/internal/config"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/storage/clickhouse"
	transportconnect "github.com/year-recap/internal/transport/connect"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	connectCtx, cancel := context.WithTimeout(rootCtx, 15*time.Second)
	repo, err := clickhouse.Connect(connectCtx, cfg.ClickHouseDSN)
	cancel()
	if err != nil {
		return fmt.Errorf("connect storage: %w", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("close clickhouse: %v", err)
		}
	}()

	schemaCtx, cancel := context.WithTimeout(rootCtx, 30*time.Second)
	err = repo.EnsureSchema(schemaCtx)
	cancel()
	if err != nil {
		return fmt.Errorf("ensure clickhouse schema: %w", err)
	}

	if cfg.SeedDemoData {
		seedCtx, cancel := context.WithTimeout(rootCtx, 30*time.Second)
		err = bootstrap.LoadDemoData(seedCtx, repo, cfg.ProfilesPath, cfg.ScenariosPath)
		cancel()
		if err != nil {
			return fmt.Errorf("bootstrap demo data: %w", err)
		}
	}

	app, err := application.NewService(repo, repo, repo, repo)
	if err != nil {
		return fmt.Errorf("build application: %w", err)
	}
	rpc, err := transportconnect.NewHandler(app)
	if err != nil {
		return fmt.Errorf("build connect handler: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	connectPath, connectHandler := recapv1connect.NewRecapServiceHandler(rpc)
	mux.Handle(connectPath, cors(cfg.AllowedOrigins, connectHandler))

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("starting api server on %s; connect path %s", cfg.HTTPAddr, connectPath)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-rootCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		return nil
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve api: %w", err)
		}
		return nil
	}
}

func cors(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
