//go:build e2e

package tests

import (
	"github.com/stretchr/testify/require"
)

func (s *Stage) a_plan_has_been_created() *Stage {
	return s.a_plan_is_created_for_the_changeset_and_component().and().
		the_plan_creation_has_succeeded()
}

func (s *Stage) a_plan_is_created_for_the_changeset_and_component() *Stage {
	return s.a_plan_is_created(s.ChangesetName, s.ComponentID)
}

func (s *Stage) a_plan_is_created_for_the_changeset(componentID string) *Stage {
	return s.a_plan_is_created(s.ChangesetName, componentID)
}

func (s *Stage) a_plan_is_created_for_the_component(changeset string) *Stage {
	return s.a_plan_is_created(changeset, s.ComponentID)
}

func (s *Stage) a_plan_is_created(changeset, componentID string) *Stage {
	return s.a_client_command_is_executed("component", "plan", componentID, "--changeset", changeset)
}

func (s *Stage) a_plan_is_created_without_changeset() *Stage {
	return s.a_client_command_is_executed("component", "plan", s.ComponentID)
}

func (s *Stage) the_plan_creation_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_plan_creation_has_failed() *Stage {
	return s.the_command_has_failed()
}

func (s *Stage) the_plan_has_succeeded() *Stage {
	return s.the_plan_has_completed("Succeeded")
}

func (s *Stage) the_plan_has_failed() *Stage {
	return s.the_plan_has_completed("Failed")
}

func (s *Stage) the_plan_has_completed(expectedState string) *Stage {
	require.NotEqual(s.t, "", s.PlanID, "No plan id")
	require.NotEqual(s.t, "", s.ChangesetName, "No changeset name")
	s.a_client_command_is_executed("plan", "get", s.PlanID, "--changeset", s.ChangesetName, "--wait-for-completion")
	require.NotNil(s.t, s.LastOutputMap, "No command output to check")

	state, ok := s.LastOutputMap["state"]
	require.True(s.t, ok, "No state field in command output")

	stateStr, ok := state.(string)
	require.True(s.t, ok, "Plan state is not a string")

	require.Equal(s.t, expectedState, stateStr, "Plan state mismatch")

	return s
}
