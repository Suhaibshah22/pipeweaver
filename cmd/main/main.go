package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Suhaibshah22/pipeweaver/cmd"
	"github.com/Suhaibshah22/pipeweaver/cmd/config"
)

func main() {
	// Load Configurations
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load application config: %v", err)
	}

	// Initialize Context
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize DI container
	container := cmd.InitializeContainer(cfg, ctx)

	// Initialize Signal Handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize HTTP Router
	router := cmd.SetupRouter(container)

	//Start Server in new goroutine
	go func() {
		address := ":" + cfg.Server.Port
		if err := router.Run(address); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-sigChan
	container.Logger.Info("Received termination signal. Shutting down gracefully...")
	cancel()

}
