package repository

import (
	"time"

	"todo-api/internal/model"

	"gorm.io/gorm"
)

// JwtDenylistRepository handles database operations for JWT denylist
type JwtDenylistRepository struct {
	db *gorm.DB
}

// NewJwtDenylistRepository creates a new JwtDenylistRepository
func NewJwtDenylistRepository(db *gorm.DB) *JwtDenylistRepository {
	return &JwtDenylistRepository{db: db}
}

// Add adds a token to the denylist
func (r *JwtDenylistRepository) Add(jti string, exp time.Time) error {
	denylist := &model.JwtDenylist{
		Jti: jti,
		Exp: exp,
	}
	return r.db.Create(denylist).Error
}

// Exists checks if a token with the given jti exists in the denylist
func (r *JwtDenylistRepository) Exists(jti string) (bool, error) {
	var count int64
	result := r.db.Model(&model.JwtDenylist{}).Where("jti = ?", jti).Count(&count)
	return count > 0, result.Error
}

// CleanupExpired removes expired tokens from the denylist
func (r *JwtDenylistRepository) CleanupExpired() error {
	return r.db.Where("exp < ?", time.Now()).Delete(&model.JwtDenylist{}).Error
}
