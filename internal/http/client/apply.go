package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/marcbran/versource/internal"
	http2 "github.com/marcbran/versource/internal/http/server"
)

func (c *Client) ListApplies(ctx context.Context) (*internal.ListAppliesResponse, error) {
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
