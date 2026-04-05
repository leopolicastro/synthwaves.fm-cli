package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/leo/synthwaves-cli/internal/models"
)

type TagService struct {
	client *Client
}

func NewTagService(c *Client) *TagService {
	return &TagService{client: c}
}

func (s *TagService) List(ctx context.Context, tagType, query string) ([]models.Tag, error) {
	params := url.Values{}
	if tagType != "" {
		params.Set("type", tagType)
	}
	if query != "" {
		params.Set("q", query)
	}
	body, _, err := s.client.Get(ctx, "/api/v1/tags", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Tags []models.Tag `json:"tags"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing tags: %w", err)
	}
	return resp.Tags, nil
}

type TaggingService struct {
	client *Client
}

func NewTaggingService(c *Client) *TaggingService {
	return &TaggingService{client: c}
}

func (s *TaggingService) Create(ctx context.Context, name, tagType, taggableType string, taggableID int64) (*models.Tagging, error) {
	body, _, err := s.client.Post(ctx, "/api/v1/taggings", map[string]any{
		"name":          name,
		"tag_type":      tagType,
		"taggable_type": taggableType,
		"taggable_id":   taggableID,
	})
	if err != nil {
		return nil, err
	}
	var tagging models.Tagging
	if err := json.Unmarshal(body, &tagging); err != nil {
		return nil, fmt.Errorf("parsing tagging: %w", err)
	}
	return &tagging, nil
}

func (s *TaggingService) Delete(ctx context.Context, id int64) error {
	_, err := s.client.Delete(ctx, fmt.Sprintf("/api/v1/taggings/%d", id))
	return err
}
