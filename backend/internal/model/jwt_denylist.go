package model

import (
	"time"
)

// JwtDenylist represents a revoked JWT token
type JwtDenylist struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Jti       string    `gorm:"index;not null;size:255" json:"jti"`
	Exp       time.Time `json:"exp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the JwtDenylist model
func (JwtDenylist) TableName() string {
	return "jwt_denylists"
}
