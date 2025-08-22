// services/seventv/seventv.go
package seventv

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gokeki/models"
	"gokeki/services/storage"

	"golang.org/x/sync/errgroup"
)

type Owner struct {
	MainConnection struct {
		PlatformDisplayName string `json:"platformDisplayName"`
	} `json:"mainConnection"`
}

type Image struct {
	URL        string `json:"url"`
	Mime       string `json:"mime"`
	Size       int    `json:"size"`
	Scale      int    `json:"scale"`
	Width      int    `json:"width"`
	FrameCount int    `json:"frameCount"`
}

type InEmoteSet struct {
	EmoteSetID string `json:"emoteSetId"`
	Emote      struct {
		ID    string `json:"id"`
		Alias string `json:"alias"`
	} `json:"emote"`
}

type Emote struct {
	ID          string       `json:"id"`
	DefaultName string       `json:"defaultName"`
	Owner       Owner        `json:"owner"`
	Images      []Image      `json:"images"`
	Ranking     int          `json:"ranking"`
	InEmoteSets []InEmoteSet `json:"inEmoteSets"`
}

type searchResponse struct {
	Data struct {
		Emotes struct {
			Search struct {
				Items      []Emote `json:"items"`
				TotalCount int     `json:"totalCount"`
				PageCount  int     `json:"pageCount"`
			} `json:"search"`
		} `json:"emotes"`
	} `json:"data"`
}

type TrendingFile struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type TrendingHost struct {
	URL   string         `json:"url"`
	Files []TrendingFile `json:"files"`
}

type TrendingItem struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Animated bool         `json:"animated"`
	Host     TrendingHost `json:"host"`
}

type trendingResponse struct {
	Data struct {
		Emotes struct {
			Items []TrendingItem `json:"items"`
		} `json:"emotes"`
	} `json:"data"`
}

func Fetch7TVEmotesAPI(query string, limit int, animatedOnly bool) []Emote {
	url := "https://api.7tv.app/v4/gql"
	gql := `
    query EmoteSearch($query: String, $tags: [String!]!, $sortBy: SortBy!, $filters: Filters, $page: Int, $perPage: Int!, $isDefaultSetSet: Boolean!, $defaultSetId: Id!) {
      emotes {
        search(
          query: $query
          tags: { tags: $tags, match: ANY }
          sort: { sortBy: $sortBy, order: DESCENDING }
          filters: $filters
          page: $page
          perPage: $perPage
        ) {
          items {
            id
            defaultName
            owner {
              mainConnection {
                platformDisplayName
              }
            }
            images {
              url
              mime
              size
              scale
              width
              frameCount
            }
            ranking(ranking: TRENDING_WEEKLY)
            inEmoteSets(emoteSetIds: [$defaultSetId]) @include(if: $isDefaultSetSet) {
              emoteSetId
              emote {
                id
                alias
              }
            }
          }
          totalCount
          pageCount
        }
      }
    }
    `
	variables := map[string]interface{}{
		"defaultSetId":    "",
		"filters":         map[string]bool{"animated": animatedOnly},
		"isDefaultSetSet": false,
		"page":            1,
		"perPage":         limit,
		"query":           query,
		"sortBy":          "TOP_ALL_TIME",
		"tags":            []string{},
	}
	payload := map[string]interface{}{
		"query":     gql,
		"variables": variables,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
		return nil
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error from 7TV API: %d", resp.StatusCode)
		return nil
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		log.Printf("Error decoding response: %v", err)
		return nil
	}

	return sr.Data.Emotes.Search.Items
}

func Fetch7TVTrendingEmotes(period string, limit int, animatedOnly bool) []Emote {
	url := "https://api.7tv.app/v4/gql"
	gql := `
    query GetTrendingEmotes($limit: Int, $filter: EmoteSearchFilter, $period: String!) {
      emotes(query: "", limit: $limit, filter: $filter, sort: { value: $period, order: DESCENDING }) {
        items {
          id
          name
          animated
          host {
            url
            files {
              name
              format
              width
              height
            }
          }
        }
      }
    }
    `
	variables := map[string]interface{}{
		"limit":  limit,
		"filter": map[string]interface{}{"animated": animatedOnly},
		"period": period,
	}
	payload := map[string]interface{}{
		"query":     gql,
		"variables": variables,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
		return nil
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error from 7TV API: %d", resp.StatusCode)
		return nil
	}

	var tr trendingResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		log.Printf("Error decoding response: %v", err)
		return nil
	}

	emotes := make([]Emote, len(tr.Data.Emotes.Items))
	for i, item := range tr.Data.Emotes.Items {
		emotes[i] = Emote{
			ID:          item.ID,
			DefaultName: item.Name,
			Images:      buildImages(item.Host, item.Animated),
			Owner:       Owner{},
			Ranking:     0,
			InEmoteSets: nil,
		}
	}
	return emotes
}

