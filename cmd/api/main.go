package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/year-recap/internal/config"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/server"
	"github.com/year-recap/internal/storage/memory"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	configured, err := config.FromEnv()
	if err != nil {
		return err
	}
	store, err := memory.Load(configured.ProfilesPath, configured.ScenariosPath)
	if err != nil {
		return err
	}
	service, err := application.NewService(store, store, store, store)
	if err != nil {
		return err
	}
	handler, err := server.NewHandler(service, server.Options{
		StaticDir:      configured.StaticDir,
		AllowedOrigins: configured.AllowedOrigins,
	})
	if err != nil {
		return err
	}
	httpServer := &http.Server{
		Addr:              configured.Address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("recap API listening on %s", configured.Address)
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), configured.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}
	err = <-serverErrors
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
