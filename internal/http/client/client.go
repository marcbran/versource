package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/marcbran/versource/internal"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(config *internal.Config) internal.Facade {
	baseURL := fmt.Sprintf("%s://%s:%s", config.HTTP.Scheme, config.HTTP.Hostname, config.HTTP.Port)
	if config.HTTP.Hostname == "" {
		baseURL = fmt.Sprintf("%s://localhost:%s", config.HTTP.Scheme, config.HTTP.Port)
	}

	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *Client) Start(ctx context.Context) {
}
