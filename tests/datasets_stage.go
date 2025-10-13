//go:build e2e && datasets

package tests

import "fmt"

func (s *Stage) the_state_is_stored_in_the_dataset(dataset Dataset) *Stage {
	return s.
		the_dataset_is_removed(dataset).and().
		a_db_query_has_been_run_as_versource(fmt.Sprintf("CALL DOLT_REMOTE('add', '%s', 'file:///datasets/%s')", dataset.Name, dataset.Name)).and().
		a_db_query_has_been_run_as_versource(fmt.Sprintf("CALL DOLT_PUSH('%s', '--all')", dataset.Name))
}

func (s *Stage) the_dataset_is_removed(dataset Dataset) *Stage {
	return s.a_command_is_executed("dolt", "rm", "-rf", fmt.Sprintf("/datasets/%s", dataset.Name)).and().
		the_command_has_to_succeed()
}
