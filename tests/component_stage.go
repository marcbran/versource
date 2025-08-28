//go:build e2e

package tests

import (
	"fmt"
)

func (s *Stage) a_component_has_been_created(variables string) *Stage {
	return s.a_component_is_created_for_the_module_and_changeset(variables).and().
		the_component_creation_has_succeeded()
}

func (s *Stage) a_component_is_created_for_the_module(changeset, variables string) *Stage {
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

func (s *Stage) a_component_is_created_for_the_module_and_changeset(variables string) *Stage {
	args := []string{"component", "create", "--changeset", s.ChangesetName, "--module-id", s.ModuleID, "--variables", variables}
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

func (s *Stage) the_component_creation_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_component_creation_has_failed() *Stage {
	return s.the_command_has_failed()
}

func (s *Stage) both_component_creations_have_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_component_is_updated(variables string) *Stage {
	return s.execCommand("component", "update", s.ComponentID, "--changeset", s.ChangesetName, "--variables", variables)
}

func (s *Stage) a_component_is_updated_for_the_changeset(componentID, variables string) *Stage {
	return s.execCommand("component", "update", componentID, "--changeset", s.ChangesetName, "--variables", variables)
}

func (s *Stage) a_component_is_updated(componentID, changeset, variables string) *Stage {
	return s.execCommand("component", "update", componentID, "--changeset", changeset, "--variables", variables)
}

func (s *Stage) the_component_is_updated_with_no_fields() *Stage {
	return s.execCommand("component", "update", s.ComponentID, "--changeset", s.ChangesetName)
}

func (s *Stage) the_component_update_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_component_update_has_failed() *Stage {
	return s.the_command_has_failed()
}
