package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hariharandr/config-service/internal/handler"
	"github.com/hariharandr/config-service/internal/repository"
	"github.com/hariharandr/config-service/internal/service"
)

func main() {
	// --- Config (from env) ---
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default
	}

	// --- DB connection with retry ---
	pool := connectWithRetry(dbURL)
	defer pool.Close()

	// --- Dependency wiring ---
	repo := repository.New(pool)
	svc := service.New(repo)
	h := handler.New(svc)

	// --- HTTP server ---
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      h.Router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("server starting on port %s", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// connectWithRetry tries to connect to DB until success.
// This is CRITICAL for Kubernetes startup ordering.
func connectWithRetry(dbURL string) *pgxpool.Pool {
	var pool *pgxpool.Pool
	var err error

	maxAttempts := 10

	for i := 1; i <= maxAttempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		pool, err = pgxpool.New(ctx, dbURL)
		cancel()

		if err == nil {
			// Verify connection actually works
			pingErr := pool.Ping(context.Background())
			if pingErr == nil {
				log.Println("connected to database")
				return pool
			}

			err = pingErr
		}

		log.Printf("db connection failed (attempt %d/%d): %v", i, maxAttempts, err)
		time.Sleep(2 * time.Second)
	}

	log.Fatal("could not connect to database after retries")
	return nil
}
