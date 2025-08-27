//go:build e2e

package tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
)

func (s *Stage) a_changeset_has_been_created(name string) *Stage {
	return s.a_changeset_is_created(name).and().
		the_changeset_is_created_successfully()
}

func (s *Stage) a_changeset_is_created(name string) *Stage {
	s.ChangesetName = name
	return s.execCommand("changeset", "create", "--name", name)
}

func (s *Stage) the_changeset_is_created_successfully() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_changeset_creation_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *Stage) the_changeset_has_been_merged() *Stage {
	return s.the_changeset_is_merged().and().
		the_changeset_is_merged_successfully()
}

func (s *Stage) the_changeset_is_merged() *Stage {
	return s.execCommand("changeset", "merge", s.ChangesetName)
}

func (s *Stage) the_changeset_is_merged_successfully() *Stage {
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}
