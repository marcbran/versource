//go:build e2e && (all || changeset)

package tests

import (
	"testing"

	"github.com/marcbran/versource/pkg/versource"
)

func TestCreateChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance)

	when.
		a_changeset_is_created("changeset1")

	then.
		the_changeset_creation_has_succeeded()
}

func TestCreateChangesetWithInvalidName(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance)

	when.
		a_changeset_is_created(".invalid-name")

	then.
		the_changeset_creation_has_failed()
}

func TestCreateChangesetWithDuplicateName(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_changeset_has_been_created("changeset1")

	when.
		a_changeset_is_created("changeset1")

	then.
		the_changeset_creation_has_failed()
}

func TestMergeChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_changeset_has_been_created("changeset1")

	when.
		the_changeset_is_merged()

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_succeeded()
}

func TestCreateChangesetAfterMerge(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_changeset_has_been_created("changeset1").and().
		the_changeset_has_been_merged()

	when.
		a_changeset_is_created("changeset1")

	then.
		the_changeset_creation_has_failed()
}
func TestCreateChangesetWithSpecialCharacters(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance)

	when.
		a_changeset_is_created("test-changeset-123")

	then.
		the_changeset_creation_has_succeeded()
}

func TestDeleteChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_changeset_has_been_created("changeset1")

	when.
		a_changeset_is_deleted("changeset1")

	then.
		the_changeset_deletion_has_succeeded()
}

func TestDeleteChangesetWithInvalidName(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance)

	when.
		a_changeset_is_deleted("nonexistent")

	then.
		the_changeset_deletion_has_failed()
}

func TestListChangesInChangesetWithComponent(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "component1"}`)

	when.
		the_changeset_changes_are_listed()

	then.
		there_are_changes(1)
}

func TestListCreatedChangesInChangesetWithComponent(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "component1"}`).and().
		the_component_has_been_updated_in_the_changeset(`{"name": "component11"}`)

	when.
		the_changeset_changes_are_listed()

	then.
		the_changeset_nth_change_is_set_to(0, versource.ChangeTypeCreated)
}

func TestMergeChangesetWithComponent(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
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
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "component1"}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged().and().
		a_changeset_has_been_created("changeset2").and().
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
		the_dataset(blank_instance).and().
		a_non_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
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
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
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
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
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
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "value1"}`).and().
		the_plan_has_succeeded().and().
		a_changeset_has_been_created("changeset2").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "value2"}`).and().
		the_plan_has_succeeded().and().
		a_changeset_has_been_merged("changeset1")

	when.
		a_changeset_is_merged("changeset2")

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_failed()
}

func TestMergeTwoComponentChangesetSequentially(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "value1"}`).and().
		the_plan_has_succeeded().and().
		a_changeset_has_been_merged("changeset1").
		a_changeset_has_been_created("changeset2").and().
		a_component_has_been_created_for_the_module_and_changeset("component2", `{"name": "value2"}`).and().
		the_plan_has_succeeded().and()

	when.
		a_changeset_is_merged("changeset2")

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_succeeded()
}

func TestMergeChangesetWithComponentUpdateConflicts(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(two_changesets_with_changes).and().
		a_changeset_has_been_merged("changeset1")

	when.
		a_changeset_is_merged("changeset2")

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_failed()
}

func TestRebaseChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_changeset_has_been_created("changeset1")

	when.
		the_changeset_is_rebased()

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_rebase_has_succeeded()
}

func TestRebaseChangesetWithChange(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "value1"}`).and().
		the_plan_has_succeeded()

	when.
		the_changeset_is_rebased()

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_rebase_has_succeeded()
}

func TestMergeChangesetWithComponentUpdateConflictsAfterRebase(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(two_changesets_with_changes).and().
		a_changeset_has_been_merged("changeset1").and().
		a_changeset_has_been_rebased("changeset2").and().
		all_changeset_plans_have_succeeded()

	when.
		a_changeset_is_merged("changeset2")

	then.
		the_changeset_creation_has_succeeded().and().
		the_changeset_merge_has_succeeded()
}
