//go:build e2e

package tests

import (
	"testing"
)

func TestCreateModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_registry_module_is_created("hashicorp/consul/aws", "0.1.0")

	then.
		the_module_is_created_successfully()
}

func TestCreateGitModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_github_module_is_created("github.com/hashicorp/example?ref=v1.2.0")

	then.
		the_module_is_created_successfully()
}

func TestCreateLocalModule(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_local_module_is_created("./local/modules/test-module")

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
