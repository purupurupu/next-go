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

	// S3 storage settings
	S3Endpoint     string `envconfig:"S3_ENDPOINT" default:"http://localhost:9000"`
	S3Region       string `envconfig:"S3_REGION" default:"us-east-1"`
	S3Bucket       string `envconfig:"S3_BUCKET" default:"todo-files"`
	S3AccessKey    string `envconfig:"S3_ACCESS_KEY" default:"rustfs-dev-access"`
	S3SecretKey    string `envconfig:"S3_SECRET_KEY" default:"rustfs-dev-secret-key"`
	S3UsePathStyle bool   `envconfig:"S3_USE_PATH_STYLE" default:"true"`
}

// S3Config holds S3 storage configuration
type S3Config struct {
	Endpoint     string
	Region       string
	Bucket       string
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
}

// GetS3Config returns S3 configuration
func (c *Config) GetS3Config() *S3Config {
	return &S3Config{
		Endpoint:     c.S3Endpoint,
		Region:       c.S3Region,
		Bucket:       c.S3Bucket,
		AccessKey:    c.S3AccessKey,
		SecretKey:    c.S3SecretKey,
		UsePathStyle: c.S3UsePathStyle,
	}
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
