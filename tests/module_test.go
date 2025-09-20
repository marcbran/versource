//go:build e2e && (all || module)

package tests

import "testing"

func TestCreateModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance)

	when.
		a_module_is_created("consul-aws", "hashicorp/consul/aws", "0.1.0")

	then.
		the_module_creation_has_succeeded()
}

func TestCreateGitModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance)

	when.
		a_module_is_created("example", "github.com/hashicorp/example?ref=v1.2.0", "")

	then.
		the_module_creation_has_succeeded()
}

func TestCreateLocalModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance)

	when.
		a_module_is_created("test-module", "./local/modules/test-module", "")

	then.
		the_module_creation_has_succeeded()
}

func TestUpdateModuleWithVersion(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0")

	when.
		the_module_is_updated("0.2.0")

	then.
		the_module_update_has_succeeded()
}

func TestUpdateModuleWithoutVersion(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_module_has_been_created("test-module", "./local/modules/test-module", "")

	when.
		the_module_is_updated("1.0.0")

	then.
		the_module_update_has_failed()
}

func TestDeleteModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0")

	when.
		the_module_is_deleted()

	then.
		the_module_deletion_has_succeeded()
}

func TestDeleteModuleWithNotYetMergedComponents(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance).and().
		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
		a_changeset_has_been_created("test1").and().
		a_component_has_been_created_for_the_module_and_changeset("component1", `{"key": "value"}`)

	when.
		the_module_is_deleted()

	then.
		the_module_deletion_has_succeeded()
}

// TODO Need a way to wait for the changeset merge to complete
// func TestDeleteModuleWithComponents(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		the_dataset(blank_instance).and().
// 		a_module_has_been_created("consul-aws", "hashicorp/consul/aws", "0.1.0").and().
// 		a_changeset_has_been_created("test1").and().
// 		a_component_has_been_created_for_the_module_and_changeset("component1", `{"key": "value"}`).and().
// 		the_changeset_has_been_merged()

// 	when.
// 		the_module_is_deleted()

// 	then.
// 		the_module_deletion_has_failed()
// }

func TestDeleteNonExistentModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_dataset(blank_instance)

	when.
		a_module_is_deleted("999")

	then.
		the_module_deletion_has_failed()
}
