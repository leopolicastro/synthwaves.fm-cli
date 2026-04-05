package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/leo/synthwaves-cli/internal/models"
)

type TrackService struct {
	client *Client
}

func NewTrackService(c *Client) *TrackService {
	return &TrackService{client: c}
}

type TrackListParams struct {
	Query     string
	AlbumID   int64
	ArtistID  int64
	Genre     string
	Language  string
	Decade    string
	Sort      string
	Direction string
	Page      int
	PerPage   int
}

func (p TrackListParams) Values() url.Values {
	v := url.Values{}
	if p.Query != "" {
		v.Set("q", p.Query)
	}
	if p.AlbumID > 0 {
		v.Set("album_id", fmt.Sprintf("%d", p.AlbumID))
	}
	if p.ArtistID > 0 {
		v.Set("artist_id", fmt.Sprintf("%d", p.ArtistID))
	}
	if p.Genre != "" {
		v.Set("genre", p.Genre)
	}
	if p.Language != "" {
		v.Set("language", p.Language)
	}
	if p.Decade != "" {
		v.Set("decade", p.Decade)
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

func (s *TrackService) List(ctx context.Context, p TrackListParams) (*PaginatedResponse[models.Track], error) {
	return FetchPage[models.Track](ctx, s.client, "/api/v1/tracks", p.Values(), "tracks")
}

func (s *TrackService) Get(ctx context.Context, id int64) (*models.Track, error) {
	body, _, err := s.client.Get(ctx, fmt.Sprintf("/api/v1/tracks/%d", id), nil)
	if err != nil {
		return nil, err
	}
	var track models.Track
	if err := json.Unmarshal(body, &track); err != nil {
		return nil, fmt.Errorf("parsing track: %w", err)
	}
	return &track, nil
}

func (s *TrackService) Create(ctx context.Context, fields map[string]any) (*models.Track, error) {
	body, _, err := s.client.Post(ctx, "/api/v1/tracks", map[string]any{"track": fields})
	if err != nil {
		return nil, err
	}
	var track models.Track
	if err := json.Unmarshal(body, &track); err != nil {
		return nil, fmt.Errorf("parsing track: %w", err)
	}
	return &track, nil
}

func (s *TrackService) Update(ctx context.Context, id int64, fields map[string]any) (*models.Track, error) {
	body, _, err := s.client.Patch(ctx, fmt.Sprintf("/api/v1/tracks/%d", id), map[string]any{"track": fields})
	if err != nil {
		return nil, err
	}
	var track models.Track
	if err := json.Unmarshal(body, &track); err != nil {
		return nil, fmt.Errorf("parsing track: %w", err)
	}
	return &track, nil
}

func (s *TrackService) Delete(ctx context.Context, id int64) error {
	_, err := s.client.Delete(ctx, fmt.Sprintf("/api/v1/tracks/%d", id))
	return err
}

func (s *TrackService) Stream(ctx context.Context, id int64) (*models.StreamInfo, error) {
	body, _, err := s.client.Get(ctx, fmt.Sprintf("/api/v1/tracks/%d/stream", id), nil)
	if err != nil {
		return nil, err
	}
	var info models.StreamInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parsing stream info: %w", err)
	}
	return &info, nil
}
