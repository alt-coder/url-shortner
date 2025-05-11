package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9" 
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock for RedisClientInterface
type MockRedisClient struct {
	mock.Mock
}

var _ RedisClientInterface = (*MockRedisClient)(nil)

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}