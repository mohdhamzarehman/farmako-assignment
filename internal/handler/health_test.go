package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHealthHandler(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Setup Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Create health handler
	handler := NewHealthHandler(db, redisClient)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	t.Run("should return healthy when all services are up", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "healthy", response.Status)
		assert.Equal(t, "healthy", response.Services["database"])
		assert.Equal(t, "healthy", response.Services["redis"])
		assert.WithinDuration(t, time.Now(), response.Timestamp, time.Second)
	})

	t.Run("should return unhealthy when database is down", func(t *testing.T) {
		// Create a handler with a closed database
		closedDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		sqlDB, _ := closedDB.DB()
		sqlDB.Close()

		handler := NewHealthHandler(closedDB, redisClient)
		router := gin.New()
		router.GET("/health", handler.HealthCheck)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "unhealthy", response.Status)
		assert.Contains(t, response.Services["database"], "unhealthy")
	})

	t.Run("should return unhealthy when redis is down", func(t *testing.T) {
		// Create a handler with a closed Redis client
		closedRedis := redis.NewClient(&redis.Options{
			Addr: "localhost:9999", // Non-existent port
		})

		handler := NewHealthHandler(db, closedRedis)
		router := gin.New()
		router.GET("/health", handler.HealthCheck)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "unhealthy", response.Status)
		assert.Contains(t, response.Services["redis"], "unhealthy")
	})
}
