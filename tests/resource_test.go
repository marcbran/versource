//go:build e2e && (all || resource)

package tests

import "testing"

func TestResourceAreCreated(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		the_resources_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged().and().
		all_applies_have_succeeded()

	when.
		the_resources_are_listed().and()

	then.
		there_are_resources(6)
}

func TestResourceArePruned(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		the_resources_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged().and().
		all_applies_have_succeeded().and().
		a_changeset_has_been_created("changeset2").and().
		the_component_has_been_updated(`{"names": "a,b,c"}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged().and().
		all_applies_have_succeeded()

	when.
		the_resources_are_listed().and()

	then.
		there_are_resources(3)
}

func TestResourceAreDropped(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		the_resources_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"drop": "d,e,f"}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged().and().
		all_applies_have_succeeded()

	when.
		the_resources_are_listed().and()

	then.
		there_are_resources(3)
}

func TestResourceAreAdded(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		the_resources_module_has_been_created().and().
		a_changeset_has_been_created("changeset1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"keep": "", "add": "p,q"}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged().and().
		all_applies_have_succeeded()

	when.
		the_resources_are_listed().and()

	then.
		there_are_resources(2)
}
