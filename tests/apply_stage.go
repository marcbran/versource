//go:build e2e

package tests

import (
	"github.com/stretchr/testify/require"
)

func (s *Stage) all_applies_have_succeeded() *Stage {
	return s.all_applies_have_completed("Succeeded")
}

func (s *Stage) all_applies_have_failed() *Stage {
	return s.all_applies_have_completed("Failed")
}

func (s *Stage) all_applies_have_completed(expectedState string) *Stage {
	s.a_client_command_is_executed("apply", "list", "--wait-for-completion", "--output", "json")

	require.NotNil(s.t, s.LastOutputArray, "No command output to check")

	for i, apply := range s.LastOutputArray {
		applyMap, ok := apply.(map[string]any)
		require.True(s.t, ok, "Apply at index %d is not a map", i)

		state, ok := applyMap["state"]
		require.True(s.t, ok, "No state field in apply at index %d", i)

		stateStr, ok := state.(string)
		require.True(s.t, ok, "Apply state is not a string at index %d", i)

		require.Equal(s.t, expectedState, stateStr, "Apply state mismatch at index %d", i)
	}

	return s
}
