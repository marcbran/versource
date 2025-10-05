//go:build e2e

package tests

import (
	"github.com/stretchr/testify/require"
)

func (s *Stage) the_resources_are_listed() *Stage {
	return s.a_client_command_is_executed("resource", "list", "--output", "json")
}

func (s *Stage) there_are_resources(expectedCount int) *Stage {
	require.NotNil(s.t, s.LastOutputArray, "No command output to check")
	require.Equal(s.t, expectedCount, len(s.LastOutputArray), "Unexpected number of resources")
	return s
}
