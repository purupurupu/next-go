package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// BcryptCost is the cost parameter for bcrypt hashing (Rails compatible)
const BcryptCost = 12

// User represents a user in the system
type User struct {
	ID                int64     `gorm:"primaryKey" json:"id"`
	Email             string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	EncryptedPassword string    `gorm:"column:encrypted_password;not null" json:"-"`
	Name              *string   `gorm:"size:255" json:"name"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// SetPassword hashes the password using bcrypt and stores it
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return err
	}
	u.EncryptedPassword = string(hash)
	return nil
}

// CheckPassword verifies if the provided password matches the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password))
	return err == nil
}
