package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	http2 "github.com/marcbran/versource/internal/http/server"
	"github.com/marcbran/versource/pkg/versource"
)

func (c *Client) GetViewResource(ctx context.Context, req versource.GetViewResourceRequest) (*versource.GetViewResourceResponse, error) {
	url := fmt.Sprintf("%s/api/v1/view-resources/%d", c.baseURL, req.ViewResourceID)
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
		err := json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var viewResourceResp versource.GetViewResourceResponse
	err = json.NewDecoder(resp.Body).Decode(&viewResourceResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &viewResourceResp, nil
}

func (c *Client) ListViewResources(ctx context.Context, req versource.ListViewResourcesRequest) (*versource.ListViewResourcesResponse, error) {
	url := fmt.Sprintf("%s/api/v1/view-resources", c.baseURL)
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
		err := json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var viewResourcesResp versource.ListViewResourcesResponse
	err = json.NewDecoder(resp.Body).Decode(&viewResourcesResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &viewResourcesResp, nil
}

func (c *Client) SaveViewResource(ctx context.Context, req versource.SaveViewResourceRequest) (*versource.SaveViewResourceResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/view-resources", c.baseURL)
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

	if resp.StatusCode != http.StatusOK {
		var errorResp http2.ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var viewResourceResp versource.SaveViewResourceResponse
	err = json.NewDecoder(resp.Body).Decode(&viewResourceResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &viewResourceResp, nil
}

func (c *Client) DeleteViewResource(ctx context.Context, req versource.DeleteViewResourceRequest) (*versource.DeleteViewResourceResponse, error) {
	url := fmt.Sprintf("%s/api/v1/view-resources/%d", c.baseURL, req.ViewResourceID)
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
		err := json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var viewResourceResp versource.DeleteViewResourceResponse
	err = json.NewDecoder(resp.Body).Decode(&viewResourceResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &viewResourceResp, nil
}
