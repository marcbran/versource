//go:build e2e && datasets

package tests

import "testing"

func TestDatasetBlankInstance(t *testing.T) {
	given, _, then := scenario(t)

	given.
		a_blank_instance()

	then.
		the_state_is_stored_in_the_blank_instance_dataset()
}

func TestDatasetModuleAndChangeset(t *testing.T) {
	given, _, then := scenario(t)

	given.
		a_blank_instance().and().
		a_module_has_been_created("hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1")

	then.
		the_state_is_stored_in_the_module_and_changeset_dataset()
}
