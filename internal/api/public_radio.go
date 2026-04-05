package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/leo/synthwaves-cli/internal/models"
)

type PublicRadioService struct {
	client *Client
}

func NewPublicRadioService(c *Client) *PublicRadioService {
	return &PublicRadioService{client: c}
}

func (s *PublicRadioService) List(ctx context.Context) ([]models.PublicRadioStation, error) {
	body, _, err := s.client.Get(ctx, "/api/v1/radio", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		RadioStations []models.PublicRadioStation `json:"radio_stations"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing public radio stations: %w", err)
	}
	return resp.RadioStations, nil
}

func (s *PublicRadioService) Get(ctx context.Context, slug string) (*models.PublicRadioStation, error) {
	body, _, err := s.client.Get(ctx, "/api/v1/radio/"+slug, nil)
	if err != nil {
		return nil, err
	}
	var station models.PublicRadioStation
	if err := json.Unmarshal(body, &station); err != nil {
		return nil, fmt.Errorf("parsing public radio station: %w", err)
	}
	return &station, nil
}
