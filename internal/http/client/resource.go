package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	http2 "github.com/marcbran/versource/internal/http/server"
	"github.com/marcbran/versource/pkg/versource"
)

func (c *Client) ListResources(ctx context.Context, req versource.ListResourcesRequest) (*versource.ListResourcesResponse, error) {
	url := fmt.Sprintf("%s/api/v1/resources", c.baseURL)
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

	var resourcesResp versource.ListResourcesResponse
	err = json.NewDecoder(resp.Body).Decode(&resourcesResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resourcesResp, nil
}
