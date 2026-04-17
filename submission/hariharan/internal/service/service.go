package service

import (
	"context"

	"github.com/hariharandr/config-service/internal/domain"
	"github.com/hariharandr/config-service/internal/repository"
)

// ConfigService defines what operations the handler can call.
// Again an interface — handler depends on this interface,
// not on the concrete struct. Easy to mock in tests.
type ConfigService interface {
	GetConfig(ctx context.Context, id string) (*domain.Config, error)
	UpsertConfig(ctx context.Context, req *domain.UpsertRequest) (*domain.Config, error)
}

// configService is the real implementation.
type configService struct {
	repo repository.ConfigRepository
}

// New creates a new configService.
// We inject the repository here — this is dependency injection.
// The service doesn't create its own repo, it receives one.
// This keeps construction logic out of business logic.
func New(repo repository.ConfigRepository) ConfigService {
	return &configService{repo: repo}
}

// GetConfig retrieves a config by ID.
// Validation of the id format could go here if needed.
// For now we delegate directly to the repository.
func (s *configService) GetConfig(ctx context.Context, id string) (*domain.Config, error) {
	return s.repo.GetByID(ctx, id)
}

// UpsertConfig validates the request then delegates to the repository.
// Validation lives here and in domain — NOT in the handler.
// Handler only deals with HTTP concerns (parsing, writing response).
func (s *configService) UpsertConfig(ctx context.Context, req *domain.UpsertRequest) (*domain.Config, error) {
	// Validate before hitting the database.
	// If validation fails, we never make a DB call.
	if err := req.Validate(); err != nil {
		return nil, err
	}

	return s.repo.Upsert(ctx, req)
}
