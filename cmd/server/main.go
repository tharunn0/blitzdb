package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/tharunn0/blitzdb/internal/config"
	"github.com/tharunn0/blitzdb/internal/http/handlers"
	"github.com/tharunn0/blitzdb/internal/service"
)

func main() {

	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	srv := service.NewService(cfg)
	h := handlers.NewHandlers(srv)

	app := fiber.New(fiber.Config{
		AppName:      "Blitz",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    10 * 1024 * 1024, // 10MB max request size
	})

	// Setup routes
	api := app.Group("/api/v1")
	api.Post("/set", h.SetHandler)
	api.Get("/get/:key", h.GetHandler)
	api.Delete("/del/:key", h.DelHandler)
	api.Get("/metrics", h.MetricsHandler)
	app.Get("/health", h.HealthHandler)

	ctx, cancel := context.WithCancel(context.Background())
	go srv.Janitor(ctx)

	serverErr := make(chan error, 1)
	go func() {
		log.Println("Server starting on :8080")
		if err := app.Listen(":1100"); err != nil {
			serverErr <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		log.Println("Shutdown signal received")
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
	}

	log.Println("Shutting down gracefully...")

	cancel()

	srv.Stop()

	if err := app.Shutdown(); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
