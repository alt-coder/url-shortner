package dataModel

import (
	"log"

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
	Email     string    `gorm:"uniqueIndex;not null"`
	FirstName string    `gorm:"not null"`
	LastName  string    `gorm:"not null"`
	APIKey    uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
}

// DataAccessLayer defines the interface for accessing data.
type DataAccessLayer interface {
	CreateURLMapping(mapping *URLMapping) error
	GetLongURL(shortURLID string) (string, error)
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	GetAPIKeyByEmail(email string) (string, error)
	CheckAPIKey(apiKey string) (bool, error)
	AutoMigrate(dst ...interface{}) error
}

// DB represents the database connection.
type DB struct {
	*gorm.DB
}

// NewDB creates a new DB instance.
func NewDB(db *gorm.DB) *DB {
	return &DB{db}
}

// CreateURLMapping creates a new URL mapping in the database.
func (db *DB) CreateURLMapping(mapping *URLMapping) error {
	return db.Create(mapping).Error
}

// GetLongURL retrieves the long URL for a given short URL ID.
func (db *DB) GetLongURL(shortURLID string) (string, error) {
	var mapping URLMapping
	err := db.Where(&URLMapping{ShortURLID: shortURLID}).First(&mapping).Error
	if err != nil {
		return "", err
	}
	return mapping.LongURL, nil
}

// CreateUser creates a new user in the database.
func (db *DB) CreateUser(user *User) error {
	return db.Create(user).Error
}

// GetUserByEmail retrieves a user from the database by email.
func (db *DB) GetUserByEmail(email string) (*User, error) {
	var user User
	err := db.Where(&User{Email: email}).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAPIKeyByEmail retrieves the API key for a given email.
func (db *DB) GetAPIKeyByEmail(email string) (string, error) {
	var user User
	err := db.Where(&User{Email: email}).First(&user).Error
	if err != nil {
		return "", err
	}
	return user.APIKey.String(), nil
}

// CheckAPIKey checks if an API key exists in the database.
func (db *DB) CheckAPIKey(apiKey string) (bool, error) {
	var user User
	api, err := uuid.Parse(apiKey)
	if err != nil {
		log.Printf("api key invalid")
		return false, err
	}
	err = db.Where(&User{APIKey: api}).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (db *DB) AutoMigrate(dst ...interface{}) error {
	return db.DB.AutoMigrate(dst...)
}
