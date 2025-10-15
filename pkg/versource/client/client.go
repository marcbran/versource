package client

import (
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/pkg/versource"
)

func New(config *versource.Config) versource.Facade {
	return client.New(config)
}
