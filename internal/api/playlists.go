package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/leo/synthwaves-cli/internal/models"
)

type PlaylistService struct {
	client *Client
}

func NewPlaylistService(c *Client) *PlaylistService {
	return &PlaylistService{client: c}
}

type PlaylistListParams struct {
	Query     string
	Sort      string
	Direction string
	Page      int
	PerPage   int
}

func (p PlaylistListParams) Values() url.Values {
	v := url.Values{}
	if p.Query != "" {
		v.Set("q", p.Query)
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

func (s *PlaylistService) List(ctx context.Context, p PlaylistListParams) (*PaginatedResponse[models.Playlist], error) {
	return FetchPage[models.Playlist](ctx, s.client, "/api/v1/playlists", p.Values(), "playlists")
}

func (s *PlaylistService) Get(ctx context.Context, id int64, query string, page, perPage int) (*models.PlaylistDetail, error) {
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
	body, _, err := s.client.Get(ctx, fmt.Sprintf("/api/v1/playlists/%d", id), params)
	if err != nil {
		return nil, err
	}
	var playlist models.PlaylistDetail
	if err := json.Unmarshal(body, &playlist); err != nil {
		return nil, fmt.Errorf("parsing playlist: %w", err)
	}
	return &playlist, nil
}

func (s *PlaylistService) Create(ctx context.Context, name string, trackIDs []int64) (*models.Playlist, error) {
	fields := map[string]any{"name": name}
	if len(trackIDs) > 0 {
		fields["track_ids"] = trackIDs
	}
	body, _, err := s.client.Post(ctx, "/api/v1/playlists", map[string]any{"playlist": fields})
	if err != nil {
		return nil, err
	}
	var playlist models.Playlist
	if err := json.Unmarshal(body, &playlist); err != nil {
		return nil, fmt.Errorf("parsing playlist: %w", err)
	}
	return &playlist, nil
}

func (s *PlaylistService) Update(ctx context.Context, id int64, name string) (*models.Playlist, error) {
	body, _, err := s.client.Patch(ctx, fmt.Sprintf("/api/v1/playlists/%d", id), map[string]any{
		"playlist": map[string]any{"name": name},
	})
	if err != nil {
		return nil, err
	}
	var playlist models.Playlist
	if err := json.Unmarshal(body, &playlist); err != nil {
		return nil, fmt.Errorf("parsing playlist: %w", err)
	}
	return &playlist, nil
}

func (s *PlaylistService) Delete(ctx context.Context, id int64) error {
	_, err := s.client.Delete(ctx, fmt.Sprintf("/api/v1/playlists/%d", id))
	return err
}

func (s *PlaylistService) AddTrack(ctx context.Context, playlistID, trackID int64) (*models.PlaylistTrackResult, error) {
	body, _, err := s.client.Post(ctx, fmt.Sprintf("/api/v1/playlists/%d/tracks", playlistID), map[string]any{
		"track_id": trackID,
	})
	if err != nil {
		return nil, err
	}
	var result models.PlaylistTrackResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing result: %w", err)
	}
	return &result, nil
}

func (s *PlaylistService) AddTracks(ctx context.Context, playlistID int64, trackIDs []int64) (*models.PlaylistTrackResult, error) {
	body, _, err := s.client.Post(ctx, fmt.Sprintf("/api/v1/playlists/%d/tracks", playlistID), map[string]any{
		"track_ids": trackIDs,
	})
	if err != nil {
		return nil, err
	}
	var result models.PlaylistTrackResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing result: %w", err)
	}
	return &result, nil
}

func (s *PlaylistService) AddAlbum(ctx context.Context, playlistID, albumID int64) (*models.PlaylistTrackResult, error) {
	body, _, err := s.client.Post(ctx, fmt.Sprintf("/api/v1/playlists/%d/tracks", playlistID), map[string]any{
		"album_id": albumID,
	})
	if err != nil {
		return nil, err
	}
	var result models.PlaylistTrackResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing result: %w", err)
	}
	return &result, nil
}

func (s *PlaylistService) RemoveTrack(ctx context.Context, playlistID, playlistTrackID int64) error {
	_, err := s.client.Delete(ctx, fmt.Sprintf("/api/v1/playlists/%d/tracks/%d", playlistID, playlistTrackID))
	return err
}

func (s *PlaylistService) Reorder(ctx context.Context, playlistID int64, playlistTrackIDs []int64) error {
	_, _, err := s.client.Patch(ctx, fmt.Sprintf("/api/v1/playlists/%d/track_order", playlistID), map[string]any{
		"playlist_track_ids": playlistTrackIDs,
	})
	return err
}
