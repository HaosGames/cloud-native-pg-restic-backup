package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloud-native-pg-restic-backup/internal/restic"
)

func main() {
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

	client := restic.NewClient(config)

	// Initialize repository
	logger.Println("Initializing repository...")
	if err := client.InitRepository(ctx); err != nil {
		logger.Printf("Failed to initialize repository: %v\n", err)
		os.Exit(1)
	}

	logger.Println("CloudNativePG Restic backup plugin initialized and ready")

	// Wait for context cancellation
	<-ctx.Done()
	logger.Println("Shutting down...")
}
