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

func (c *Client) GetPlan(ctx context.Context, req internal.GetPlanRequest) (*internal.GetPlanResponse, error) {
	url := fmt.Sprintf("%s/api/v1/plans/%d", c.baseURL, req.PlanID)
	if req.ChangesetName != nil {
		url = fmt.Sprintf("%s/api/v1/changesets/%s/plans/%d", c.baseURL, *req.ChangesetName, req.PlanID)
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

	var planResp internal.GetPlanResponse
	if err := json.NewDecoder(resp.Body).Decode(&planResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &planResp, nil
}

func (c *Client) GetPlanLog(ctx context.Context, req internal.GetPlanLogRequest) (*internal.GetPlanLogResponse, error) {
	url := fmt.Sprintf("%s/api/v1/plans/%d/logs", c.baseURL, req.PlanID)
	if req.ChangesetName != nil {
		url = fmt.Sprintf("%s/api/v1/changesets/%s/plans/%d/logs", c.baseURL, *req.ChangesetName, req.PlanID)
	}
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

	return &internal.GetPlanLogResponse{
		Content: resp.Body,
	}, nil
}

func (c *Client) ListPlans(ctx context.Context, req internal.ListPlansRequest) (*internal.ListPlansResponse, error) {
	url := fmt.Sprintf("%s/api/v1/plans", c.baseURL)
	if req.ChangesetName != "" {
		url = fmt.Sprintf("%s/api/v1/changesets/%s/plans", c.baseURL, req.ChangesetName)
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

	var plansResp internal.ListPlansResponse
	if err := json.NewDecoder(resp.Body).Decode(&plansResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &plansResp, nil
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
		var errorResp http2.ErrorResponse
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

func (c *Client) RunPlan(ctx context.Context, planID uint) error {
	req := map[string]uint{"plan_id": planID}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/plans/run", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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
