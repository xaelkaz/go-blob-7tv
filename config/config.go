// config/config.go
package config

import (
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
	return &Config{
		RedisHost:        getEnvWithDefault("REDIS_HOST", "localhost"),
		RedisPort:        getEnvWithDefault("REDIS_PORT", "6379"),
		RedisDB:          db,
		RedisPassword:    getEnvWithDefault("REDIS_PASSWORD", ""),
		RedisURL:         getEnvWithDefault("REDIS_URL", ""),
		AzureConnStr:     getEnvWithDefault("AZURE_CONNECTION_STRING", ""),
		ContainerName:    getEnvWithDefault("CONTAINER_NAME", "emotes"),
		CacheTTL:         time.Duration(ttl) * time.Second,
		TrendingCacheTTL: time.Duration(trendingTTL) * time.Second,
		APITitle:         getEnvWithDefault("API_TITLE", "7TV Emote API"),
		APIDesc:          getEnvWithDefault("API_DESCRIPTION", "API for fetching and storing 7TV emotes"),
		APIVersion:       getEnvWithDefault("API_VERSION", "1.0.0"),
	}
}
