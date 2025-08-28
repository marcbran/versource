//go:build e2e

package tests

func (s *Stage) a_plan_has_been_created() *Stage {
	return s.a_plan_is_created().and().
		the_plan_creation_has_succeeded()
}

func (s *Stage) a_plan_is_created() *Stage {
	return s.execCommand("plan", "--component-id", s.ComponentID, "--changeset", s.ChangesetName)
}

func (s *Stage) a_plan_is_created_for_the_component(componentID string) *Stage {
	s.ComponentID = componentID
	return s.execCommand("plan", "--component-id", componentID, "--changeset", s.ChangesetName)
}

func (s *Stage) a_plan_is_created_for_the_changeset(changeset string) *Stage {
	s.ChangesetName = changeset
	return s.execCommand("plan", "--component-id", s.ComponentID, "--changeset", changeset)
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
