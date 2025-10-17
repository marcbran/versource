//go:build e2e

package tests

import (
	"fmt"

	"github.com/marcbran/versource/pkg/versource"
)

func (s *Stage) an_existing_module_has_been_created() *Stage {
	return s.a_module_is_created(
		"jsonnet",
		"https://github.com/marcbran/versource/tests/modules/jsonnet",
		"c550182f94d192177a47c5b29c6b06be5ddb6bb3",
	).and().the_module_creation_has_succeeded()
}

func (s *Stage) the_resources_module_has_been_created() *Stage {
	return s.a_module_is_created(
		"resources",
		"https://github.com/marcbran/versource/tests/modules/resources",
		"847f622fede575bac37247af57d9bd0494f7be52",
	).and().the_module_creation_has_succeeded()
}

func (s *Stage) a_module_that_will_fail_on_apply_has_been_created() *Stage {
	return s.a_module_is_created(
		"fail",
		"https://github.com/marcbran/versource/tests/modules/fail",
		"81840a7d4b0cda6ebe61d5476be8f7334e86b6fb",
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
	response := unmarshalResponse[versource.CreateModuleResponse](s.t, s.LastOutput)
	s.ModuleID = fmt.Sprintf("%d", response.Module.ID)
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
