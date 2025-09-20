//go:build e2e && datasets

package tests

func (s *Stage) the_state_is_stored_in_the_dataset(dataset Dataset) *Stage {
	return s.
		the_dataset_is_removed(dataset).and().
		a_db_query_is_run_as_versource("CALL DOLT_REMOTE('add', 'origin', 'file:///datasets/" + dataset.Name + "')").and().
		a_db_query_is_run_as_versource("CALL DOLT_PUSH('origin', 'main')")
}

func (s *Stage) the_dataset_is_removed(dataset Dataset) *Stage {
	return s.a_command_is_executed("dolt", "rm", "-rf", "/datasets/"+dataset.Name).and().
		the_command_has_to_succeeded()
}
