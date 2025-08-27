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
		a_module_is_created("test-module", "1.0.0")

	then.
		the_module_is_created_successfully()
}

func TestCreateModuleWithEmptySource(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created_with_empty_source("1.0.0")

	then.
		the_module_creation_has_failed()
}

func TestCreateModuleWithEmptyVersion(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created_with_empty_version("test-module")

	then.
		the_module_creation_has_failed()
}

func TestCreateModuleWithComplexSource(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("github.com/terraform-aws-modules/terraform-aws-vpc", "5.0.0")

	then.
		the_module_is_created_successfully()
}

func TestCreateModuleWithSemanticVersion(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("test-module", "v1.2.3")

	then.
		the_module_is_created_successfully()
}

func TestCreateModuleWithLongSource(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("https://github.com/very-long-organization-name/very-long-repository-name/terraform-aws-very-long-module-name", "1.0.0")

	then.
		the_module_is_created_successfully()
}

func TestCreateModuleWithSpecialCharacters(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("test-module-with-special-chars_123", "1.0.0-beta.1")

	then.
		the_module_is_created_successfully()
}

func TestCreateModuleWithLocalPath(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("./local/modules/test-module", "1.0.0")

	then.
		the_module_is_created_successfully()
}

func TestCreateModuleWithRegistrySource(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_blank_instance()

	when.
		a_module_is_created("registry.terraform.io/hashicorp/aws/5.0.0", "5.0.0")

	then.
		the_module_is_created_successfully()
}
