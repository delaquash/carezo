package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/delaquash/carezo/carezo_backends/configs"
	"github.com/delaquash/carezo/configs"
	"github.com/redis/go-redis/v9"
)

// global redis connection
var RedisClient *redis.Client


// ConnectRedis establishes connection to redis
func ConnectRedis (cfg *configs.Config) (*redis.Client, error){
	// Create redis client with configuration
	client := redis.NewClient(&redis.Options{
		Addr:	fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB: 0,
	})

	// test connection with a ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx), Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
}
