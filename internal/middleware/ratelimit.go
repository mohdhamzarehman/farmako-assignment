package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redisClient *redis.Client
	limit       int
	window      time.Duration
}

func NewRateLimiter(redisClient *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		limit:       limit,
		window:      window,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "rate_limit:" + c.ClientIP()
		ctx := c.Request.Context()

		count, err := rl.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
			c.Abort()
			return
		}

		if count >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": rl.window.Seconds(),
			})
			c.Abort()
			return
		}

		pipe := rl.redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, rl.window)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit update failed"})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(rl.limit-count-1))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.window).Unix(), 10))

		c.Next()
	}
}
