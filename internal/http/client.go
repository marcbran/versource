package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/marcbran/versource/internal"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(config *internal.Config) *Client {
	baseURL := fmt.Sprintf("http://%s:%s", config.HTTP.Hostname, config.HTTP.Port)
	if config.HTTP.Hostname == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", config.HTTP.Port)
	}

	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
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
		var errorResp ErrorResponse
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

func (c *Client) CreateComponent(ctx context.Context, req internal.CreateComponentRequest) (*internal.CreateComponentResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/changesets/%s/components", c.baseURL, req.Changeset)
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
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var componentResp internal.CreateComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&componentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &componentResp, nil
}

func (c *Client) UpdateComponent(ctx context.Context, req internal.UpdateComponentRequest) (*internal.UpdateComponentResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/changesets/%s/components/%d", c.baseURL, req.Changeset, req.ComponentID)
	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(jsonData))
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
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var componentResp internal.UpdateComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&componentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &componentResp, nil
}

func (c *Client) CreatePlan(ctx context.Context, req internal.CreatePlanRequest) (*internal.CreatePlanResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets/%s/components/%d/plans", c.baseURL, req.Changeset, req.ComponentID)
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

	if resp.StatusCode != http.StatusCreated {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var planResp internal.CreatePlanResponse
	if err := json.NewDecoder(resp.Body).Decode(&planResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &planResp, nil
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
		var errorResp ErrorResponse
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
