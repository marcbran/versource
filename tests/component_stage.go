//go:build e2e

package tests

import (
	"fmt"
)

func (s *Stage) a_component_has_been_created_for_the_module_and_changeset(name, variables string) *Stage {
	return s.a_component_is_created_for_the_module_and_changeset(name, variables).and().
		the_component_creation_has_succeeded()
}

func (s *Stage) a_component_is_created_for_the_module(changeset, name, variables string) *Stage {
	return s.a_component_is_created(changeset, s.ModuleID, name, variables)
}

func (s *Stage) a_component_is_created_for_the_module_and_changeset(name, variables string) *Stage {
	return s.a_component_is_created(s.ChangesetName, s.ModuleID, name, variables)
}

func (s *Stage) a_component_is_created(changeset, moduleID, name, variables string) *Stage {
	args := []string{"component", "create", "--name", name, "--changeset", changeset, "--module-id", moduleID}
	args = append(args, s.parseVariablesToArgs(variables)...)
	s.execCommand(args...)
	if s.LastOutputMap != nil {
		if id, ok := s.LastOutputMap["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				s.ComponentID = fmt.Sprintf("%.0f", idFloat)
			}
		}
		if id, ok := s.LastOutputMap["plan_id"]; ok {
			if idFloat, ok := id.(float64); ok {
				s.PlanID = fmt.Sprintf("%.0f", idFloat)
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
	args := []string{"component", "update", s.ComponentID, "--changeset", s.ChangesetName}
	args = append(args, s.parseVariablesToArgs(variables)...)
	return s.execCommand(args...)
}

func (s *Stage) a_component_is_updated_for_the_changeset(componentID, variables string) *Stage {
	return s.a_component_is_updated(componentID, s.ChangesetName, variables)
}

func (s *Stage) a_component_is_updated(componentID, changeset, variables string) *Stage {
	args := []string{"component", "update", componentID, "--changeset", changeset}
	args = append(args, s.parseVariablesToArgs(variables)...)
	return s.execCommand(args...)
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
