package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache interface defines methods for caching
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Close() error
}

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	prefix string
}

// RedisCacheConfig holds configuration for Redis cache
type RedisCacheConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	Prefix   string
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(cfg RedisCacheConfig) (*RedisCache, error) {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 12429
	}
	if cfg.Prefix == "" {
		cfg.Prefix = "appname:"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		prefix: cfg.Prefix,
	}, nil
}

// buildKey creates a Redis key with the configured prefix
func (c *RedisCache) buildKey(key string) string {
	return c.prefix + key
}

// Get retrieves a value from Redis
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := c.client.Get(ctx, c.buildKey(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("cache miss")
		}
		return nil, fmt.Errorf("getting from Redis: %w", err)
	}
	return data, nil
}

// Set stores a value in Redis with an expiration
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := c.client.Set(ctx, c.buildKey(key), value, ttl).Err()
	if err != nil {
		return fmt.Errorf("setting in Redis: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
