// routes/routes.go
package routes

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.POST("/search-emotes", getEmoteLimiter(), searchEmotes)

	trending := r.Group("/api/trending")
	trending.GET("/emotes", getTrendingLimiter(), trendingEmotes)

	storageGroup := r.Group("/api/storage")
	storageGroup.GET("/trending-emotes", getStorageLimiter(), getTrendingEmotesFromStorage)
	storageGroup.GET("/emote-api", getStorageLimiter(), getEmotesFromStorage)

	cacheGroup := r.Group("/api/cache")
	cacheGroup.GET("/status", getCacheStatusLimiter(), cacheStatus)
	cacheGroup.POST("/clear", getCacheClearLimiter(), clearCache)
}
