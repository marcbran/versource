package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/marcbran/versource/internal"
	http2 "github.com/marcbran/versource/internal/http/server"
)

func (c *Client) GetRebase(ctx context.Context, req internal.GetRebaseRequest) (*internal.GetRebaseResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets/%s/rebases/%d", c.baseURL, req.ChangesetName, req.RebaseID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp http2.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var rebaseResp internal.GetRebaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&rebaseResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &rebaseResp, nil
}

func (c *Client) ListRebases(ctx context.Context, req internal.ListRebasesRequest) (*internal.ListRebasesResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets/%s/rebases", c.baseURL, req.ChangesetName)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp http2.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var rebasesResp internal.ListRebasesResponse
	if err := json.NewDecoder(resp.Body).Decode(&rebasesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &rebasesResp, nil
}

func (c *Client) CreateRebase(ctx context.Context, req internal.CreateRebaseRequest) (*internal.CreateRebaseResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets/%s/rebases", c.baseURL, req.ChangesetName)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errorResp http2.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var rebaseResp internal.CreateRebaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&rebaseResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &rebaseResp, nil
}
