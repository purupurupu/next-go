package config

import (
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	Port string `envconfig:"PORT" default:"3000"`

	// Database settings
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// JWT settings
	JWTSecret          string `envconfig:"JWT_SECRET" required:"true"`
	JWTExpirationHours int    `envconfig:"JWT_EXPIRATION_HOURS" default:"24"`

	// Environment
	Env string `envconfig:"ENV" default:"development"`

	// CORS settings
	CORSAllowOrigins string `envconfig:"CORS_ALLOW_ORIGINS" default:"http://localhost:3000"`
	CORSMaxAge       int    `envconfig:"CORS_MAX_AGE" default:"86400"`
}

// GetCORSOrigins returns the CORS allowed origins as a slice
func (c *Config) GetCORSOrigins() []string {
	if c.CORSAllowOrigins == "*" {
		return []string{"*"}
	}
	// Split by comma for multiple origins
	origins := []string{}
	for _, origin := range splitAndTrim(c.CORSAllowOrigins, ",") {
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

// splitAndTrim splits a string by separator and trims whitespace
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}
