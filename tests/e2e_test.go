//go:build e2e
// +build e2e

package tests

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	runDockerCompose("down")
	runDockerCompose("build", "--no-cache")
	runDockerCompose("up", "-d")

	code := m.Run()

	runDockerCompose("down")

	os.Exit(code)
}

func TestCreateChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_changeset_is_created("test1")

	then.
		the_changeset_is_created_successfully()
}

func TestCreateChangesetWithInvalidName(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_changeset_is_created(".invalid-name")

	then.
		the_changeset_creation_has_failed()
}

func TestCreateChangesetWithDuplicateName(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1")

	when.
		a_changeset_is_created("test1")

	then.
		the_changeset_creation_has_failed()
}

func TestMergeChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1")

	when.
		the_changeset_is_merged()

	then.
		the_changeset_is_merged_successfully()
}

func TestCreateChangesetAfterMerge(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		the_changeset_has_been_merged()

	when.
		a_changeset_is_created("test1")

	then.
		the_changeset_is_created_successfully()
}

func TestCreateModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1")

	when.
		a_module_is_created("test1", "test-module", "1.0.0", `{"key": "value"}`)

	then.
		the_module_is_created_successfully()
}

func TestCreateModuleWithoutChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("nonexistent", "test-module", "1.0.0", `{"key": "value"}`)

	then.
		the_module_is_created_successfully()
}

func TestCreateModuleWithInvalidVariables(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1")

	when.
		a_module_is_created("test1", "test-module", "1.0.0", `{"invalid": json`)

	then.
		the_module_creation_has_failed()
}

func TestCreateModuleWithEmptySource(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1")

	when.
		a_module_is_created_with_empty_source("test1", "1.0.0", `{"key": "value"}`)

	then.
		the_module_creation_has_failed()
}

func TestUpdateModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("test1", "test-module", "1.0.0", `{"key": "value"}`)

	when.
		the_module_is_updated("test1", "1", "updated-module", "2.0.0", `{"key": "updated"}`)

	then.
		the_module_is_updated_successfully()
}

func TestUpdateModuleWithNonexistentID(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1")

	when.
		the_module_is_updated("test1", "999", "updated-module", "2.0.0", `{"key": "updated"}`)

	then.
		the_module_update_has_failed()
}

func TestUpdateModuleWithInvalidChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("test1", "test-module", "1.0.0", `{"key": "value"}`)

	when.
		the_module_is_updated("nonexistent", "1", "updated-module", "2.0.0", `{"key": "updated"}`)

	then.
		the_module_update_has_failed()
}

func TestUpdateModuleWithInvalidVariables(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("test1", "test-module", "1.0.0", `{"key": "value"}`)

	when.
		the_module_is_updated("test1", "1", "updated-module", "2.0.0", `{"invalid": json`)

	then.
		the_module_update_has_failed()
}

func TestUpdateModuleWithNoFields(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("test1", "test-module", "1.0.0", `{"key": "value"}`)

	when.
		the_module_is_updated_with_no_fields("test1", "1")

	then.
		the_module_update_has_failed()
}

func TestCreatePlan(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("test1", "test-module", "1.0.0", `{"key": "value"}`)

	when.
		a_plan_is_created("test1", "1")

	then.
		the_plan_is_created_successfully()
}

func TestCreatePlanWithNonexistentModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1")

	when.
		a_plan_is_created("test1", "999")

	then.
		the_plan_creation_has_failed()
}

func TestCreatePlanWithNonexistentChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("test1", "test-module", "1.0.0", `{"key": "value"}`)

	when.
		a_plan_is_created("nonexistent", "1")

	then.
		the_plan_creation_has_failed()
}

func TestCreatePlanWithoutChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("test1", "test-module", "1.0.0", `{"key": "value"}`)

	when.
		a_plan_is_created_without_changeset("1")

	then.
		the_plan_creation_has_failed()
}

func TestCreatePlanWithoutModuleID(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("test1", "test-module", "1.0.0", `{"key": "value"}`)

	when.
		a_plan_is_created_without_module_id("test1")

	then.
		the_plan_creation_has_failed()
}

func TestMergeChangesetWithModules(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("test1", "test-module", "1.0.0", `{"key": "value"}`).and().
		a_plan_has_been_created("test1", "1")

	when.
		the_changeset_is_merged()

	then.
		the_changeset_is_merged_successfully()
}

func TestCreateMultipleModulesInSameChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1")

	when.
		a_module_is_created("test1", "module1", "1.0.0", `{"key1": "value1"}`).and().
		a_module_is_created("test1", "module2", "2.0.0", `{"key2": "value2"}`)

	then.
		both_modules_are_created_successfully()
}

func TestCreateModulesInDifferentChangesets(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("changeset1").and().
		a_changeset_has_been_created("changeset2")

	when.
		a_module_is_created("changeset1", "module1", "1.0.0", `{"key1": "value1"}`).and().
		a_module_is_created("changeset2", "module2", "2.0.0", `{"key2": "value2"}`)

	then.
		both_modules_are_created_successfully()
}

func TestMergeChangesetWithMultipleModules(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("test1", "module1", "1.0.0", `{"key1": "value1"}`).and().
		a_module_has_been_created("test1", "module2", "2.0.0", `{"key2": "value2"}`).and().
		a_plan_has_been_created("test1", "1").and().
		a_plan_has_been_created("test1", "2")

	when.
		the_changeset_is_merged()

	then.
		the_changeset_is_merged_successfully()
}

func TestCreateChangesetWithSpecialCharacters(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_changeset_is_created("test-changeset-123")

	then.
		the_changeset_is_created_successfully()
}

func TestCreateModuleWithComplexVariables(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1")

	when.
		a_module_is_created("test1", "complex-module", "1.0.0", `{"nested": {"key": "value"}, "array": [1, 2, 3], "boolean": true, "number": 42}`)

	then.
		the_module_is_created_successfully()
}
