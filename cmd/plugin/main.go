package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"cloud-native-pg-restic-backup/internal/logging"
	"cloud-native-pg-restic-backup/internal/plugin"
	"cloud-native-pg-restic-backup/internal/restic"
)

var (
	listenAddr = flag.String("listen", ":8080", "HTTP server listen address")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	logJSON    = flag.Bool("log-json", false, "Output logs in JSON format")
)

func main() {
	flag.Parse()

	// Initialize logger
	logger := logging.NewLogger(logging.Config{
		Level:      *logLevel,
		JSONOutput: *logJSON,
	})

	mainLogger := logger.Component("main")
	mainLogger.Info().
		Str("version", "1.0.0").
		Str("listen_addr", *listenAddr).
		Msg("Starting CloudNativePG Restic backup plugin")

	// Create context that will be cancelled on SIGINT/SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		mainLogger.Info().
			Str("signal", sig.String()).
			Msg("Received termination signal")
		cancel()
	}()

	// Initialize restic client with configuration from environment
	config := restic.Config{
		Repository:  os.Getenv("RESTIC_REPOSITORY"),
		Password:    os.Getenv("RESTIC_PASSWORD"),
		S3Endpoint:  os.Getenv("S3_ENDPOINT"),
		S3AccessKey: os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey: os.Getenv("S3_SECRET_KEY"),
	}

	// Validate required environment variables
	if config.Repository == "" {
		mainLogger.Fatal().Msg("RESTIC_REPOSITORY environment variable is required")
	}
	if config.Password == "" {
		mainLogger.Fatal().Msg("RESTIC_PASSWORD environment variable is required")
	}

	// Create and initialize plugin
	p := plugin.NewPlugin(config, logger.Component("plugin"))

	// Create HTTP server
	server := &http.Server{
		Addr:    *listenAddr,
		Handler: p,
	}

	// Initialize repository
	mainLogger.Info().Msg("Initializing repository...")
	client := restic.NewClient(config)
	if err := client.InitRepository(ctx); err != nil {
		mainLogger.Fatal().Err(err).Msg("Failed to initialize repository")
	}

	// Start HTTP server
	mainLogger.Info().
		Str("addr", server.Addr).
		Msg("Starting HTTP server")

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			mainLogger.Error().Err(err).Msg("HTTP server error")
			cancel()
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	mainLogger.Info().Msg("Shutting down HTTP server...")
	if err := server.Shutdown(context.Background()); err != nil {
		mainLogger.Error().Err(err).Msg("Error during server shutdown")
	}

	mainLogger.Info().Msg("Server shutdown complete")
}
