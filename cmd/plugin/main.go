package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"cloud-native-pg-restic-backup/internal/plugin"
	"cloud-native-pg-restic-backup/internal/restic"
)

var (
	listenAddr = flag.String("listen", ":8080", "HTTP server listen address")
)

func main() {
	flag.Parse()
	logger := log.New(os.Stdout, "restic-plugin: ", log.LstdFlags)

	// Create context that will be cancelled on SIGINT/SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Println("Received termination signal")
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
		logger.Fatal("RESTIC_REPOSITORY environment variable is required")
	}
	if config.Password == "" {
		logger.Fatal("RESTIC_PASSWORD environment variable is required")
	}

	// Create and initialize plugin
	p := plugin.NewPlugin(config, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:    *listenAddr,
		Handler: p,
	}

	// Initialize repository
	logger.Println("Initializing repository...")
	client := restic.NewClient(config)
	if err := client.InitRepository(ctx); err != nil {
		logger.Printf("Failed to initialize repository: %v\n", err)
		os.Exit(1)
	}

	// Start HTTP server
	logger.Printf("Starting HTTP server on %s", *listenAddr)
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Printf("HTTP server error: %v", err)
			cancel()
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	logger.Println("Shutting down HTTP server...")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Printf("Error during server shutdown: %v", err)
	}
}
