package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/database"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

// this helps to limit request er IP address
func RateLimitMiddleware(cfg *configs.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP address
		ip := c.ClientIP()

		// create redis key for the IP
		key := fmt.Sprintf("rate_limit:%s", ip)

		// Increment counter
		ctx := c.Request.Context()
		count, err := database.RedisClient.Incr(ctx, key).Result()

		if err != nil {
			// allow request if redis fail
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			database.RedisClient.Expire(ctx, key, time.Duration(cfg.RateLimitWindowSeconds) *time.Second)
		}

		// check if limit exceeded
		if count > int64(cfg.RateLimitRequests) {
			response.Error(c, http.StatusTooManyRequests, "Rate Limit exceeded!!! Please try again later")
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RateLimitRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", cfg.RateLimitRequests-int(count)))
		c.Next()
	}
}