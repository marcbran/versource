//go:build e2e && (all || changeset)

package tests

import "testing"

func TestCreateChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset()

	when.
		a_changeset_is_created("test1")

	then.
		the_changeset_creation_has_succeeded()
}

func TestCreateChangesetWithInvalidName(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset()

	when.
		a_changeset_is_created(".invalid-name")

	then.
		the_changeset_creation_has_failed()
}

func TestCreateChangesetWithDuplicateName(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_changeset_has_been_created("test1")

	when.
		a_changeset_is_created("test1")

	then.
		the_changeset_creation_has_failed()
}

func TestMergeChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_changeset_has_been_created("test1")

	when.
		the_changeset_is_merged()

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_succeeded()
}

func TestCreateChangesetAfterMerge(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_changeset_has_been_created("test1").and().
		the_changeset_has_been_merged()

	when.
		a_changeset_is_created("test1")

	then.
		the_changeset_creation_has_failed()
}
func TestCreateChangesetWithSpecialCharacters(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_changeset_is_created("test-changeset-123")

	then.
		the_changeset_creation_has_succeeded()
}

func TestMergeChangesetWithComponent(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "component1"}`).and().
		the_plan_has_succeeded()

	when.
		the_changeset_is_merged()

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_succeeded()
}

func TestMergeChangesetWithComponentUpdate(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "component1"}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged().and().
		a_changeset_has_been_created("test2").and().
		the_component_has_been_updated_in_the_changeset(`{"name": "value2"}`).and().
		the_plan_has_succeeded()

	when.
		the_changeset_is_merged()

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_succeeded()
}

func TestMergeChangesetWithNonexistingModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_non_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "component1"}`).and().
		the_plan_has_failed()

	when.
		the_changeset_is_merged()

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_failed()
}

func TestMergeChangesetWithInvalidComponent(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{}`).and().
		the_plan_has_failed()

	when.
		the_changeset_is_merged()

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_failed()
}

func TestMergeChangesetWithMultipleComponents(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "value1"}`).and().
		the_plan_has_succeeded().and().
		a_component_has_been_created_for_the_module_and_changeset("component2", `{"name": "value2"}`).and().
		the_plan_has_succeeded()

	when.
		the_changeset_is_merged()

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_succeeded()
}

func TestMergeChangesetWithComponentConflicts(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "value1"}`).and().
		the_plan_has_succeeded().and().
		a_changeset_has_been_created("test2").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "value2"}`).and().
		the_plan_has_succeeded().and().
		a_changeset_has_been_merged("test1")

	when.
		a_changeset_is_merged("test2")

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_failed()
}

func TestMergeChangesetWithComponentUpdateConflicts(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "value"}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged().and().
		a_changeset_has_been_created("test1").and().
		the_component_has_been_updated_in_a_changeset("test1", `{"name": "value2"}`).and().
		the_plan_has_succeeded().and().
		a_changeset_has_been_created("test2").and().
		the_component_has_been_updated_in_a_changeset("test2", `{"name": "value2"}`).and().
		the_plan_has_succeeded().and().
		a_changeset_has_been_merged("test1")

	when.
		a_changeset_is_merged("test2")

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_failed()
}
