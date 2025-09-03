package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/marcbran/versource/internal"
	http2 "github.com/marcbran/versource/internal/http/server"
)

func (c *Client) ListChangesets(ctx context.Context) (*internal.ListChangesetsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets", c.baseURL)
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

	var changesetsResp internal.ListChangesetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&changesetsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &changesetsResp, nil
}

func (c *Client) CreateChangeset(ctx context.Context, req internal.CreateChangesetRequest) (*internal.CreateChangesetResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/changesets", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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

	var changesetResp internal.CreateChangesetResponse
	if err := json.NewDecoder(resp.Body).Decode(&changesetResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &changesetResp, nil
}

func (c *Client) MergeChangeset(ctx context.Context, req internal.MergeChangesetRequest) (*internal.MergeChangesetResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets/%s/merge", c.baseURL, req.ChangesetName)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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

	var changesetResp internal.MergeChangesetResponse
	if err := json.NewDecoder(resp.Body).Decode(&changesetResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &changesetResp, nil
}
