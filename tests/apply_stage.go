//go:build e2e

package tests

import (
	"github.com/marcbran/versource/pkg/versource"
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

	applies := unmarshalArray[versource.Apply](s.t, s.LastOutput)

	for i, apply := range applies {
		require.Equal(s.t, expectedState, string(apply.State), "Apply state mismatch at index %d", i)
	}

	return s
}
