//go:build e2e

package tests

import (
	"fmt"
)

func (s *Stage) an_existing_module_has_been_created() *Stage {
	return s.a_module_is_created(
		"jsonnet",
		"https://github.com/marcbran/versource/tests/modules/jsonnet",
		"199fa5704319b958d47b791f063729a83ec83f15",
	).and().the_module_creation_has_succeeded()
}

func (s *Stage) a_non_existing_module_has_been_created() *Stage {
	return s.a_module_is_created(
		"not-an-existing-module",
		"https://github.com/marcbran/versource/tests/modules/nothing",
		"invalid",
	).and().the_module_creation_has_succeeded()
}

func (s *Stage) a_module_has_been_created(name, source, version string) *Stage {
	return s.a_module_is_created(name, source, version).and().
		the_module_creation_has_succeeded()
}

func (s *Stage) a_module_is_created(name, source, version string) *Stage {
	args := []string{"module", "create", "--name", name, "--source", source}
	if version != "" {
		args = append(args, "--version", version)
	}
	s.a_client_command_is_executed(args...)
	if s.LastOutputMap != nil {
		if id, ok := s.LastOutputMap["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				s.ModuleID = fmt.Sprintf("%.0f", idFloat)
			}
		}
	}
	return s
}

func (s *Stage) the_module_creation_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_module_creation_has_failed() *Stage {
	return s.the_command_has_failed()
}

func (s *Stage) the_module_is_updated(version string) *Stage {
	return s.a_module_is_updated(s.ModuleID, version)
}

func (s *Stage) a_module_is_updated(moduleID, version string) *Stage {
	args := []string{"module", "update", moduleID, "--version", version}
	return s.a_client_command_is_executed(args...)
}

func (s *Stage) the_module_update_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_module_update_has_failed() *Stage {
	return s.the_command_has_failed()
}

func (s *Stage) the_module_is_deleted() *Stage {
	return s.a_module_is_deleted(s.ModuleID)
}

func (s *Stage) a_module_is_deleted(moduleID string) *Stage {
	args := []string{"module", "delete", moduleID}
	return s.a_client_command_is_executed(args...)
}

func (s *Stage) the_module_deletion_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_module_deletion_has_failed() *Stage {
	return s.the_command_has_failed()
}
