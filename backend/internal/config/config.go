package config

import (
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
