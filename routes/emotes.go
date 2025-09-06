// routes/emotes.go
package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"gokeki/config"
	"gokeki/models"
	"gokeki/services/cache"
	"gokeki/services/seventv"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

func getEmoteLimiter() gin.HandlerFunc {
	store, err := sredis.NewStore(cache.RedisClient)
	if err != nil {
		panic(err)
	}
	rate := limiter.Rate{Period: 15 * time.Minute, Limit: 100}
	l := limiter.New(store, rate)
	return mgin.NewMiddleware(l)
}

func searchEmotes(c *gin.Context) {
    start := time.Now()
    var req models.SearchRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // Support both `limit` (internal) and `perPage` (7TV naming)
    if req.Limit == 0 && req.PerPage > 0 {
        req.Limit = req.PerPage
    }
    if req.Query == "" {
        c.JSON(http.StatusBadRequest, gin.H{"detail": "Query parameter is required"})
        return
    }
    if req.Limit == 0 || req.Limit > 200 {
        req.Limit = 100
    }

	cacheKey := cache.GetCacheKey(req.Query, req.Limit, req.AnimatedOnly)
	cached, err := cache.GetFromCache(cacheKey)
	if err == nil && cached != nil {
		var resp models.SearchResponse
		if err := json.Unmarshal(cached, &resp); err == nil {
			resp.ProcessingTime = time.Since(start).Seconds()
			resp.Cached = true
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	emotes := seventv.Fetch7TVEmotesAPI(req.Query, req.Limit, req.AnimatedOnly)
	if len(emotes) == 0 {
		resp := models.SearchResponse{
			Success:        true,
			TotalFound:     0,
			Emotes:         []models.EmoteResponse{},
			Message:        "No emotes found for the given query",
			ProcessingTime: time.Since(start).Seconds(),
		}
		cache.SaveToCache(cacheKey, resp, config.LoadConfig().CacheTTL)
		c.JSON(http.StatusOK, resp)
		return
	}

	processed := seventv.ProcessEmotesBatch(emotes, "emote_api")

	resp := models.SearchResponse{
		Success:        true,
		TotalFound:     len(emotes),
		Emotes:         processed,
		ProcessingTime: time.Since(start).Seconds(),
	}
	cache.SaveToCache(cacheKey, resp, config.LoadConfig().CacheTTL)
	c.JSON(http.StatusOK, resp)
}
