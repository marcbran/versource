//go:build e2e

package tests

import (
	"fmt"

	"github.com/marcbran/versource/internal"
	"github.com/stretchr/testify/require"
)

func (s *Stage) a_changeset_has_been_created(name string) *Stage {
	return s.a_changeset_is_created(name).and().
		the_changeset_creation_has_succeeded()
}

func (s *Stage) a_changeset_is_created(name string) *Stage {
	s.ChangesetName = name
	return s.a_client_command_is_executed("changeset", "create", "--name", name)
}

func (s *Stage) the_changeset_creation_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_changeset_creation_has_failed() *Stage {
	return s.the_command_has_failed()
}

func (s *Stage) the_changeset_has_been_merged() *Stage {
	return s.the_changeset_is_merged().and().
		the_changeset_merge_creation_has_succeeded().and().
		the_changeset_merge_has_succeeded()
}

func (s *Stage) the_changeset_is_merged() *Stage {
	return s.a_changeset_is_merged(s.ChangesetName)
}

func (s *Stage) a_changeset_has_been_merged(changesetName string) *Stage {
	return s.a_changeset_is_merged(changesetName).and().
		the_changeset_merge_creation_has_succeeded().and().
		the_changeset_merge_has_succeeded()
}

func (s *Stage) a_changeset_is_merged(changesetName string) *Stage {
	s.ChangesetName = changesetName
	s.a_client_command_is_executed("changeset", "merge", changesetName)
	if s.LastOutputMap != nil {
		if id, ok := s.LastOutputMap["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				s.MergeID = fmt.Sprintf("%.0f", idFloat)
			}
		}
	}
	return s
}

func (s *Stage) the_changeset_merge_creation_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_changeset_merge_creation_has_failed() *Stage {
	return s.the_command_has_failed()
}

func (s *Stage) the_changeset_merge_has_succeeded() *Stage {
	return s.the_changeset_merge_has_completed("Succeeded")
}

func (s *Stage) the_changeset_merge_has_failed() *Stage {
	return s.the_changeset_merge_has_completed("Failed")
}

func (s *Stage) the_changeset_merge_has_completed(expectedState string) *Stage {
	require.NotEqual(s.t, "", s.MergeID, "No merge id")
	require.NotEqual(s.t, "", s.ChangesetName, "No changeset name")
	s.a_client_command_is_executed("merge", "get", s.MergeID, "--changeset", s.ChangesetName, "--output", "json", "--wait-for-completion")
	require.NotNil(s.t, s.LastOutputMap, "No command output to check")

	state, ok := s.LastOutputMap["state"]
	require.True(s.t, ok, "No state field in command output")

	stateStr, ok := state.(string)
	require.True(s.t, ok, "Merge state is not a string")

	require.Equal(s.t, expectedState, stateStr, "Merge state mismatch")

	return s
}

func (s *Stage) the_changeset_has_been_rebased() *Stage {
	return s.the_changeset_is_rebased().and().
		the_changeset_rebase_creation_has_succeeded().and().
		the_changeset_rebase_has_succeeded()
}

func (s *Stage) the_changeset_is_rebased() *Stage {
	return s.a_changeset_is_rebased(s.ChangesetName)
}

func (s *Stage) a_changeset_has_been_rebased(changesetName string) *Stage {
	return s.a_changeset_is_rebased(changesetName).and().
		the_changeset_rebase_creation_has_succeeded().and().
		the_changeset_rebase_has_succeeded()
}

func (s *Stage) a_changeset_is_rebased(changesetName string) *Stage {
	s.ChangesetName = changesetName
	s.a_client_command_is_executed("changeset", "rebase", changesetName)
	if s.LastOutputMap != nil {
		if id, ok := s.LastOutputMap["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				s.RebaseID = fmt.Sprintf("%.0f", idFloat)
			}
		}
	}
	return s
}

func (s *Stage) the_changeset_rebase_creation_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_changeset_rebase_creation_has_failed() *Stage {
	return s.the_command_has_failed()
}

func (s *Stage) the_changeset_rebase_has_succeeded() *Stage {
	return s.the_changeset_rebase_has_completed("Succeeded")
}

