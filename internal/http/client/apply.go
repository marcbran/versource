package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/marcbran/versource/internal"
	http2 "github.com/marcbran/versource/internal/http/server"
)

func (c *Client) GetApply(ctx context.Context, req internal.GetApplyRequest) (*internal.GetApplyResponse, error) {
	url := fmt.Sprintf("%s/api/v1/applies/%d", c.baseURL, req.ApplyID)
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

	var applyResp internal.GetApplyResponse
	if err := json.NewDecoder(resp.Body).Decode(&applyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &applyResp, nil
}

func (c *Client) GetApplyLog(ctx context.Context, req internal.GetApplyLogRequest) (*internal.GetApplyLogResponse, error) {
	url := fmt.Sprintf("%s/api/v1/applies/%d/logs", c.baseURL, req.ApplyID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		var errorResp http2.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	return &internal.GetApplyLogResponse{
		Content: resp.Body,
	}, nil
}

func (c *Client) ListApplies(ctx context.Context, req internal.ListAppliesRequest) (*internal.ListAppliesResponse, error) {
	url := fmt.Sprintf("%s/api/v1/applies", c.baseURL)
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

	var appliesResp internal.ListAppliesResponse
	if err := json.NewDecoder(resp.Body).Decode(&appliesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &appliesResp, nil
}

func (c *Client) RunApply(ctx context.Context, applyID uint) error {
	url := fmt.Sprintf("%s/api/v1/applies/%d/run", c.baseURL, applyID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp http2.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return fmt.Errorf("failed to decode error response: %w", err)
		}
		return fmt.Errorf("server error: %s", errorResp.Message)
	}

	return nil
}
