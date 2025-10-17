//go:build e2e

package tests

import (
	"fmt"

	"github.com/marcbran/versource/pkg/versource"
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
	response := unmarshalResponse[versource.CreateMergeResponse](s.t, s.LastOutput)
	s.MergeID = fmt.Sprintf("%d", response.Merge.ID)
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
	require.NotEqual(s.t, "", s.LastOutput, "No command output to check")

	response := unmarshalResponse[versource.GetMergeResponse](s.t, s.LastOutput)

	require.Equal(s.t, expectedState, string(response.Merge.State), "Merge state mismatch")

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
	response := unmarshalResponse[versource.CreateRebaseResponse](s.t, s.LastOutput)
	s.RebaseID = fmt.Sprintf("%d", response.Rebase.ID)
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

	require.NotEqual(s.t, "", s.LastOutput, "No command output to check")

	response := unmarshalResponse[versource.GetRebaseResponse](s.t, s.LastOutput)

	require.Equal(s.t, expectedState, string(response.Rebase.State), "Rebase state mismatch")

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
	changes := unmarshalArray[versource.ComponentChange](s.t, s.LastOutput)
	require.Equal(s.t, expectedCount, len(changes), "Unexpected number of changes")
	return s
}

func (s *Stage) the_changeset_nth_change_is_set_to(index int, changeType versource.ChangeType) *Stage {
	changes := unmarshalArray[versource.ComponentChange](s.t, s.LastOutput)
	require.True(s.t, index < len(changes), "Index %d is out of bounds for changes array of length %d", index, len(changes))
	require.Equal(s.t, changeType, changes[index].ChangeType, "Change type mismatch at index %d", index)
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

	changes := unmarshalArray[versource.ComponentChange](s.t, s.LastOutput)

	for i, change := range changes {
		require.NotNil(s.t, change.Plan, "Plan is nil in change at index %d", i)
		require.Equal(s.t, expectedState, string(change.Plan.State), "Plan state mismatch at index %d", i)
	}

	return s
}
