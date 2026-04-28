package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hariharandr/config-service/internal/domain"
)

// fakeService is a mock that implements service.ConfigService
// We use this so tests never touch a real database
type fakeService struct {
	configs map[string]*domain.Config
}

func newFakeService() *fakeService {
	return &fakeService{configs: make(map[string]*domain.Config)}
}

func (f *fakeService) GetConfig(ctx context.Context, id string) (*domain.Config, error) {
	cfg, ok := f.configs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cfg, nil
}

func (f *fakeService) UpsertConfig(ctx context.Context, req *domain.UpsertRequest) (*domain.Config, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	cfg := &domain.Config{
		ID:        req.ID,
		Host:      req.Host,
		Port:      req.Port,
		AppName:   req.AppName,
		LogLevel:  req.LogLevel,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	f.configs[req.ID] = cfg
	return cfg, nil
}

// TestPing checks the health endpoint returns pong
func TestPing(t *testing.T) {
	svc := newFakeService()
	h := New(svc)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 got %d", rec.Code)
	}
	if rec.Body.String() != "pong" {
		t.Errorf("expected pong got %s", rec.Body.String())
	}
}

// TestUpsertAndGetConfig checks create then retrieve works
func TestUpsertAndGetConfig(t *testing.T) {
	svc := newFakeService()
	h := New(svc)
	router := h.Router()

	// Create a config
	body := `{"id":"cfg_test","host":"localhost","port":8080,"app_name":"test","log_level":"INFO"}`
	req := httptest.NewRequest(http.MethodPost, "/configs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("upsert: expected 200 got %d — body: %s", rec.Code, rec.Body.String())
	}

	// Retrieve it
	req = httptest.NewRequest(http.MethodGet, "/configs/cfg_test", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("get: expected 200 got %d", rec.Code)
	}

	var cfg domain.Config
	if err := json.NewDecoder(rec.Body).Decode(&cfg); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if cfg.ID != "cfg_test" {
		t.Errorf("expected id cfg_test got %s", cfg.ID)
	}
}

// TestGetConfigNotFound checks 404 for missing id
func TestGetConfigNotFound(t *testing.T) {
	svc := newFakeService()
	h := New(svc)

	req := httptest.NewRequest(http.MethodGet, "/configs/doesnotexist", nil)
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 got %d", rec.Code)
	}
}

// TestUpsertValidation checks bad input returns 400
func TestUpsertValidation(t *testing.T) {
	svc := newFakeService()
	h := New(svc)

	// Missing required fields
	body := `{"id":"","host":"","port":0,"app_name":"","log_level":""}`
	req := httptest.NewRequest(http.MethodPost, "/configs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 got %d", rec.Code)
	}
}
