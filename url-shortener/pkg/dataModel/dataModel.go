package dataModel

import (
	"log"
	"net/url" 
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// URLMapping represents the mapping between short URL ID and long URL.
type URLMapping struct {
	gorm.Model
	ShortURLID string `gorm:"uniqueIndex"`
	LongURL    string `gorm:"uniqueIndex"`
	DomainName string `gorm:"index"` // Added for metrics
}

// User represents a user in the system.
type User struct {
	gorm.Model
	Email     string    `gorm:"uniqueIndex;not null"`
	FirstName string    `gorm:"not null"`
	LastName  string    `gorm:"not null"`
	APIKey    uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
}

// DomainCount holds the domain name and its count.
type DomainCount struct {
	DomainName string
	Count      int64
}

// DataAccessLayer defines the interface for accessing data.
type DataAccessLayer interface {
	CreateURLMapping(mapping *URLMapping) error
	GetLongURL(shortURLID string) (string, error)
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	GetAPIKeyByEmail(email string) (string, error)
	CheckAPIKey(apiKey string) (bool, error)
	GetTopDomains(limit int) ([]DomainCount, error) // Added for metrics
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
	// Parse domain from LongURL
	parsedURL, err := url.Parse(mapping.LongURL)
	if err != nil {
		log.Printf("Error parsing LongURL %s: %v", mapping.LongURL, err)
		return fmt.Errorf("invalid Url as parsing failed")
	} else {
		mapping.DomainName = parsedURL.Hostname()
		// Ensure DomainName is "" if Hostname() returns empty (e.g. for file URLs)
		if mapping.DomainName == "" {
			log.Printf("Parsed URL %s resulted in an empty hostname.", mapping.LongURL)
		}
	}
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
	if err := db.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("failed to create uuid-ossp extension: %v", err)
		return err
	}
	return db.DB.AutoMigrate(dst...)
}

// GetTopDomains retrieves the top N domains with the most shortened URLs.
func (db *DB) GetTopDomains(limit int) ([]DomainCount, error) {
	var results []DomainCount
	err := db.Model(&URLMapping{}).
		Select("domain_name, count(*) as count").
		Where("domain_name IS NOT NULL AND domain_name != ''"). // Ensure domain_name is not empty or null
		Group("domain_name").
		Order("count desc").
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
