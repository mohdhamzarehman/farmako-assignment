package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	redisClient.FlushAll(ctx)
	defer redisClient.FlushAll(ctx)

	limiter := NewRateLimiter(redisClient, 2, time.Second)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(limiter.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("should allow requests within limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "2", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "1", w.Header().Get("X-RateLimit-Remaining"))

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "2", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
	})

	t.Run("should block requests exceeding limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "rate limit exceeded")
	})

	t.Run("should reset after window expires", func(t *testing.T) {
		time.Sleep(time.Second)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "2", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "1", w.Header().Get("X-RateLimit-Remaining"))
	})
}
