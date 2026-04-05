package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/leo/synthwaves-cli/internal/models"
)

type ProfileService struct {
	client *Client
}

func NewProfileService(c *Client) *ProfileService {
	return &ProfileService{client: c}
}

func (s *ProfileService) Get(ctx context.Context) (*models.Profile, error) {
	body, _, err := s.client.Get(ctx, "/api/v1/me", nil)
	if err != nil {
		return nil, err
	}
	var profile models.Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("parsing profile: %w", err)
	}
	return &profile, nil
}

func (s *ProfileService) Update(ctx context.Context, fields map[string]any) (*models.Profile, error) {
	body, _, err := s.client.Patch(ctx, "/api/v1/me", fields)
	if err != nil {
		return nil, err
	}
	var profile models.Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("parsing profile: %w", err)
	}
	return &profile, nil
}
