package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/leo/synthwaves-cli/internal/models"
)

type PlayHistoryService struct {
	client *Client
}

func NewPlayHistoryService(c *Client) *PlayHistoryService {
	return &PlayHistoryService{client: c}
}

func (s *PlayHistoryService) List(ctx context.Context, query string, page, perPage int) (*PaginatedResponse[models.PlayHistory], error) {
	params := url.Values{}
	if query != "" {
		params.Set("q", query)
	}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if perPage > 0 {
		params.Set("per_page", fmt.Sprintf("%d", perPage))
	}
	return FetchPage[models.PlayHistory](ctx, s.client, "/api/v1/play_histories", params, "play_histories")
}

func (s *PlayHistoryService) Record(ctx context.Context, trackID int64) (*models.PlayHistory, error) {
	body, _, err := s.client.Post(ctx, "/api/v1/play_histories", map[string]any{
		"track_id": trackID,
	})
	if err != nil {
		return nil, err
	}
	var ph models.PlayHistory
	if err := json.Unmarshal(body, &ph); err != nil {
		return nil, fmt.Errorf("parsing play history: %w", err)
	}
	return &ph, nil
}
