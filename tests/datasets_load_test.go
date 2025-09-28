//go:build e2e && datasets && load

package tests

import "testing"

func TestLoadDatasetBlankInstance(t *testing.T) {
	given, _, _ := scenario(t)

	given.
		the_dataset(blank_instance)
}

func TestLoadDatasetModuleAndChangeset(t *testing.T) {
	given, _, _ := scenario(t)

	given.
		the_dataset(module_and_changeset)
}

func TestLoadDatasetDeletedComponent(t *testing.T) {
	given, _, _ := scenario(t)

	given.
		the_dataset(deleted_component)
}

func TestLoadDatasetChangesetAndNewComponent(t *testing.T) {
	given, _, _ := scenario(t)

	given.
		the_dataset(changeset_and_new_component)
}

func TestLoadDatasetTwoChangesetWithChanges(t *testing.T) {
	given, _, _ := scenario(t)

	given.
		the_dataset(two_changesets_with_changes)
}
