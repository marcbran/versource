package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/marcbran/versource/internal"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(config *internal.Config) *Client {
	baseURL := fmt.Sprintf("%s://%s:%s", config.HTTP.Scheme, config.HTTP.Hostname, config.HTTP.Port)
	if config.HTTP.Hostname == "" {
		baseURL = fmt.Sprintf("%s://localhost:%s", config.HTTP.Scheme, config.HTTP.Port)
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

func (c *Client) CreateModule(ctx context.Context, req internal.CreateModuleRequest) (*internal.CreateModuleResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/modules", c.baseURL)
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

	var moduleResp internal.CreateModuleResponse
	if err := json.NewDecoder(resp.Body).Decode(&moduleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &moduleResp, nil
}

func (c *Client) UpdateModule(ctx context.Context, moduleID uint, req internal.UpdateModuleRequest) (*internal.UpdateModuleResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/modules/%d", c.baseURL, moduleID)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
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

	var moduleResp internal.UpdateModuleResponse
	if err := json.NewDecoder(resp.Body).Decode(&moduleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &moduleResp, nil
}

func (c *Client) DeleteModule(ctx context.Context, moduleID uint) (*internal.DeleteModuleResponse, error) {
	url := fmt.Sprintf("%s/api/v1/modules/%d", c.baseURL, moduleID)
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
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var moduleResp internal.DeleteModuleResponse
	if err := json.NewDecoder(resp.Body).Decode(&moduleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &moduleResp, nil
}

func (c *Client) ListModules(ctx context.Context) (*internal.ListModulesResponse, error) {
	url := fmt.Sprintf("%s/api/v1/modules", c.baseURL)
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
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var modulesResp internal.ListModulesResponse
	if err := json.NewDecoder(resp.Body).Decode(&modulesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &modulesResp, nil
}

func (c *Client) ListComponents(ctx context.Context, req internal.ListComponentsRequest) (*internal.ListComponentsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/components", c.baseURL)

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
		var errorResp ErrorResponse
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

func (c *Client) ListPlans(ctx context.Context) (*internal.ListPlansResponse, error) {
	url := fmt.Sprintf("%s/api/v1/plans", c.baseURL)
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
		var errorResp ErrorResponse
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
		var errorResp ErrorResponse
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
		var errorResp ErrorResponse
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

func (c *Client) ListModuleVersions(ctx context.Context) (*internal.ListModuleVersionsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/module-versions", c.baseURL)
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
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var moduleVersionsResp internal.ListModuleVersionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&moduleVersionsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &moduleVersionsResp, nil
}

func (c *Client) ListModuleVersionsForModule(ctx context.Context, moduleID uint) (*internal.ListModuleVersionsForModuleResponse, error) {
	url := fmt.Sprintf("%s/api/v1/modules/%d/versions", c.baseURL, moduleID)
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
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	var moduleVersionsResp internal.ListModuleVersionsForModuleResponse
	if err := json.NewDecoder(resp.Body).Decode(&moduleVersionsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &moduleVersionsResp, nil
}
