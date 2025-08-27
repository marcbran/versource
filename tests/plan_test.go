//go:build e2e

package tests

// import (
// 	"testing"
// )

// func TestCreatePlan(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1").and().
// 		a_component_has_been_created("test1", "test-component", "1.0.0", `{"key": "value"}`)

// 	when.
// 		a_plan_is_created("test1", "1")

// 	then.
// 		the_plan_is_created_successfully()
// }

// func TestCreatePlanWithNonexistentComponent(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1")

// 	when.
// 		a_plan_is_created("test1", "999")

// 	then.
// 		the_plan_creation_has_failed()
// }

// func TestCreatePlanWithNonexistentChangeset(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1").and().
// 		a_component_has_been_created("test1", "test-component", "1.0.0", `{"key": "value"}`)

// 	when.
// 		a_plan_is_created("nonexistent", "1")

// 	then.
// 		the_plan_creation_has_failed()
// }

// func TestCreatePlanWithoutChangeset(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1").and().
// 		a_component_has_been_created("test1", "test-component", "1.0.0", `{"key": "value"}`)

// 	when.
// 		a_plan_is_created_without_changeset("1")

// 	then.
// 		the_plan_creation_has_failed()
// }

// func TestCreatePlanWithoutComponentID(t *testing.T) {
// 	given, when, then := scenario(t)

// 	given.
// 		a_blank_instance().and().
// 		a_changeset_has_been_created("test1").and().
// 		a_component_has_been_created("test1", "test-component", "1.0.0", `{"key": "value"}`)

// 	when.
// 		a_plan_is_created_without_component_id("test1")

// 	then.
// 		the_plan_creation_has_failed()
// }
