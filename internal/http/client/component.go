package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/marcbran/versource/internal"
	http2 "github.com/marcbran/versource/internal/http/server"
)

func (c *Client) GetComponent(ctx context.Context, req internal.GetComponentRequest) (*internal.GetComponentResponse, error) {
	var url string
	if req.Changeset != nil {
		url = fmt.Sprintf("%s/api/v1/changesets/%s/components/%d", c.baseURL, *req.Changeset, req.ComponentID)
	} else {
		url = fmt.Sprintf("%s/api/v1/components/%d", c.baseURL, req.ComponentID)
	}

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

	var componentResp internal.GetComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&componentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &componentResp, nil
}

func (c *Client) ListComponents(ctx context.Context, req internal.ListComponentsRequest) (*internal.ListComponentsResponse, error) {
	var url string
	if req.Changeset != nil {
		url = fmt.Sprintf("%s/api/v1/changesets/%s/components", c.baseURL, *req.Changeset)
	} else {
		url = fmt.Sprintf("%s/api/v1/components", c.baseURL)
	}

	params := make([]string, 0)
	if req.ModuleID != nil {
		params = append(params, fmt.Sprintf("module-id=%d", *req.ModuleID))
	}
	if req.ModuleVersionID != nil {
		params = append(params, fmt.Sprintf("module-version-id=%d", *req.ModuleVersionID))
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

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

	var componentsResp internal.ListComponentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&componentsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &componentsResp, nil
}

func (c *Client) ListComponentDiffs(ctx context.Context, req internal.ListComponentDiffsRequest) (*internal.ListComponentDiffsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets/%s/components/diffs", c.baseURL, req.Changeset)

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

	var diffsResp internal.ListComponentDiffsResponse
	if err := json.NewDecoder(resp.Body).Decode(&diffsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &diffsResp, nil
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
		var errorResp http2.ErrorResponse
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
		var errorResp http2.ErrorResponse
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

func (c *Client) DeleteComponent(ctx context.Context, req internal.DeleteComponentRequest) (*internal.DeleteComponentResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets/%s/components/%d", c.baseURL, req.Changeset, req.ComponentID)
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
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

	var componentResp internal.DeleteComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&componentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &componentResp, nil
}

func (c *Client) RestoreComponent(ctx context.Context, req internal.RestoreComponentRequest) (*internal.RestoreComponentResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets/%s/components/%d/restore", c.baseURL, req.Changeset, req.ComponentID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
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

	var componentResp internal.RestoreComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&componentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &componentResp, nil
}
