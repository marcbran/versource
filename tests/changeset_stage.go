//go:build e2e

package tests

func (s *Stage) a_changeset_has_been_created(name string) *Stage {
	return s.a_changeset_is_created(name).and().
		the_changeset_creation_has_succeeded()
}

func (s *Stage) a_changeset_is_created(name string) *Stage {
	s.ChangesetName = name
	return s.execCommand("changeset", "create", "--name", name)
}

func (s *Stage) the_changeset_creation_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_changeset_creation_has_failed() *Stage {
	return s.the_command_has_failed()
}

func (s *Stage) the_changeset_has_been_merged() *Stage {
	return s.the_changeset_is_merged().and().
		the_changeset_merge_has_succeeded()
}

func (s *Stage) the_changeset_is_merged() *Stage {
	return s.execCommand("changeset", "merge", s.ChangesetName)
}

func (s *Stage) the_changeset_merge_has_succeeded() *Stage {
	return s.the_command_has_succeeded()
}

func (s *Stage) the_changeset_merge_has_failed() *Stage {
	return s.the_command_has_failed()
}
