// routes/trending.go
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

func getTrendingLimiter() gin.HandlerFunc {
	store, err := sredis.NewStore(cache.RedisClient)
	if err != nil {
		panic(err)
	}
	rate := limiter.Rate{Period: 15 * time.Minute, Limit: 100}
	l := limiter.New(store, rate)
	return mgin.NewMiddleware(l)
}

func trendingEmotes(c *gin.Context) {
	start := time.Now()
	periodStr := c.Query("period")
	if periodStr == "" {
		periodStr = string(models.Weekly)
	}
	period := models.TrendingPeriod(periodStr)

	limitStr := c.Query("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	pageStr := c.Query("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	animatedOnlyStr := c.Query("animated_only")
	animatedOnly, _ := strconv.ParseBool(animatedOnlyStr)

	// Nuevo parámetro emote_type para manejo más granular
	emoteType := c.Query("emote_type")
	if emoteType == "" && animatedOnlyStr != "" {
		// Mantener compatibilidad con animated_only
		if animatedOnly {
			emoteType = "animated"
		} else {
			emoteType = "all"
		}
	}
	if emoteType == "" {
		emoteType = "all" // Por defecto todos los emotes
	}

	// Validar emote_type
	var animationFilter seventv.AnimationFilter
	switch emoteType {
	case "animated":
		animationFilter = seventv.AnimatedOnly
	case "static":
		animationFilter = seventv.StaticOnly
	case "all":
		animationFilter = seventv.AllEmotes
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid emote_type. Use 'all', 'animated', or 'static'",
		})
		return
	}

	fetchLimit := page * limit
	if fetchLimit > 300 {
		fetchLimit = 300
	}

	cacheKey := cache.GetTrendingCacheKey(string(period), limit, page, emoteType)
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

	emotes := seventv.Fetch7TVTrendingEmotesAdvanced(string(period), fetchLimit, animationFilter)
	if len(emotes) == 0 {
		resp := models.SearchResponse{
			Success:        true,
			TotalFound:     0,
			Emotes:         []models.EmoteResponse{},
			Message:        fmt.Sprintf("No trending emotes found for period: %s", period),
			ProcessingTime: time.Since(start).Seconds(),
			Page:           page,
			TotalPages:     0,
			ResultsPerPage: limit,
			HasNextPage:    false,
		}
		cache.SaveToCache(cacheKey, resp, config.LoadConfig().TrendingCacheTTL)
		c.JSON(http.StatusOK, resp)
		return
	}

	totalFound := len(emotes)
	totalPages := (totalFound + limit - 1) / limit

	startIdx := (page - 1) * limit
	endIdx := startIdx + limit
	if endIdx > totalFound {
		endIdx = totalFound
	}
	pageEmotes := emotes[startIdx:endIdx]

	processed := seventv.ProcessEmotesBatch(pageEmotes, "trending_emotes")

	resp := models.SearchResponse{
		Success:        true,
		TotalFound:     totalFound,
		Emotes:         processed,
		ProcessingTime: time.Since(start).Seconds(),
		Page:           page,
		TotalPages:     totalPages,
		ResultsPerPage: limit,
		HasNextPage:    page < totalPages,
	}
	cache.SaveToCache(cacheKey, resp, config.LoadConfig().TrendingCacheTTL)
	c.JSON(http.StatusOK, resp)
}
