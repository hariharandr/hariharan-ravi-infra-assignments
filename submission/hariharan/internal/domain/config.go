package domain

import "time"

// Config represents a configuration record stored in the database.
// This is the single source of truth for what a config looks like
// across all layers — handler, service, repository.
type Config struct {
	ID        string    `json:"id"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	AppName   string    `json:"app_name"`
	LogLevel  string    `json:"log_level"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpsertRequest is what the HTTP handler receives from the client.
// We keep it separate from Config because the client should NOT
// be able to set created_at/updated_at — those are server-controlled.
type UpsertRequest struct {
	ID       string `json:"id"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	AppName  string `json:"app_name"`
	LogLevel string `json:"log_level"`
}

// Validate checks that required fields are present.
// We do validation here in domain, not in handler,
// so validation logic is reusable across any entry point.
func (r *UpsertRequest) Validate() error {
	if r.ID == "" {
		return ErrMissingField("id")
	}
	if r.Host == "" {
		return ErrMissingField("host")
	}
	if r.Port <= 0 || r.Port > 65535 {
		return ErrInvalidField("port", "must be between 1 and 65535")
	}
	if r.AppName == "" {
		return ErrMissingField("app_name")
	}
	if r.LogLevel == "" {
		return ErrMissingField("log_level")
	}
	return nil
}
