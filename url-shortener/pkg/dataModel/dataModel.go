package dataModel

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// URLMapping represents the mapping between short URL ID and long URL.
type URLMapping struct {
	gorm.Model
	ShortURLID string `gorm:"uniqueIndex"`
	LongURL    string
}

// User represents a user in the system.
type User struct {
	gorm.Model
	Email     string    `gorm:"uniqueIndex"`
	FirstName string
	LastName  string
	APIKey    uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
}