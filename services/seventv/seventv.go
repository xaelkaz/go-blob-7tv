// services/seventv/seventv.go
package seventv

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "log"
    "net/http"
    "sort"

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
    // Build filters: align with client boolean semantics.
    // animated_only=true  => animated: true (solo animados)
    // animated_only=false => animated: false (solo estáticos)
    filters := map[string]interface{}{"animated": animatedOnly}

    variables := map[string]interface{}{
        "defaultSetId":    "",
        "filters":         filters,
        "isDefaultSetSet": false,
        "page":            1,
        "perPage":         limit,
        "query":           query,
        "sortBy":          "TOP_ALL_TIME",
        "tags":            []string{},
    }
    payload := map[string]interface{}{
        "operationName": "EmoteSearch",
        "query":         gql,
        "variables":     variables,
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

// AnimationFilter represents the type of emotes to fetch based on animation
type AnimationFilter int

const (
	AllEmotes    AnimationFilter = iota // Todos los emotes
	AnimatedOnly                        // Solo emotes animados
	StaticOnly                          // Solo emotes estáticos
)

func Fetch7TVTrendingEmotes(period string, limit int, animatedOnly bool) []Emote {
	// Convert boolean to AnimationFilter for backward compatibility
	var animationFilter AnimationFilter
	if animatedOnly {
		animationFilter = AnimatedOnly
	} else {
		animationFilter = AllEmotes
	}

	// Use the advanced function internally
	return Fetch7TVTrendingEmotesAdvanced(period, limit, animationFilter)
} // Fetch7TVTrendingEmotesAdvanced allows more granular control over animation filtering
func Fetch7TVTrendingEmotesAdvanced(period string, limit int, animationFilter AnimationFilter) []Emote {
	url := "https://api.7tv.app/v4/gql"
	gql := `
	query EmoteSearch($query: String, $tags: [String!]!, $sortBy: SortBy!, $filters: Filters, $page: Int, $perPage: Int!, $isDefaultSetSet: Boolean!, $defaultSetId: Id!) {
	  emotes {
	    search(
	      query: $query
	      tags: {tags: $tags, match: ANY}
	      sort: {sortBy: $sortBy, order: DESCENDING}
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
	            __typename
	          }
	          style {
	            activePaint {
	              id
	              name
	              data {
	                layers {
	                  id
	                  ty {
	                    __typename
	                    ... on PaintLayerTypeSingleColor {
	                      color {
	                        hex
	                        __typename
	                      }
	                      __typename
	                    }
	                    ... on PaintLayerTypeLinearGradient {
	                      angle
	                      repeating
	                      stops {
	                        at
	                        color {
	                          hex
	                          __typename
	                        }
	                        __typename
	                      }
	                      __typename
	                    }
	                    ... on PaintLayerTypeRadialGradient {
	                      repeating
	                      stops {
	                        at
	                        color {
	                          hex
	                          __typename
	                        }
	                        __typename
	                      }
	                      shape
	                      __typename
	                    }
	                    ... on PaintLayerTypeImage {
	                      images {
	                        url
	                        mime
	                        size
	                        scale
	                        width
	                        height
	                        frameCount
	                        __typename
	                      }
	                      __typename
	                    }
	                  }
	                  opacity
	                  __typename
	                }
	                shadows {
	                  color {
	                    hex
	                    __typename
	                  }
	                  offsetX
	                  offsetY
	                  blur
	                  __typename
	                }
	                __typename
	              }
	              __typename
	            }
	            __typename
	          }
	          highestRoleColor {
	            hex
	            __typename
	          }
	          __typename
	        }
	        deleted
	        flags {
	          defaultZeroWidth
	          private
	          publicListed
	          __typename
	        }
	        imagesPending
	        images {
	          url
	          mime
	          size
	          scale
	          width
	          frameCount
	          __typename
	        }
	        ranking(ranking: TRENDING_WEEKLY)
	        inEmoteSets(emoteSetIds: [$defaultSetId]) @include(if: $isDefaultSetSet) {
	          emoteSetId
	          emote {
	            id
	            alias
	            __typename
	          }
	          __typename
	        }
	        __typename
	      }
	      totalCount
	      pageCount
	      __typename
	    }
	    __typename
	  }
	}
	`

	// Convert period to sortBy format
	var sortBy string
	switch period {
	case "trending_daily":
		sortBy = "TRENDING_DAILY"
	case "trending_weekly":
		sortBy = "TRENDING_WEEKLY"
	case "trending_monthly":
		sortBy = "TRENDING_MONTHLY"
	default:
		sortBy = "TRENDING_MONTHLY"
	}

	// Build filters object based on animation filter
	var filters map[string]interface{}
	switch animationFilter {
	case AnimatedOnly:
		filters = map[string]interface{}{
			"animated": true,
		}
	case StaticOnly:
		filters = map[string]interface{}{
			"animated": false,
		}
	default: // AllEmotes
		filters = map[string]interface{}{}
	}

	variables := map[string]interface{}{
		"defaultSetId":    "",
		"filters":         filters,
		"isDefaultSetSet": false,
		"page":            1,
		"perPage":         limit,
		"query":           nil,
		"sortBy":          sortBy,
		"tags":            []string{},
	}
	payload := map[string]interface{}{
		"operationName": "EmoteSearch",
		"query":         gql,
		"variables":     variables,
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

// name sanitizer removed; filenames now use emote ID to ensure uniqueness

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

    // Use stable unique naming to avoid collisions between emotes sharing names
    // and between static/animated variants of the same emote.
    variant := "static"
    if bestImage.FrameCount > 1 {
        variant = "anim"
    }
    fileName := e.ID + "_" + variant + extension
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
