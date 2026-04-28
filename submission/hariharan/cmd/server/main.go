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
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect with retry
	pool := connectWithRetry(dbURL)
	defer pool.Close()

	// Run migrations automatically on startup
	// This is safe to run multiple times — uses CREATE TABLE IF NOT EXISTS
	runMigrations(pool)

	// Wire layers
	repo := repository.New(pool)
	svc := service.New(repo)
	h := handler.New(svc)

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

func runMigrations(pool *pgxpool.Pool) {
	log.Println("running migrations...")

	query := `
		CREATE TABLE IF NOT EXISTS configs (
			id         TEXT PRIMARY KEY,
			host       TEXT NOT NULL,
			port       INT  NOT NULL,
			app_name   TEXT NOT NULL,
			log_level  TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`

	_, err := pool.Exec(context.Background(), query)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migrations complete")
}

func connectWithRetry(dbURL string) *pgxpool.Pool {
	var pool *pgxpool.Pool
	var err error

	maxAttempts := 10

	for i := 1; i <= maxAttempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pool, err = pgxpool.New(ctx, dbURL)
		cancel()

		if err == nil {
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
