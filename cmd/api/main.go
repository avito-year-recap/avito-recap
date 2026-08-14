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

	ollamanarrative "github.com/year-recap/internal/ai/ollama"
	"github.com/year-recap/internal/bootstrap"
	"github.com/year-recap/internal/config"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/narrative"
	"github.com/year-recap/internal/server"
	"github.com/year-recap/internal/storage/clickhouse"
	"github.com/year-recap/internal/storage/memory"
)

type applicationStorage interface {
	application.ProfileStorage
	application.AnalyticsStorage
	application.ActionStateStorage
	application.RecapStorage
}

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

	repo, closeStorage, err := openStorage(rootCtx, cfg)
	if err != nil {
		return err
	}
	defer closeStorage()

	aiGenerator, err := buildAIGenerator(rootCtx, cfg)
	if err != nil {
		return fmt.Errorf("build AI generator: %w", err)
	}

	options := make([]application.Option, 0, 3)
	if aiGenerator != nil {
		options = append(options,
			application.WithNarrativeEnricher(buildNarrativeEnricher(aiGenerator, cfg)),
			application.WithInsightGenerator(ollamanarrative.InsightGenerator{Generator: aiGenerator}),
		)
	}
	if eventStorage, ok := repo.(application.EventRangeStorage); ok {
		options = append(options, application.WithEventRangeStorage(eventStorage))
	}

	app, err := application.NewService(repo, repo, repo, repo, options...)
	if err != nil {
		return fmt.Errorf("build application: %w", err)
	}
	handler, err := server.NewHandler(app, server.Options{
		StaticDir:      cfg.StaticDir,
		FrontendDir:    cfg.FrontendDir,
		AllowedOrigins: cfg.AllowedOrigins,
	})
	if err != nil {
		return fmt.Errorf("build http handler: %w", err)
	}

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("starting api server on %s (storage=%s)", cfg.HTTPAddr, cfg.StorageBackend)
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case <-rootCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
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

func openStorage(ctx context.Context, cfg config.Config) (applicationStorage, func(), error) {
	switch cfg.StorageBackend {
	case config.StorageMemory:
		repo, err := memory.Load(cfg.ProfilesPath, cfg.ScenariosPath)
		if err != nil {
			return nil, func() {}, fmt.Errorf("load in-memory demo storage: %w", err)
		}
		return repo, func() {}, nil

	case config.StorageClickHouse:
		connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		repo, err := clickhouse.Connect(connectCtx, cfg.ClickHouseDSN)
		cancel()
		if err != nil {
			return nil, func() {}, fmt.Errorf("connect storage: %w", err)
		}

		closeStorage := func() {
			if err := repo.Close(); err != nil {
				log.Printf("close clickhouse: %v", err)
			}
		}

		schemaCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err = repo.EnsureSchema(schemaCtx)
		cancel()
		if err != nil {
			closeStorage()
			return nil, func() {}, fmt.Errorf("ensure clickhouse schema: %w", err)
		}

		if cfg.SeedDemoData {
			seedCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			err = bootstrap.LoadDemoData(seedCtx, repo, cfg.ProfilesPath, cfg.ScenariosPath)
			cancel()
			if err != nil {
				closeStorage()
				return nil, func() {}, fmt.Errorf("bootstrap demo data: %w", err)
			}
		}

		return repo, closeStorage, nil

	default:
		return nil, func() {}, fmt.Errorf("unsupported storage backend %q", cfg.StorageBackend)
	}
}

func buildAIGenerator(ctx context.Context, cfg config.Config) (*ollamanarrative.Generator, error) {
	provider := cfg.NarrativeProvider
	if provider == config.NarrativeOff {
		return nil, nil
	}

	generator, err := ollamanarrative.New(ollamanarrative.Config{
		Model:     cfg.OllamaModel,
		BaseURL:   cfg.OllamaBaseURL,
		Timeout:   cfg.OllamaTimeout,
		KeepAlive: cfg.OllamaKeepAlive,
	})
	if err != nil {
		return nil, err
	}

	probeTimeout := 2 * time.Second
	if cfg.OllamaTimeout < probeTimeout {
		probeTimeout = cfg.OllamaTimeout
	}
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	probeErr := generator.Check(probeCtx)
	cancel()
	if probeErr != nil {
		if provider == config.NarrativeAuto {
			log.Printf("AI disabled: local Ollama/model unavailable: %v", probeErr)
			return nil, nil
		}
		return nil, fmt.Errorf("local Ollama is not ready: %w", probeErr)
	}

	log.Printf(
		"AI enabled (provider=ollama model=%s base_url=%s max_concurrency=%d)",
		cfg.OllamaModel,
		cfg.OllamaBaseURL,
		cfg.AINarrativeMaxConcurrency,
	)
	return generator, nil
}

func buildNarrativeEnricher(generator *ollamanarrative.Generator, cfg config.Config) narrative.Enricher {
	limited, err := narrative.NewLimited(generator, cfg.AINarrativeMaxConcurrency)
	if err != nil {
		log.Printf("AI narrative disabled: %v", err)
		return nil
	}
	return narrative.BestEffort{
		Primary: limited,
		OnError: func(err error) {
			log.Printf("AI narrative fallback to deterministic copy: %v", err)
		},
	}
}
