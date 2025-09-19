//go:build e2e

package tests

import (
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (s *Stage) a_changeset_has_been_created(name string) *Stage {
	return s.a_changeset_is_created(name).and().
		the_changeset_creation_has_succeeded()
}

func (s *Stage) a_changeset_is_created(name string) *Stage {
	s.ChangesetName = name
	return s.execCommand("changeset", "create", "--name", name)
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
	s.execCommand("changeset", "merge", changesetName)
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
	s.execCommand("merge", "get", s.MergeID, "--changeset", s.ChangesetName, "--output", "json", "--wait-for-completion")

	assert.NotNil(s.t, s.LastOutputMap, "No command output to check")

	state, ok := s.LastOutputMap["state"]
	assert.True(s.t, ok, "No state field in command output")

	stateStr, ok := state.(string)
	assert.True(s.t, ok, "Merge state is not a string")

	assert.Equal(s.t, expectedState, stateStr, "Merge state mismatch")

	return s
}
