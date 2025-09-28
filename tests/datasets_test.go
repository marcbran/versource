//go:build e2e && datasets

package tests

import (
	"testing"
)

func TestDatasetBlankInstance(t *testing.T) {
	given, _, then := scenario(t)

	given.
		a_clean_slate()

	then.
		the_state_is_stored_in_the_dataset(blank_instance)
}

func TestDatasetModuleAndChangeset(t *testing.T) {
	given, _, then := scenario(t)

	given.
		a_clean_slate().and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1")

	then.
		the_state_is_stored_in_the_dataset(module_and_changeset)
}

func TestDatasetDeletedComponent(t *testing.T) {
	given, _, then := scenario(t)

	given.
		a_clean_slate().and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component", `{"name": "value"}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged().and().
		a_changeset_has_been_created("test2").and().
		the_component_has_been_deleted().and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged()

	then.
		the_state_is_stored_in_the_dataset(deleted_component)
}

func TestDatasetChangesetAndNewComponent(t *testing.T) {
	given, _, then := scenario(t)

	given.
		a_clean_slate().and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "component1"}`).and().
		the_component_has_been_updated_in_the_changeset(`{"name": "component11"}`)

	then.
		the_state_is_stored_in_the_dataset(changeset_and_new_component)
}

func TestDatasetTwoChangesetWithChanges(t *testing.T) {
	given, _, then := scenario(t)

	given.
		a_clean_slate().and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "value"}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged().and().
		a_changeset_has_been_created("changeset1").and().
		the_component_has_been_updated_in_a_changeset("changeset1", `{"name": "value1"}`).and().
		the_plan_has_succeeded().and().
		a_changeset_has_been_created("changeset2").and().
		the_component_has_been_updated_in_a_changeset("changeset2", `{"name": "value2"}`).and().
		the_plan_has_succeeded()

	then.
		the_state_is_stored_in_the_dataset(two_changesets_with_changes)
}
