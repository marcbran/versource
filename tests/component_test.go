//go:build e2e

package tests

// import (
// 	"testing"
// )

// func TestCreateComponent(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1")

// 	when.
// 		a_component_is_created("test1", "test-component", "1.0.0", `{"key": "value"}`)

// 	then.
// 		the_component_is_created_successfully()
// }

// func TestCreateComponentWithoutChangeset(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance()

// 	when.
// 		a_component_is_created("nonexistent", "test-component", "1.0.0", `{"key": "value"}`)

// 	then.
// 		the_component_is_created_successfully()
// }

// func TestCreateComponentWithInvalidVariables(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1")

// 	when.
// 		a_component_is_created("test1", "test-component", "1.0.0", `{"invalid": json`)

// 	then.
// 		the_component_creation_has_failed()
// }

// func TestCreateComponentWithEmptySource(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1")

// 	when.
// 		a_component_is_created_with_empty_source("test1", "1.0.0", `{"key": "value"}`)

// 	then.
// 		the_component_creation_has_failed()
// }

// func TestUpdateComponent(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1").and().
// 		a_component_has_been_created("test1", "test-component", "1.0.0", `{"key": "value"}`)

// 	when.
// 		the_component_is_updated("test1", "1", "updated-component", "2.0.0", `{"key": "updated"}`)

// 	then.
// 		the_component_is_updated_successfully()
// }

// func TestUpdateComponentWithNonexistentID(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1")

// 	when.
// 		the_component_is_updated("test1", "999", "updated-component", "2.0.0", `{"key": "updated"}`)

// 	then.
// 		the_component_update_has_failed()
// }

// func TestUpdateComponentWithInvalidChangeset(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1").and().
// 		a_component_has_been_created("test1", "test-component", "1.0.0", `{"key": "value"}`)

// 	when.
// 		the_component_is_updated("nonexistent", "1", "updated-component", "2.0.0", `{"key": "updated"}`)

// 	then.
// 		the_component_update_has_failed()
// }

// func TestUpdateComponentWithInvalidVariables(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1").and().
// 		a_component_has_been_created("test1", "test-component", "1.0.0", `{"key": "value"}`)

// 	when.
// 		the_component_is_updated("test1", "1", "updated-component", "2.0.0", `{"invalid": json`)

// 	then.
// 		the_component_update_has_failed()
// }

// func TestUpdateComponentWithNoFields(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1").and().
// 		a_component_has_been_created("test1", "test-component", "1.0.0", `{"key": "value"}`)

// 	when.
// 		the_component_is_updated_with_no_fields("test1", "1")

// 	then.
// 		the_component_update_has_failed()
// }

// func TestCreateMultipleComponentsInSameChangeset(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1")

// 	when.
// 		a_component_is_created("test1", "component1", "1.0.0", `{"key1": "value1"}`).and().
// 		a_component_is_created("test1", "component2", "2.0.0", `{"key2": "value2"}`)

// 	then.
// 		both_components_are_created_successfully()
// }

// func TestCreateComponentsInDifferentChangesets(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("changeset1").and().
// 		a_changeset_has_been_created("changeset2")

// 	when.
// 		a_component_is_created("changeset1", "component1", "1.0.0", `{"key1": "value1"}`).and().
// 		a_component_is_created("changeset2", "component2", "2.0.0", `{"key2": "value2"}`)

// 	then.
// 		both_components_are_created_successfully()
// }

// func TestCreateChangesetWithSpecialCharacters(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance()

// 	when.
// 		a_changeset_is_created("test-changeset-123")

// 	then.
// 		the_changeset_is_created_successfully()
// }

// func TestCreateComponentWithComplexVariables(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1")

// 	when.
// 		a_component_is_created("test1", "complex-component", "1.0.0", `{"nested": {"key": "value"}, "array": [1, 2, 3], "boolean": true, "number": 42}`)

// 	then.
// 		the_component_is_created_successfully()
// }
