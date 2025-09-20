//go:build e2e && datasets

package tests

import "testing"

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
