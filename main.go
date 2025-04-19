package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/erik-schuetze/hetzner-ddns/internal/config"
	"github.com/erik-schuetze/hetzner-ddns/internal/controller"
)

func main() {
	// set configPath and initialize config
	configPath := "/config/config.yaml"
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown on SIGTERM or SIGINT
	// This will allow the controller to finish its current work before exiting
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		cancel()
	}()

	// Initialize the controller and start it
	c := controller.NewController(cfg, configPath)
	if err := c.Run(ctx); err != nil {
		log.Fatalf("Error running controller: %v", err)
	}

}
