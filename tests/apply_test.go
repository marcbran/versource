//go:build e2e && (all || apply)

package tests

import "testing"

func TestApplySucceeding(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		the_resources_module_has_been_created().and().
		a_changeset_has_been_created("changeset1")

	when.
		a_component_has_been_created_for_the_module_and_changeset("component1", `{}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged()

	then.
		all_applies_have_succeeded()
}

func TestApplyFailing(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_module_that_will_fail_on_apply_has_been_created().and().
		a_changeset_has_been_created("changeset1")

	when.
		a_component_has_been_created_for_the_module_and_changeset("component1", `{}`).and().
		the_plan_has_succeeded().and().
		the_changeset_has_been_merged()

	then.
		all_applies_have_failed()
}
