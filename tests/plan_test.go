//go:build e2e && (all || plan)

package tests

import (
	"testing"
)

func TestDefaultPlanFailing(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_non_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1")

	when.
		a_component_has_been_created_for_the_module_and_changeset("plan-test-component", `{"key": "value"}`)

	then.
		the_plan_creation_has_succeeded().and().
		the_plan_has_failed()
}

func TestDefaultPlanSucceeding(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().
		a_changeset_has_been_created("test1")

	when.
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "component1"}`)

	then.
		the_plan_creation_has_succeeded().and().
		the_plan_has_succeeded()
}

func TestDefaultPlanSucceedingWithVariables(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().
		a_changeset_has_been_created("test1")

	when.
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"name": "component1", "age": 40, "enabled": false}`)

	then.
		the_plan_creation_has_succeeded().and().
		the_plan_has_succeeded()
}

func TestCreatePlan(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("plan-test-component", `{"key": "value"}`)

	when.
		a_plan_is_created_for_the_changeset_and_component()

	then.
		the_plan_creation_has_succeeded()
}

func TestCreatePlanWithNonexistentComponent(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	when.
		a_plan_is_created_for_the_changeset("nonexistent")

	then.
		the_plan_creation_has_failed()
}

func TestCreatePlanWithNonexistentChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	when.
		a_plan_is_created_for_the_component("nonexistent")

	then.
		the_plan_creation_has_failed()
}

func TestCreatePlanWithoutChangeset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		an_existing_module_has_been_created().and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	when.
		a_plan_is_created_without_changeset()

	then.
		the_plan_creation_has_failed()
}
