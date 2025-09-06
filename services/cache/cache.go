// services/cache/cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gokeki/config"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis(cfg *config.Config) {
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			panic(err)
		}
		RedisClient = redis.NewClient(opt)
	} else {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
	}
	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}

func GetCacheKey(query string, limit int, animatedOnly bool) string {
	return fmt.Sprintf("emote_search:%s:%d:%t", query, limit, animatedOnly)
}

func GetTrendingCacheKey(period string, limit int, page int, emoteType string) string {
	return fmt.Sprintf("trending:%s:%d:%d:%s", period, limit, page, emoteType)
}

// GetTrendingCacheKeyLegacy mantiene compatibilidad con animated_only boolean
func GetTrendingCacheKeyLegacy(period string, limit int, page int, animatedOnly bool) string {
	emoteType := "all"
	if animatedOnly {
		emoteType = "animated"
	}
	return GetTrendingCacheKey(period, limit, page, emoteType)
}

func GetFromCache(key string) ([]byte, error) {
	val, err := RedisClient.Get(context.Background(), key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func SaveToCache(key string, data interface{}, ttl time.Duration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return RedisClient.Set(context.Background(), key, bytes, ttl).Err()
}
