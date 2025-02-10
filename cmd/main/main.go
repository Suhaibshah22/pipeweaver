package main

import (
	"log"

	"github.com/Suhaibshah22/pipeweaver/cmd"
	"github.com/Suhaibshah22/pipeweaver/cmd/config"
)

func main() {
	// Load Configurations
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load application config: %v", err)
	}

	// Initialize DI container
	container := cmd.InitializeContainer(cfg)

	// Initialize HTTP Router
	router := cmd.SetupRouter(container)

	//Start Server
	address := ":" + cfg.Server.Port
	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
