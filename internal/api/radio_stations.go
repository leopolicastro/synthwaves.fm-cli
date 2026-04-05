package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/leo/synthwaves-cli/internal/models"
)

type RadioStationService struct {
	client *Client
}

func NewRadioStationService(c *Client) *RadioStationService {
	return &RadioStationService{client: c}
}

func (s *RadioStationService) List(ctx context.Context) ([]models.RadioStation, error) {
	body, _, err := s.client.Get(ctx, "/api/v1/radio_stations", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		RadioStations []models.RadioStation `json:"radio_stations"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing radio stations: %w", err)
	}
	return resp.RadioStations, nil
}

func (s *RadioStationService) Get(ctx context.Context, id int64) (*models.RadioStation, error) {
	body, _, err := s.client.Get(ctx, fmt.Sprintf("/api/v1/radio_stations/%d", id), nil)
	if err != nil {
		return nil, err
	}
	var station models.RadioStation
	if err := json.Unmarshal(body, &station); err != nil {
		return nil, fmt.Errorf("parsing radio station: %w", err)
	}
	return &station, nil
}

func (s *RadioStationService) Create(ctx context.Context, playlistID int64) (*models.RadioStation, error) {
	body, _, err := s.client.Post(ctx, "/api/v1/radio_stations", map[string]any{
		"playlist_id": playlistID,
	})
	if err != nil {
		return nil, err
	}
	var station models.RadioStation
	if err := json.Unmarshal(body, &station); err != nil {
		return nil, fmt.Errorf("parsing radio station: %w", err)
	}
	return &station, nil
}

func (s *RadioStationService) Update(ctx context.Context, id int64, fields map[string]any) (*models.RadioStation, error) {
	body, _, err := s.client.Patch(ctx, fmt.Sprintf("/api/v1/radio_stations/%d", id), map[string]any{
		"radio_station": fields,
	})
	if err != nil {
		return nil, err
	}
	var station models.RadioStation
	if err := json.Unmarshal(body, &station); err != nil {
		return nil, fmt.Errorf("parsing radio station: %w", err)
	}
	return &station, nil
}

func (s *RadioStationService) Delete(ctx context.Context, id int64) error {
	_, err := s.client.Delete(ctx, fmt.Sprintf("/api/v1/radio_stations/%d", id))
	return err
}

func (s *RadioStationService) Control(ctx context.Context, id int64, action string) (*models.RadioControlResult, error) {
	body, _, err := s.client.Post(ctx, fmt.Sprintf("/api/v1/radio_stations/%d/control", id), map[string]any{
		"action_name": action,
	})
	if err != nil {
		return nil, err
	}
	var result models.RadioControlResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing control result: %w", err)
	}
	return &result, nil
}
