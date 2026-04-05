package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/leo/synthwaves-cli/internal/config"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	cfg        *config.Config
	token      string
	tokenExp   time.Time
	mu         sync.Mutex
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		BaseURL: cfg.BaseURL,
		cfg:     cfg,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) ensureAuth(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Until(c.tokenExp) > 60*time.Second {
		return nil
	}

	// Try loading cached token from disk
	if tc, err := config.LoadToken(); err == nil && tc.Valid() {
		c.token = tc.Token
		c.tokenExp = tc.ExpiresAt
		return nil
	}

	return c.authenticate(ctx)
}

func (c *Client) authenticate(ctx context.Context) error {
	payload, _ := json.Marshal(map[string]string{
		"client_id":  c.cfg.ClientID,
		"secret_key": c.cfg.SecretKey,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/auth/token", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("authenticating: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return parseError(resp.StatusCode, body)
	}

	var result struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing auth response: %w", err)
	}

	c.token = result.Token
	c.tokenExp = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	// Cache to disk
	_ = config.SaveToken(&config.TokenCache{
		Token:     c.token,
		ExpiresAt: c.tokenExp,
	})

	return nil
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	if err := c.ensureAuth(ctx); err != nil {
		return nil, 0, err
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request to %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// On 401, try re-authenticating once
	if resp.StatusCode == http.StatusUnauthorized {
		c.mu.Lock()
		err := c.authenticate(ctx)
		c.mu.Unlock()
		if err != nil {
			return nil, resp.StatusCode, parseError(resp.StatusCode, respBody)
		}
		// Retry the request
		return c.do(ctx, method, path, body)
	}

	if resp.StatusCode >= 400 {
		return respBody, resp.StatusCode, parseError(resp.StatusCode, respBody)
	}

	return respBody, resp.StatusCode, nil
}

func (c *Client) Get(ctx context.Context, path string, params url.Values) ([]byte, int, error) {
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) Post(ctx context.Context, path string, body any) ([]byte, int, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) Patch(ctx context.Context, path string, body any) ([]byte, int, error) {
	return c.do(ctx, http.MethodPatch, path, body)
}

func (c *Client) Delete(ctx context.Context, path string) (int, error) {
	_, code, err := c.do(ctx, http.MethodDelete, path, nil)
	return code, err
}
