//go:build e2e

package tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
)

func (s *Stage) a_plan_has_been_created(changeset, componentID string) *Stage {
	return s.a_plan_is_created(changeset, componentID).and().
		the_plan_is_created_successfully()
}

func (s *Stage) a_plan_is_created(changeset, componentID string) *Stage {
	s.ChangesetName = changeset
	s.ComponentID = componentID
	return s.execCommand("plan", "--component-id", componentID, "--changeset", changeset)
}

func (s *Stage) a_plan_is_created_without_changeset(componentID string) *Stage {
	s.ComponentID = componentID
	return s.execCommand("plan", "--component-id", componentID)
}

func (s *Stage) a_plan_is_created_without_component_id(changeset string) *Stage {
	s.ChangesetName = changeset
	return s.execCommand("plan", "--changeset", changeset)
}

func (s *Stage) the_plan_is_created_successfully() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_plan_creation_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}
