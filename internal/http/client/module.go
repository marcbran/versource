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
		var errorResp http2.ErrorResponse
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
		var errorResp http2.ErrorResponse
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
		var errorResp http2.ErrorResponse
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
		var errorResp http2.ErrorResponse
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
		var errorResp http2.ErrorResponse
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
		var errorResp http2.ErrorResponse
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
