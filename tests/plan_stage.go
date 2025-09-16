//go:build e2e

package tests

func (s *Stage) a_plan_has_been_created() *Stage {
	return s.a_plan_is_created_for_the_changeset_and_component().and().
		the_plan_creation_has_succeeded()
}

func (s *Stage) a_plan_is_created_for_the_changeset_and_component() *Stage {
	return s.a_plan_is_created(s.ChangesetName, s.ComponentID)
}

func (s *Stage) a_plan_is_created_for_the_changeset(componentID string) *Stage {
	return s.a_plan_is_created(s.ChangesetName, componentID)
}

func (s *Stage) a_plan_is_created_for_the_component(changeset string) *Stage {
	return s.a_plan_is_created(changeset, s.ComponentID)
}

func (s *Stage) a_plan_is_created(changeset, componentID string) *Stage {
	return s.execCommand("plan", "--changeset", changeset, "--component-id", componentID)
}

func (s *Stage) a_plan_is_created_without_changeset() *Stage {
	return s.execCommand("plan", "--component-id", s.ComponentID)
}

func (s *Stage) a_plan_is_created_without_component_id() *Stage {
	return s.execCommand("plan", "--changeset", s.ChangesetName)
}

func (s *Stage) the_plan_creation_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_plan_creation_has_failed() *Stage {
	return s.the_command_has_failed()
}
