package base

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// RedisConfig stores the configuration for the Redis connection.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisClient creates a new Redis client.
func NewRedisClient(config RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test the connection
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
		return nil, err
	}

	fmt.Println("Connected to Redis!")
	return rdb, nil
}
