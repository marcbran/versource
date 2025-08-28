//go:build e2e

package tests

import (
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (s *Stage) a_component_has_been_created(changeset, variables string) *Stage {
	return s.a_component_is_created_for_the_module(changeset, variables).and().
		the_component_is_created_successfully()
}

func (s *Stage) a_component_is_created_for_the_module(changeset, variables string) *Stage {
	s.ChangesetName = changeset
	args := []string{"component", "create", "--changeset", changeset, "--module-id", s.ModuleID, "--variables", variables}
	s.execCommand(args...)
	if s.LastOutputMap != nil {
		if id, ok := s.LastOutputMap["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				s.ComponentID = fmt.Sprintf("%.0f", idFloat)
			}
		}
	}
	return s
}

func (s *Stage) a_component_is_created_with_empty_module_id(changeset, variables string) *Stage {
	s.ChangesetName = changeset
	args := []string{"component", "create", "--changeset", changeset, "--module-id", "", "--variables", variables}
	return s.execCommand(args...)
}

func (s *Stage) the_component_is_created_successfully() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_component_creation_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *Stage) both_components_are_created_successfully() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_component_is_updated(changeset, componentID, moduleID, variables string) *Stage {
	s.ChangesetName = changeset
	s.ComponentID = componentID
	args := []string{"component", "update", componentID, "--changeset", changeset}
	if moduleID != "" {
		args = append(args, "--module-id", moduleID)
	}
	if variables != "" {
		args = append(args, "--variables", variables)
	}
	return s.execCommand(args...)
}

func (s *Stage) the_component_is_updated_with_no_fields(changeset, componentID string) *Stage {
	s.ChangesetName = changeset
	s.ComponentID = componentID
	return s.execCommand("component", "update", componentID, "--changeset", changeset)
}

func (s *Stage) the_component_is_updated_successfully() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_component_update_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}
