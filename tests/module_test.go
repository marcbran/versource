//go:build e2e && (all || module)

package tests

import "testing"

func TestCreateModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("hashicorp/consul/aws", "0.1.0")

	then.
		the_module_is_created_successfully()
}

func TestCreateGitModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("github.com/hashicorp/example?ref=v1.2.0", "")

	then.
		the_module_is_created_successfully()
}

func TestCreateLocalModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("./local/modules/test-module", "")

	then.
		the_module_is_created_successfully()
}

func TestCreateModuleFailure(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("hashicorp/consul/aws", "")

	then.
		the_module_creation_has_failed()
}

func TestUpdateModuleWithVersion(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("hashicorp/consul/aws", "0.1.0")

	when.
		a_module_is_updated(1, "0.2.0")

	then.
		the_module_is_updated_successfully()
}

func TestUpdateModuleWithoutVersion(t *testing.T) {
	given, when, then := scenario(t)

	given.
		the_blank_instance_dataset().and().
		a_module_has_been_created("./local/modules/test-module", "")

	when.
		a_module_is_updated(1, "1.0.0")

	then.
		the_module_update_has_failed()
}
