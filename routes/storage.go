// routes/storage.go
package routes

import (
	"fmt"
	"hash/crc32"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gokeki/models"
	"gokeki/services/cache"
	"gokeki/services/storage"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

func getStorageLimiter() gin.HandlerFunc {
	store, err := sredis.NewStore(cache.RedisClient)
	if err != nil {
		panic(err)
	}
	rate := limiter.Rate{Period: 15 * time.Minute, Limit: 50}
	l := limiter.New(store, rate)
	return mgin.NewMiddleware(l)
}

func getTrendingEmotesFromStorage(c *gin.Context) {
	start := time.Now()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	if !storage.AzureStorageAvailable() {
		c.JSON(http.StatusOK, models.SearchResponse{
			Success:        false,
			TotalFound:     0,
			Emotes:         []models.EmoteResponse{},
			Message:        "Azure Storage is not properly configured or unavailable",
			ProcessingTime: time.Since(start).Seconds(),
			Page:           page,
			TotalPages:     0,
			ResultsPerPage: limit,
			HasNextPage:    false,
		})
		return
	}

	prefix := "trending_emotes/"
	blobList, err := storage.ListBlobsWithPrefix(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SearchResponse{
			Success:        false,
			Message:        fmt.Sprintf("Error accessing Azure Storage: %v", err),
			ProcessingTime: time.Since(start).Seconds(),
		})
		return
	}

	sort.Slice(blobList, func(i, j int) bool {
		if blobList[i].Name == nil || blobList[j].Name == nil {
			return false
		}
		return *blobList[i].Name < *blobList[j].Name
	})

	totalFound := len(blobList)
	totalPages := (totalFound + limit - 1) / limit

	if totalFound == 0 {
		c.JSON(http.StatusOK, models.SearchResponse{
			Success:        true,
			TotalFound:     0,
			Emotes:         []models.EmoteResponse{},
			Message:        "No trending emotes found in storage",
			ProcessingTime: time.Since(start).Seconds(),
			Page:           page,
			TotalPages:     0,
			ResultsPerPage: limit,
			HasNextPage:    false,
		})
		return
	}

	startIdx := (page - 1) * limit
	if startIdx >= totalFound {
		c.JSON(http.StatusOK, models.SearchResponse{
			Success:        false,
			TotalFound:     totalFound,
			Emotes:         []models.EmoteResponse{},
			Message:        fmt.Sprintf("Page %d exceeds available pages (total: %d)", page, totalPages),
			ProcessingTime: time.Since(start).Seconds(),
			Page:           page,
			TotalPages:     totalPages,
			ResultsPerPage: limit,
			HasNextPage:    false,
		})
		return
	}

	endIdx := startIdx + limit
	if endIdx > totalFound {
		endIdx = totalFound
	}
	pageBlobs := blobList[startIdx:endIdx]

	processed := []models.EmoteResponse{}
	for _, b := range pageBlobs {
		if b.Name == nil {
			continue
		}
		fileName := strings.TrimPrefix(*b.Name, prefix)
		if fileName == "" || strings.HasSuffix(fileName, "/") {
			continue
		}
		blobURL := fmt.Sprintf("%s/%s", storage.ContainerURL(), *b.Name)
		emoteName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		hashValue := crc32.ChecksumIEEE([]byte(*b.Name))
		emoteID := fmt.Sprintf("storage_%d", hashValue%10000000)
		processed = append(processed, models.EmoteResponse{
			FileName:  fileName,
			URL:       blobURL,
			EmoteID:   emoteID,
			EmoteName: emoteName,
		})
	}

	c.JSON(http.StatusOK, models.SearchResponse{
		Success:        true,
		TotalFound:     totalFound,
		Emotes:         processed,
		ProcessingTime: time.Since(start).Seconds(),
		Page:           page,
		TotalPages:     totalPages,
		ResultsPerPage: limit,
		HasNextPage:    page < totalPages,
	})
}

func getEmotesFromStorage(c *gin.Context) {
	start := time.Now()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	if !storage.AzureStorageAvailable() {
		c.JSON(http.StatusOK, models.SearchResponse{
			Success:        false,
			TotalFound:     0,
			Emotes:         []models.EmoteResponse{},
			Message:        "Azure Storage is not properly configured or unavailable",
			ProcessingTime: time.Since(start).Seconds(),
			Page:           page,
			TotalPages:     0,
			ResultsPerPage: limit,
			HasNextPage:    false,
		})
		return
	}

	prefix := "emote_api/"
	blobList, err := storage.ListBlobsWithPrefix(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SearchResponse{
			Success:        false,
			Message:        fmt.Sprintf("Error accessing Azure Storage: %v", err),
			ProcessingTime: time.Since(start).Seconds(),
		})
		return
	}

	sort.Slice(blobList, func(i, j int) bool {
		if blobList[i].Name == nil || blobList[j].Name == nil {
			return false
		}
		return *blobList[i].Name < *blobList[j].Name
	})

	totalFound := len(blobList)
	totalPages := (totalFound + limit - 1) / limit

	if totalFound == 0 {
		c.JSON(http.StatusOK, models.SearchResponse{
			Success:        true,
			TotalFound:     0,
			Emotes:         []models.EmoteResponse{},
			Message:        "No emotes found in storage",
			ProcessingTime: time.Since(start).Seconds(),
			Page:           page,
			TotalPages:     0,
			ResultsPerPage: limit,
			HasNextPage:    false,
		})
		return
	}

	startIdx := (page - 1) * limit
	if startIdx >= totalFound {
		c.JSON(http.StatusOK, models.SearchResponse{
			Success:        false,
			TotalFound:     totalFound,
			Emotes:         []models.EmoteResponse{},
			Message:        fmt.Sprintf("Page %d exceeds available pages (total: %d)", page, totalPages),
			ProcessingTime: time.Since(start).Seconds(),
			Page:           page,
			TotalPages:     totalPages,
			ResultsPerPage: limit,
			HasNextPage:    false,
		})
		return
	}

	endIdx := startIdx + limit
	if endIdx > totalFound {
		endIdx = totalFound
	}
	pageBlobs := blobList[startIdx:endIdx]

	processed := []models.EmoteResponse{}
	for _, b := range pageBlobs {
		if b.Name == nil {
			continue
		}
		fileName := strings.TrimPrefix(*b.Name, prefix)
		if fileName == "" || strings.HasSuffix(fileName, "/") {
			continue
		}
		blobURL := fmt.Sprintf("%s/%s", storage.ContainerURL(), *b.Name)
		emoteName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		hashValue := crc32.ChecksumIEEE([]byte(*b.Name))
		emoteID := fmt.Sprintf("storage_%d", hashValue%10000000)
		processed = append(processed, models.EmoteResponse{
			FileName:  fileName,
			URL:       blobURL,
			EmoteID:   emoteID,
			EmoteName: emoteName,
		})
	}

	c.JSON(http.StatusOK, models.SearchResponse{
		Success:        true,
		TotalFound:     totalFound,
		Emotes:         processed,
		ProcessingTime: time.Since(start).Seconds(),
		Page:           page,
		TotalPages:     totalPages,
		ResultsPerPage: limit,
		HasNextPage:    page < totalPages,
	})
}
