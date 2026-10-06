// Command api runs the print quote HTTP service.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/adapter/clock"
	httpadapter "github.com/GabrielPedrosoFerreira/print-quote-service/internal/adapter/http"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/adapter/memory"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/command"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/query"
)

const shutdownTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("service stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Composition root: the only place that knows about concrete adapters.
	quotes := memory.NewQuoteRepository()
	systemClock := clock.System{}
	useCases := httpadapter.UseCases{
		CreateQuote: command.NewCreateQuoteHandler(quotes, systemClock, clock.RandomIDs{}),
		AcceptQuote: command.NewAcceptQuoteHandler(quotes, systemClock),
		GetQuote:    query.NewGetQuoteHandler(quotes, systemClock),
		ListQuotes:  query.NewListQuotesHandler(quotes, systemClock),
	}

	server := &http.Server{
		Addr:              ":" + envOrDefault("PORT", "8080"),
		Handler:           httpadapter.NewRouter(useCases, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
