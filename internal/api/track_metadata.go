package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/leo/synthwaves-cli/internal/models"
)

type TrackMetadataService struct {
	client *Client
}

func NewTrackMetadataService(c *Client) *TrackMetadataService {
	return &TrackMetadataService{client: c}
}

func (s *TrackMetadataService) Get(ctx context.Context) (*models.TrackMetadata, error) {
	body, _, err := s.client.Get(ctx, "/api/v1/track_metadata", nil)
	if err != nil {
		return nil, err
	}
	var meta models.TrackMetadata
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, fmt.Errorf("parsing track metadata: %w", err)
	}
	return &meta, nil
}
