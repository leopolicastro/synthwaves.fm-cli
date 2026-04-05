package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/leo/synthwaves-cli/internal/models"
)

type FavoriteService struct {
	client *Client
}

func NewFavoriteService(c *Client) *FavoriteService {
	return &FavoriteService{client: c}
}

type FavoriteListParams struct {
	Type    string
	Query   string
	Page    int
	PerPage int
}

func (p FavoriteListParams) Values() url.Values {
	v := url.Values{}
	if p.Type != "" {
		v.Set("type", p.Type)
	}
	if p.Query != "" {
		v.Set("q", p.Query)
	}
	if p.Page > 0 {
		v.Set("page", fmt.Sprintf("%d", p.Page))
	}
	if p.PerPage > 0 {
		v.Set("per_page", fmt.Sprintf("%d", p.PerPage))
	}
	return v
}

func (s *FavoriteService) List(ctx context.Context, p FavoriteListParams) (*PaginatedResponse[models.Favorite], error) {
	return FetchPage[models.Favorite](ctx, s.client, "/api/v1/favorites", p.Values(), "favorites")
}

func (s *FavoriteService) Add(ctx context.Context, favorableType string, favorableID int64) (*models.Favorite, error) {
	body, _, err := s.client.Post(ctx, "/api/v1/favorites", map[string]any{
		"favorable_type": favorableType,
		"favorable_id":   favorableID,
	})
	if err != nil {
		return nil, err
	}
	var fav models.Favorite
	if err := json.Unmarshal(body, &fav); err != nil {
		return nil, fmt.Errorf("parsing favorite: %w", err)
	}
	return &fav, nil
}

func (s *FavoriteService) Remove(ctx context.Context, id int64) error {
	_, err := s.client.Delete(ctx, fmt.Sprintf("/api/v1/favorites/%d", id))
	return err
}

func (s *FavoriteService) RemoveByTarget(ctx context.Context, favorableType string, favorableID int64) error {
	path := fmt.Sprintf("/api/v1/favorites?favorable_type=%s&favorable_id=%d", favorableType, favorableID)
	_, _, err := s.client.do(ctx, "DELETE", path, nil)
	return err
}
