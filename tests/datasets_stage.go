//go:build e2e

package tests

import (
	"fmt"
)

func (s *Stage) a_blank_instance() *Stage {
	s.execRootQuery("DROP DATABASE IF EXISTS versource;")
	s.runDockerCompose("restart", "db-init")
	s.runDockerCompose("restart", "migrate")
	s.runDockerCompose("restart", "server")
	return s
}

func (s *Stage) the_blank_instance_dataset() *Stage {
	return s.the_dataset("blank-instance")
}

func (s *Stage) the_state_is_stored_in_the_blank_instance_dataset() *Stage {
	return s.the_state_is_stored_in_the_dataset("blank-instance")
}

func (s *Stage) the_module_and_changeset_dataset() *Stage {
	return s.the_dataset("module-and-changeset")
}

func (s *Stage) the_state_is_stored_in_the_module_and_changeset_dataset() *Stage {
	return s.the_state_is_stored_in_the_dataset("module-and-changeset")
}

func (s *Stage) the_dataset(name string) *Stage {
	s.execRootQuery("DROP DATABASE IF EXISTS versource;")
	s.runDockerCompose("restart", "db-init")
	s.runDockerCompose("restart", "migrate")
	s.runDockerCompose("restart", "server")
	s.execRootQuery("CALL DOLT_CLONE('file:///datasets/" + name + "', 'versource')")
	return s
}

func (s *Stage) the_state_is_stored_in_the_dataset(name string) *Stage {
	fmt.Println("Pushing dataset: " + name)
	fmt.Println(s.execQuery("CALL DOLT_REMOTE('add', 'origin', 'file:///datasets/" + name + "')"))
	fmt.Println(s.execQuery("CALL DOLT_PUSH('origin', 'main')"))
	return s
}
