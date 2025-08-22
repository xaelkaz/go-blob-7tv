// routes/cache.go
package routes

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gokeki/services/cache"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

func getCacheStatusLimiter() gin.HandlerFunc {
	store, err := sredis.NewStore(cache.RedisClient)
	if err != nil {
		panic(err)
	}
	rate := limiter.Rate{Period: time.Minute, Limit: 20}
	l := limiter.New(store, rate)
	return mgin.NewMiddleware(l)
}

func getCacheClearLimiter() gin.HandlerFunc {
	store, err := sredis.NewStore(cache.RedisClient)
	if err != nil {
		panic(err)
	}
	rate := limiter.Rate{Period: time.Minute, Limit: 5}
	l := limiter.New(store, rate)
	return mgin.NewMiddleware(l)
}

func cacheStatus(c *gin.Context) {
	info, err := cache.RedisClient.Info(context.Background()).Result()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
		return
	}

	dbsize, _ := cache.RedisClient.DBSize(context.Background()).Result()
	emoteSearchKeys, _ := cache.RedisClient.Keys(context.Background(), "emote_search:*").Result()
	trendingKeys, _ := cache.RedisClient.Keys(context.Background(), "trending:*").Result()

	usedMemory := "unknown"
	hits := 0
	misses := 0
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "used_memory_human:") {
			usedMemory = strings.TrimPrefix(line, "used_memory_human:")
		}
		if strings.HasPrefix(line, "keyspace_hits:") {
			hitsStr := strings.TrimPrefix(line, "keyspace_hits:")
			hits, _ = strconv.Atoi(hitsStr)
		}
		if strings.HasPrefix(line, "keyspace_misses:") {
			missesStr := strings.TrimPrefix(line, "keyspace_misses:")
			misses, _ = strconv.Atoi(missesStr)
		}
	}

	hitRatio := 0.0
	total := hits + misses
	if total > 0 {
		hitRatio = float64(hits) / float64(total) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "connected",
		"totalKeys":       dbsize,
		"emoteSearchKeys": len(emoteSearchKeys),
		"trendingKeys":    len(trendingKeys),
		"usedMemory":      usedMemory,
		"hitRatio":        hitRatio,
	})
}

func clearCache(c *gin.Context) {
	cacheType := c.Query("cache_type")
	if cacheType == "" {
		cacheType = "all"
	}

	var patterns []string
	switch cacheType {
	case "all":
		patterns = []string{"emote_search:*", "trending:*"}
	case "search":
		patterns = []string{"emote_search:*"}
	case "trending":
		patterns = []string{"trending:*"}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid cache_type. Options are: all, search, trending"})
		return
	}

	var keys []string
	for _, pattern := range patterns {
		k, _ := cache.RedisClient.Keys(context.Background(), pattern).Result()
		keys = append(keys, k...)
	}

	if len(keys) > 0 {
		cache.RedisClient.Del(context.Background(), keys...)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Cache cleared. %d entries removed.", len(keys)),
		"type":    cacheType,
	})
}
