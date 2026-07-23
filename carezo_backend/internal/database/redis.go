package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/delaquash/carezo/configs"
	"github.com/redis/go-redis/v9"
)

// global redis connection
var RedisClient *redis.Client


// ConnectRedis establishes connection to redis
func ConnectRedis(cfg *configs.Config) (*redis.Client, error) {
	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	}

	// Upstash (and most managed Redis providers) require TLS on every
	// connection — local Docker Redis doesn
	
	// two separate code paths someone has to remember to keep in sync.
	if cfg.AppEnv == "production" {
		options.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Println("Connected to redis successfully")

	RedisClient = client

	return client, nil
}

// Close Redis connection
func CloseRedis() error {
	if RedisClient != nil {
		log.Println("Closing Redis connection")
		return RedisClient.Close()
	}
	return nil
}
