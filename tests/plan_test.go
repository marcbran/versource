//go:build e2e && (all || plan)

package tests

import (
	"testing"
)

func TestCreatePlan(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created("plan-test-component", `{"key": "value"}`)

	when.
		a_plan_is_created()

	then.
		the_plan_creation_has_succeeded()
}

func TestCreatePlanWithNonexistentComponent(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created("component1", `{"key": "value"}`)

	when.
		a_plan_is_created_for_the_component("nonexistent")

	then.
		the_plan_creation_has_failed()
}

func TestCreatePlanWithNonexistentChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created("component1", `{"key": "value"}`)

	when.
		a_plan_is_created_for_the_changeset("nonexistent")

	then.
		the_plan_creation_has_failed()
}

func TestCreatePlanWithoutChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created("component1", `{"key": "value"}`)

	when.
		a_plan_is_created_without_changeset()

	then.
		the_plan_creation_has_failed()
}

func TestCreatePlanWithoutComponentID(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created("component1", `{"key": "value"}`)

	when.
		a_plan_is_created_without_component_id()

	then.
		the_plan_creation_has_failed()
}
