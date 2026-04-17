package repository

import (
	"context"
	"errors"

	"github.com/hariharandr/config-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConfigRepository defines the interface for config persistence.
// We define an interface here so the service layer depends on the
// interface, not the concrete struct. This makes testing easy —
// you can swap in a mock without touching a real database.
type ConfigRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Config, error)
	Upsert(ctx context.Context, req *domain.UpsertRequest) (*domain.Config, error)
}

// postgresRepo is the real Postgres implementation of ConfigRepository.
// It is unexported — callers get it via New(), not directly.
type postgresRepo struct {
	pool *pgxpool.Pool
}

// New creates a new postgresRepo.
// We accept a *pgxpool.Pool instead of a connection string because:
// 1. The pool is already configured and connected by the time we get here
// 2. This function stays focused — it doesn't do connection logic
func New(pool *pgxpool.Pool) ConfigRepository {
	return &postgresRepo{pool: pool}
}

// GetByID fetches a single config by its ID.
// Returns domain.ErrNotFound if the row does not exist.
func (r *postgresRepo) GetByID(ctx context.Context, id string) (*domain.Config, error) {
	query := `
		SELECT id, host, port, app_name, log_level, created_at, updated_at
		FROM configs
		WHERE id = $1
	`

	var cfg domain.Config

	// pgx.ErrNoRows is what pgx returns when SELECT finds nothing.
	// We translate it into our own domain.ErrNotFound so upper layers
	// never need to import pgx just to check "was it a not found error?"
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&cfg.ID,
		&cfg.Host,
		&cfg.Port,
		&cfg.AppName,
		&cfg.LogLevel,
		&cfg.CreatedAt,
		&cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &cfg, nil
}

// Upsert inserts a config or updates it if the ID already exists.
// ON CONFLICT (id) DO UPDATE means:
//   if a row with this id exists → update host, port, app_name, log_level, updated_at
//   if it does not exist         → insert a new row
//
// RETURNING gives us back the full row after the write,
// so we don't need a second SELECT query.
func (r *postgresRepo) Upsert(ctx context.Context, req *domain.UpsertRequest) (*domain.Config, error) {
	query := `
		INSERT INTO configs (id, host, port, app_name, log_level)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			host       = EXCLUDED.host,
			port       = EXCLUDED.port,
			app_name   = EXCLUDED.app_name,
			log_level  = EXCLUDED.log_level,
			updated_at = NOW()
		RETURNING id, host, port, app_name, log_level, created_at, updated_at
	`

	var cfg domain.Config

	err := r.pool.QueryRow(ctx, query,
		req.ID,
		req.Host,
		req.Port,
		req.AppName,
		req.LogLevel,
	).Scan(
		&cfg.ID,
		&cfg.Host,
		&cfg.Port,
		&cfg.AppName,
		&cfg.LogLevel,
		&cfg.CreatedAt,
		&cfg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
