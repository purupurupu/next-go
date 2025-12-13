package constants

import "time"

// Validation limits
const (
	MinPasswordLength = 6
	MinNameLength     = 2
	MaxNameLength     = 50
	MaxTitleLength    = 255
	MaxDescLength     = 10000
)

// Priority values
const (
	PriorityLow    = 0
	PriorityMedium = 1
	PriorityHigh   = 2
)

// Status values
const (
	StatusPending    = 0
	StatusInProgress = 1
	StatusCompleted  = 2
)

// Time formats
const (
	DateFormat     = "2006-01-02"
	DateTimeFormat = "2006-01-02T15:04:05Z"
)

// Default values
const (
	DefaultJWTExpirationHours = 24
	DefaultCORSMaxAge         = 86400 // 24 hours in seconds
	DefaultServerPort         = "3000"
	GracefulShutdownTimeout   = 10 * time.Second
)

// Environment names
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)
