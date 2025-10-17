//go:build e2e

package tests

import (
	"github.com/marcbran/versource/pkg/versource"
	"github.com/stretchr/testify/require"
)

func (s *Stage) the_resources_are_listed() *Stage {
	return s.a_client_command_is_executed("resource", "list", "--output", "json")
}

func (s *Stage) there_are_resources(expectedCount int) *Stage {
	resources := unmarshalArray[versource.Resource](s.t, s.LastOutput)
	require.Equal(s.t, expectedCount, len(resources), "Unexpected number of resources")
	return s
}
