package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/leo/synthwaves-cli/internal/models"
)

type SearchService struct {
	client *Client
}

func NewSearchService(c *Client) *SearchService {
	return &SearchService{client: c}
}

type SearchParams struct {
	Query         string
	Types         string
	Genre         string
	YearFrom      int
	YearTo        int
	Limit         int
	FavoritesOnly bool
	Category      string
	Tags          string
}

func (p SearchParams) Values() url.Values {
	v := url.Values{}
	v.Set("q", p.Query)
	if p.Types != "" {
		v.Set("types", p.Types)
	}
	if p.Genre != "" {
		v.Set("genre", p.Genre)
	}
	if p.YearFrom > 0 {
		v.Set("year_from", fmt.Sprintf("%d", p.YearFrom))
	}
	if p.YearTo > 0 {
		v.Set("year_to", fmt.Sprintf("%d", p.YearTo))
	}
	if p.Limit > 0 {
		v.Set("limit", fmt.Sprintf("%d", p.Limit))
	}
	if p.FavoritesOnly {
		v.Set("favorites_only", "1")
	}
	if p.Category != "" {
		v.Set("category", p.Category)
	}
	if p.Tags != "" {
		v.Set("tags", p.Tags)
	}
	return v
}

func (s *SearchService) Search(ctx context.Context, p SearchParams) (*models.SearchResult, error) {
	body, _, err := s.client.Get(ctx, "/api/v1/search", p.Values())
	if err != nil {
		return nil, err
	}
	var result models.SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing search results: %w", err)
	}
	return &result, nil
}
