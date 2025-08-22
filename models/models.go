// models/models.go
package models

type EmoteResponse struct {
	FileName  string `json:"fileName"`
	URL       string `json:"url"`
	EmoteID   string `json:"emoteId"`
	EmoteName string `json:"emoteName"`
	Owner     string `json:"owner,omitempty"`
	Animated  bool   `json:"animated,omitempty"`
	Scale     int    `json:"scale,omitempty"`
	Mime      string `json:"mime,omitempty"`
}

type SearchResponse struct {
	Success        bool            `json:"success"`
	TotalFound     int             `json:"totalFound"`
	Emotes         []EmoteResponse `json:"emotes"`
	Message        string          `json:"message,omitempty"`
	Cached         bool            `json:"cached,omitempty"`
	ProcessingTime float64         `json:"processingTime,omitempty"`
	Page           int             `json:"page,omitempty"`
	TotalPages     int             `json:"totalPages,omitempty"`
	ResultsPerPage int             `json:"resultsPerPage,omitempty"`
	HasNextPage    bool            `json:"hasNextPage,omitempty"`
}

type SearchRequest struct {
	Query        string `json:"query"`
	Limit        int    `json:"limit,omitempty"`
	AnimatedOnly bool   `json:"animated_only,omitempty"`
}

type TrendingPeriod string

const (
	Daily   TrendingPeriod = "trending_daily"
	Weekly  TrendingPeriod = "trending_weekly"
	Monthly TrendingPeriod = "trending_monthly"
	AllTime TrendingPeriod = "popularity"
)
