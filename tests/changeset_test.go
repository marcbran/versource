//go:build e2e

package tests

import (
	"testing"
)

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

func TestMergeChangesetWithComponents(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created("test1", "test-component", "1.0.0", `{"key": "value"}`).and().
		a_plan_has_been_created("test1", "1")

	when.
		the_changeset_is_merged()

	then.
		the_changeset_is_merged_successfully()
}

func TestMergeChangesetWithMultipleComponents(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created("test1", "component1", "1.0.0", `{"key1": "value1"}`).and().
		a_component_has_been_created("test1", "component2", "2.0.0", `{"key2": "value2"}`).and().
		a_plan_has_been_created("test1", "1").and().
		a_plan_has_been_created("test1", "2")

	when.
		the_changeset_is_merged()

	then.
		the_changeset_is_merged_successfully()
}
