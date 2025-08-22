// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gokeki/config"
	"gokeki/routes"
	"gokeki/services/cache"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  Warning: Could not load .env file: %v", err)
		log.Println("ℹ️  Continuing with system environment variables...")
	} else {
		log.Println("✅ .env file loaded successfully")
	}

	// Load config
	cfg := config.LoadConfig()

	// Initialize Redis
	cache.InitRedis(cfg)

	// Setup Gin router
	r := gin.Default()

	// Include routes
	routes.SetupRoutes(r)

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the 7TV Emote API",
			"endpoints": gin.H{
				"search":           "/api/search-emotes",
				"trending_emotes":  "/api/trending/emotes",
				"storage_trending": "/api/storage/trending-emotes",
				"storage_emotes":   "/api/storage/emote-api",
				"cache_status":     "/api/cache/status",
				"clear_cache":      "/api/cache/clear",
				"health":           "/health",
			},
			"documentation": "/docs",
		})
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		redisStatus := "connected"
		if err := cache.RedisClient.Ping(context.Background()).Err(); err != nil {
			redisStatus = "disconnected"
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"redis":     redisStatus,
		})
	})

	// Process time middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		processTime := time.Since(start).Seconds()
		c.Header("X-Process-Time", fmt.Sprintf("%f", processTime))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
