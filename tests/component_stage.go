//go:build e2e

package tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
)

func (s *Stage) a_component_has_been_created(changeset, source, version, variables string) *Stage {
	return s.a_component_is_created(changeset, source, version, variables).and().
		the_component_is_created_successfully()
}

func (s *Stage) a_component_is_created(changeset, source, version, variables string) *Stage {
	s.ChangesetName = changeset
	args := []string{"component", "create", "--changeset", changeset, "--source", source, "--variables", variables}
	if version != "" {
		args = append(args, "--version", version)
	}
	return s.execCommand(args...)
}

func (s *Stage) a_component_is_created_with_empty_source(changeset, version, variables string) *Stage {
	s.ChangesetName = changeset
	args := []string{"component", "create", "--changeset", changeset, "--source", "", "--variables", variables}
	if version != "" {
		args = append(args, "--version", version)
	}
	return s.execCommand(args...)
}

func (s *Stage) the_component_is_created_successfully() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_component_creation_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *Stage) both_components_are_created_successfully() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_component_is_updated(changeset, componentID, source, version, variables string) *Stage {
	s.ChangesetName = changeset
	s.ComponentID = componentID
	args := []string{"component", "update", componentID, "--changeset", changeset}
	if source != "" {
		args = append(args, "--source", source)
	}
	if version != "" {
		args = append(args, "--version", version)
	}
	if variables != "" {
		args = append(args, "--variables", variables)
	}
	return s.execCommand(args...)
}

func (s *Stage) the_component_is_updated_with_no_fields(changeset, componentID string) *Stage {
	s.ChangesetName = changeset
	s.ComponentID = componentID
	return s.execCommand("component", "update", componentID, "--changeset", changeset)
}

func (s *Stage) the_component_is_updated_successfully() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_component_update_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}
