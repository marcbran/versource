//go:build e2e && (all || component)

package tests

import (
	"testing"
)

func TestCreateComponent(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1")

	when.
		a_component_is_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	then.
		the_component_creation_has_succeeded()
}

func TestCreateComponentWithoutModuleInChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_changeset_has_been_created("test1").and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0")

	when.
		a_component_is_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	then.
		the_component_creation_has_failed()
}

func TestCreateComponentWithoutChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0")

	when.
		a_component_is_created_for_the_module("test1", "component1", `{"key": "value"}`)

	then.
		the_component_creation_has_succeeded()
}

func TestCreateComponentWithInvalidVariables(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1")

	when.
		a_component_is_created_for_the_module_and_changeset("component1", `{"invalid": "{"}`)

	then.
		the_component_creation_has_failed()
}

func TestUpdateComponent(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	when.
		the_component_is_updated(`{"key": "updated"}`)

	then.
		the_component_update_has_succeeded()
}

func TestUpdateComponentWithNonexistentID(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	when.
		a_component_is_updated_for_the_changeset("999", `{"key": "updated"}`)

	then.
		the_component_update_has_failed()
}

func TestUpdateComponentWithInvalidChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	when.
		a_component_is_updated("1", "does-not-exist", `{"key": "updated"}`)

	then.
		the_component_update_has_failed()
}

func TestUpdateComponentWithInvalidVariables(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	when.
		the_component_is_updated(`{"invalid": json`)

	then.
		the_component_update_has_failed()
}

func TestUpdateComponentWithNoFields(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	when.
		the_component_is_updated("")

	then.
		the_component_update_has_failed()
}

func TestCreateMultipleComponentsInSameChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1")

	when.
		a_component_is_created_for_the_module_and_changeset("component1", `{"key1": "value1"}`).and().
		a_component_is_created_for_the_module_and_changeset("component2", `{"key2": "value2"}`)

	then.
		both_component_creations_have_succeeded()
}

func TestCreateComponentsInDifferentChangesets(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("changeset1").and().
		a_changeset_has_been_created("changeset2")

	when.
		a_component_is_created_for_the_module("changeset1", "component1", `{"key1": "value1"}`).and().
		a_component_is_created_for_the_module("changeset2", "component2", `{"key2": "value2"}`)

	then.
		both_component_creations_have_succeeded()
}

func TestCreateComponentWithComplexVariables(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1")

	when.
		a_component_is_created_for_the_module_and_changeset("component1", `{"nested": {"key": "value"}, "array": [1, 2, 3], "boolean": true, "number": 42}`)

	then.
		the_component_creation_has_succeeded()
}

func TestCreateComponentWithDuplicateName(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component", `{"key": "value"}`)

	when.
		a_component_is_created_for_the_module_and_changeset("component", `{"key2": "value2"}`)

	then.
		the_component_creation_has_failed()
}
