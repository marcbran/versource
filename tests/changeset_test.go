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
		the_changeset_creation_has_succeeded()
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

func TestMergeChangesetWithComponents(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created("component1", `{"key": "value"}`)

	when.
		the_changeset_is_merged()

	then.
		the_changeset_creation_has_succeeded()
}

func TestMergeChangesetWithMultipleComponents(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created("component1", `{"key1": "value1"}`).and().
		a_component_has_been_created("component2", `{"key2": "value2"}`)

	when.
		the_changeset_is_merged()

	then.
		the_changeset_creation_has_succeeded()
}
