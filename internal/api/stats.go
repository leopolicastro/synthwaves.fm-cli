package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/leo/synthwaves-cli/internal/models"
)

type StatsService struct {
	client *Client
}

func NewStatsService(c *Client) *StatsService {
	return &StatsService{client: c}
}

func (s *StatsService) Get(ctx context.Context, timeRange string) (*models.Stats, error) {
	params := url.Values{}
	if timeRange != "" {
		params.Set("time_range", timeRange)
	}
	body, _, err := s.client.Get(ctx, "/api/v1/stats", params)
	if err != nil {
		return nil, err
	}
	var stats models.Stats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("parsing stats: %w", err)
	}
	return &stats, nil
}
