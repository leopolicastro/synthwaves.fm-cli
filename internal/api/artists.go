package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/leo/synthwaves-cli/internal/models"
)

type ArtistService struct {
	client *Client
}

func NewArtistService(c *Client) *ArtistService {
	return &ArtistService{client: c}
}

type ArtistListParams struct {
	Query     string
	Category  string
	Sort      string
	Direction string
	Page      int
	PerPage   int
}

func (p ArtistListParams) Values() url.Values {
	v := url.Values{}
	if p.Query != "" {
		v.Set("q", p.Query)
	}
	if p.Category != "" {
		v.Set("category", p.Category)
	}
	if p.Sort != "" {
		v.Set("sort", p.Sort)
	}
	if p.Direction != "" {
		v.Set("sort_direction", p.Direction)
	}
	if p.Page > 0 {
		v.Set("page", fmt.Sprintf("%d", p.Page))
	}
	if p.PerPage > 0 {
		v.Set("per_page", fmt.Sprintf("%d", p.PerPage))
	}
	return v
}

func (s *ArtistService) List(ctx context.Context, p ArtistListParams) (*PaginatedResponse[models.Artist], error) {
	return FetchPage[models.Artist](ctx, s.client, "/api/v1/artists", p.Values(), "artists")
}

func (s *ArtistService) Get(ctx context.Context, id int64) (*models.ArtistDetail, error) {
	body, _, err := s.client.Get(ctx, fmt.Sprintf("/api/v1/artists/%d", id), nil)
	if err != nil {
		return nil, err
	}
	var artist models.ArtistDetail
	if err := json.Unmarshal(body, &artist); err != nil {
		return nil, fmt.Errorf("parsing artist: %w", err)
	}
	return &artist, nil
}

func (s *ArtistService) Create(ctx context.Context, name, category string) (*models.Artist, error) {
	payload := map[string]any{
		"artist": map[string]any{
			"name":     name,
			"category": category,
		},
	}
	body, _, err := s.client.Post(ctx, "/api/v1/artists", payload)
	if err != nil {
		return nil, err
	}
	var artist models.Artist
	if err := json.Unmarshal(body, &artist); err != nil {
		return nil, fmt.Errorf("parsing artist: %w", err)
	}
	return &artist, nil
}

func (s *ArtistService) Update(ctx context.Context, id int64, name, category string) (*models.Artist, error) {
	fields := map[string]any{}
	if name != "" {
		fields["name"] = name
	}
	if category != "" {
		fields["category"] = category
	}
	payload := map[string]any{"artist": fields}
	body, _, err := s.client.Patch(ctx, fmt.Sprintf("/api/v1/artists/%d", id), payload)
	if err != nil {
		return nil, err
	}
	var artist models.Artist
	if err := json.Unmarshal(body, &artist); err != nil {
		return nil, fmt.Errorf("parsing artist: %w", err)
	}
	return &artist, nil
}

func (s *ArtistService) Delete(ctx context.Context, id int64) error {
	_, err := s.client.Delete(ctx, fmt.Sprintf("/api/v1/artists/%d", id))
	return err
}
