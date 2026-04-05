package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/leo/synthwaves-cli/internal/models"
)

type AlbumService struct {
	client *Client
}

func NewAlbumService(c *Client) *AlbumService {
	return &AlbumService{client: c}
}

type AlbumListParams struct {
	Query     string
	ArtistID  int64
	Sort      string
	Direction string
	Page      int
	PerPage   int
}

func (p AlbumListParams) Values() url.Values {
	v := url.Values{}
	if p.Query != "" {
		v.Set("q", p.Query)
	}
	if p.ArtistID > 0 {
		v.Set("artist_id", fmt.Sprintf("%d", p.ArtistID))
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

func (s *AlbumService) List(ctx context.Context, p AlbumListParams) (*PaginatedResponse[models.Album], error) {
	return FetchPage[models.Album](ctx, s.client, "/api/v1/albums", p.Values(), "albums")
}

func (s *AlbumService) Get(ctx context.Context, id int64) (*models.AlbumDetail, error) {
	body, _, err := s.client.Get(ctx, fmt.Sprintf("/api/v1/albums/%d", id), nil)
	if err != nil {
		return nil, err
	}
	var album models.AlbumDetail
	if err := json.Unmarshal(body, &album); err != nil {
		return nil, fmt.Errorf("parsing album: %w", err)
	}
	return &album, nil
}

func (s *AlbumService) Create(ctx context.Context, title string, artistID int64, year int, genre string) (*models.Album, error) {
	fields := map[string]any{
		"title":     title,
		"artist_id": artistID,
	}
	if year > 0 {
		fields["year"] = year
	}
	if genre != "" {
		fields["genre"] = genre
	}
	body, _, err := s.client.Post(ctx, "/api/v1/albums", map[string]any{"album": fields})
	if err != nil {
		return nil, err
	}
	var album models.Album
	if err := json.Unmarshal(body, &album); err != nil {
		return nil, fmt.Errorf("parsing album: %w", err)
	}
	return &album, nil
}

func (s *AlbumService) Update(ctx context.Context, id int64, fields map[string]any) (*models.Album, error) {
	body, _, err := s.client.Patch(ctx, fmt.Sprintf("/api/v1/albums/%d", id), map[string]any{"album": fields})
	if err != nil {
		return nil, err
	}
	var album models.Album
	if err := json.Unmarshal(body, &album); err != nil {
		return nil, fmt.Errorf("parsing album: %w", err)
	}
	return &album, nil
}

func (s *AlbumService) Delete(ctx context.Context, id int64) error {
	_, err := s.client.Delete(ctx, fmt.Sprintf("/api/v1/albums/%d", id))
	return err
}
