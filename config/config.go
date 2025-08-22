// config/config.go
package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RedisHost        string
	RedisPort        string
	RedisDB          int
	RedisPassword    string
	RedisURL         string
	AzureConnStr     string
	ContainerName    string
	CacheTTL         time.Duration
	TrendingCacheTTL time.Duration
	APITitle         string
	APIDesc          string
	APIVersion       string
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func LoadConfig() *Config {
	db, _ := strconv.Atoi(getEnvWithDefault("REDIS_DB", "0"))
	ttl, _ := strconv.ParseInt(getEnvWithDefault("CACHE_TTL", "3600"), 10, 64)
	trendingTTL, _ := strconv.ParseInt(getEnvWithDefault("TRENDING_CACHE_TTL", "900"), 10, 64)

	// Get Azure connection string with logging
	azureConnStr := os.Getenv("AZURE_CONNECTION_STRING")

	config := &Config{
		RedisHost:        getEnvWithDefault("REDIS_HOST", "localhost"),
		RedisPort:        getEnvWithDefault("REDIS_PORT", "6379"),
		RedisDB:          db,
		RedisPassword:    getEnvWithDefault("REDIS_PASSWORD", ""),
		RedisURL:         getEnvWithDefault("REDIS_URL", ""),
		AzureConnStr:     azureConnStr,
		ContainerName:    getEnvWithDefault("CONTAINER_NAME", "emotes"),
		CacheTTL:         time.Duration(ttl) * time.Second,
		TrendingCacheTTL: time.Duration(trendingTTL) * time.Second,
		APITitle:         getEnvWithDefault("API_TITLE", "7TV Emote API"),
		APIDesc:          getEnvWithDefault("API_DESCRIPTION", "API for fetching and storing 7TV emotes"),
		APIVersion:       getEnvWithDefault("API_VERSION", "1.0.0"),
	}

	// Log configuration with sensitive data masked
	LogConfiguration(config)

	return config
}

// LogConfiguration logs the current configuration with sensitive data masked
func LogConfiguration(cfg *Config) {
	log.Println("🔧 Configuration loaded:")
	log.Printf("  Redis: %s:%s (DB: %d)", cfg.RedisHost, cfg.RedisPort, cfg.RedisDB)
	log.Printf("  Cache TTL: %v | Trending TTL: %v", cfg.CacheTTL, cfg.TrendingCacheTTL)

	// Azure connection status
	if cfg.AzureConnStr == "" {
		log.Printf("  ⚠️  Azure Storage: DISABLED (connection string not set)")
	} else {
		log.Printf("  ✅ Azure Storage: ENABLED (Container: %s)", cfg.ContainerName)
	}
}
