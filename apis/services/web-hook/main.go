package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/arjun-armada/skydio-webhook/apis/services/web-hook/config"
	"github.com/arjun-armada/skydio-webhook/apis/services/web-hook/mux"
	"github.com/arjun-armada/skydio-webhook/foundation/logger"
)

func main() {

	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT *******")
		},
	}

	traceIDFn := func(ctx context.Context) string {
		return "" //otel.GetTraceID(ctx)
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "skydio-webhook", traceIDFn, events)

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "msg", err)
		os.Exit(1)
	}

}

func run(ctx context.Context, log *logger.Logger) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	appConfig := config.LoadAppConfig("./")

	// Start API Service
	log.Info(ctx, "startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	api := http.Server{
		Addr:         appConfig.Web.APIHost,
		Handler:      mux.WebAPI(),
		ReadTimeout:  appConfig.Web.ReadTimeout,
		WriteTimeout: appConfig.Web.WriteTimeout,
		IdleTimeout:  appConfig.Web.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil

}