func (s *Stage) the_changeset_rebase_has_failed() *Stage {
	return s.the_changeset_rebase_has_completed("Failed")
}

func (s *Stage) the_changeset_rebase_has_completed(expectedState string) *Stage {
	require.NotEqual(s.t, "", s.RebaseID, "No rebase id")
	require.NotEqual(s.t, "", s.ChangesetName, "No changeset name")
	s.a_client_command_is_executed("rebase", "get", s.RebaseID, "--changeset", s.ChangesetName, "--output", "json", "--wait-for-completion")

	require.NotNil(s.t, s.LastOutputMap, "No command output to check")

	state, ok := s.LastOutputMap["state"]
	require.True(s.t, ok, "No state field in command output")

	stateStr, ok := state.(string)
	require.True(s.t, ok, "Rebase state is not a string")

	require.Equal(s.t, expectedState, stateStr, "Rebase state mismatch")

	return s
}

func (s *Stage) a_changeset_has_been_deleted(changesetName string) *Stage {
	return s.a_changeset_is_deleted(changesetName).and().
		the_changeset_deletion_has_succeeded()
}

func (s *Stage) a_changeset_is_deleted(changesetName string) *Stage {
	return s.a_client_command_is_executed("changeset", "delete", changesetName)
}

func (s *Stage) the_changeset_deletion_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_changeset_deletion_has_failed() *Stage {
	return s.the_command_has_failed()
}

func (s *Stage) the_changeset_changes_are_listed() *Stage {
	return s.a_client_command_is_executed("changeset", "change", "list", "--changeset", s.ChangesetName)
}

func (s *Stage) there_are_changes(expectedCount int) *Stage {
	require.NotNil(s.t, s.LastOutputArray, "No command output to check")

	require.Equal(s.t, expectedCount, len(s.LastOutputArray), "Unexpected number of changes")

	return s
}

func (s *Stage) the_changeset_nth_change_is_set_to(index int, changeType internal.ChangeType) *Stage {
	require.NotNil(s.t, s.LastOutputArray, "No command output to check")
	require.True(s.t, index < len(s.LastOutputArray), "Index %d is out of bounds for changes array of length %d", index, len(s.LastOutputArray))

	change, ok := s.LastOutputArray[index].(map[string]any)
	require.True(s.t, ok, "Change at index %d is not a map", index)

	actualType, ok := change["changeType"]
	require.True(s.t, ok, "No type field in change at index %d", index)

	actualTypeStr, ok := actualType.(string)
	require.True(s.t, ok, "Type field is not a string in change at index %d", index)

	require.Equal(s.t, string(changeType), actualTypeStr, "Change type mismatch at index %d", index)

	return s
}

func (s *Stage) all_changeset_plans_have_succeeded() *Stage {
	return s.all_changeset_plans_have_completed("Succeeded")
}

func (s *Stage) all_changeset_plans_have_failed() *Stage {
	return s.all_changeset_plans_have_completed("Failed")
}

func (s *Stage) all_changeset_plans_have_completed(expectedState string) *Stage {
	require.NotEqual(s.t, "", s.ChangesetName, "No changeset name")
	s.a_client_command_is_executed("changeset", "change", "list", "--changeset", s.ChangesetName, "--wait-for-completion", "--output", "json")

	require.NotNil(s.t, s.LastOutputArray, "No command output to check")

	for i, change := range s.LastOutputArray {
		changeMap, ok := change.(map[string]any)
		require.True(s.t, ok, "Change at index %d is not a map", i)

		plan, ok := changeMap["plan"]
		if !ok {
			require.Fail(s.t, "No plan field in change at index %d", i)
			continue
		}

		if plan == nil {
			require.Fail(s.t, "Plan is nil in change at index %d", i)
			continue
		}

		planMap, ok := plan.(map[string]any)
		require.True(s.t, ok, "Plan at index %d is not a map", i)

		state, ok := planMap["state"]
		require.True(s.t, ok, "No state field in plan at index %d", i)

		stateStr, ok := state.(string)
		require.True(s.t, ok, "Plan state is not a string at index %d", i)

		require.Equal(s.t, expectedState, stateStr, "Plan state mismatch at index %d", i)
	}

	return s
}
