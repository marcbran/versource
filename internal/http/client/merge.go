package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	http2 "github.com/marcbran/versource/internal/http/server"
	"github.com/marcbran/versource/pkg/versource"
)

func (c *Client) GetMerge(ctx context.Context, req versource.GetMergeRequest) (*versource.GetMergeResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets/%s/merges/%d", c.baseURL, req.ChangesetName, req.MergeID)
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

	var mergeResp versource.GetMergeResponse
	err = json.NewDecoder(resp.Body).Decode(&mergeResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &mergeResp, nil
}

func (c *Client) ListMerges(ctx context.Context, req versource.ListMergesRequest) (*versource.ListMergesResponse, error) {
	url := fmt.Sprintf("%s/api/v1/changesets/%s/merges", c.baseURL, req.ChangesetName)
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

	var mergesResp versource.ListMergesResponse
	err = json.NewDecoder(resp.Body).Decode(&mergesResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &mergesResp, nil
}