func buildImages(host TrendingHost, animated bool) []Image {
	images := make([]Image, len(host.Files))
	for i, f := range host.Files {
		scaleStr := strings.TrimSuffix(f.Name, "x."+strings.ToLower(f.Format))
		scale, _ := strconv.Atoi(scaleStr)
		mime := "image/" + strings.ToLower(f.Format)
		url := "https:" + host.URL + "/" + f.Name
		frameCount := 1
		if animated {
			frameCount = 2 // Arbitrary value >1 to indicate animated
		}
		images[i] = Image{
			URL:        url,
			Mime:       mime,
			Scale:      scale,
			Width:      f.Width,
			FrameCount: frameCount,
		}
	}
	return images
}

func selectBestImage(images []Image) *Image {
	if len(images) == 0 {
		return nil
	}

	// Prioritize animated 4x.webp, etc.
	animatedImages := []Image{}
	staticImages := []Image{}
	for _, img := range images {
		if img.FrameCount > 1 {
			animatedImages = append(animatedImages, img)
		} else {
			staticImages = append(staticImages, img)
		}
	}

	var candidates []Image
	if len(animatedImages) > 0 {
		candidates = animatedImages
	} else {
		candidates = staticImages
	}

	// Sort by preference: webp > gif > avif > png, higher scale
	preferredMimes := map[string]int{"image/webp": 4, "image/gif": 3, "image/avif": 2, "image/png": 1}
	sort.Slice(candidates, func(i, j int) bool {
		mi := preferredMimes[candidates[i].Mime]
		mj := preferredMimes[candidates[j].Mime]
		if mi != mj {
			return mi > mj
		}
		return candidates[i].Scale > candidates[j].Scale
	})

	return &candidates[0]
}

var safeNameRe = regexp.MustCompile(`[^a-zA-Z0-9\._\- ]`)

func processEmote(e Emote, folder string) *models.EmoteResponse {
	bestImage := selectBestImage(e.Images)
	if bestImage == nil {
		return nil
	}

	resp, err := http.Get(bestImage.URL)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Failed to download emote %s: %v", e.DefaultName, err)
		return nil
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	extension := ".png"
	switch bestImage.Mime {
	case "image/webp":
		extension = ".webp"
	case "image/gif":
		extension = ".gif"
	case "image/avif":
		extension = ".avif"
	}

	safeName := safeNameRe.ReplaceAllString(e.DefaultName, "_")
	fileName := safeName + extension
	blobName := folder + "/" + fileName

	url, err := storage.UploadToAzureBlob(data, blobName, bestImage.Mime)
	if err != nil || url == "" {
		log.Printf("Error uploading emote %s: %v", e.DefaultName, err)
		return nil
	}

	return &models.EmoteResponse{
		FileName:  fileName,
		URL:       url,
		EmoteID:   e.ID,
		EmoteName: e.DefaultName,
		Owner:     e.Owner.MainConnection.PlatformDisplayName,
		Animated:  bestImage.FrameCount > 1,
		Scale:     bestImage.Scale,
		Mime:      bestImage.Mime,
	}
}

func ProcessEmotesBatch(emotes []Emote, folder string) []models.EmoteResponse {
	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(10)

	processedChan := make(chan *models.EmoteResponse, len(emotes))

	for _, e := range emotes {
		e := e
		g.Go(func() error {
			res := processEmote(e, folder)
			if res != nil {
				processedChan <- res
			}
			return nil
		})
	}

	_ = g.Wait()
	close(processedChan)

	var result []models.EmoteResponse
	for res := range processedChan {
		result = append(result, *res)
	}
	return result
}
