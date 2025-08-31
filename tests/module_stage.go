//go:build e2e

package tests

import (
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (s *Stage) a_module_has_been_created(source, version string) *Stage {
	s.a_module_is_created(source, version)
	return s.the_module_creation_has_succeeded()
}

func (s *Stage) a_module_is_created(source, version string) *Stage {
	args := []string{"module", "create", "--source", source}
	if version != "" {
		args = append(args, "--version", version)
	}
	s.execCommand(args...)
	if s.LastOutputMap != nil {
		if id, ok := s.LastOutputMap["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				s.ModuleID = fmt.Sprintf("%.0f", idFloat)
			}
		}
	}
	return s
}

func (s *Stage) the_module_creation_has_succeeded() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_module_creation_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *Stage) a_module_is_updated(moduleID int, version string) *Stage {
	args := []string{"module", "update", fmt.Sprintf("%d", moduleID), "--version", version}
	return s.execCommand(args...)
}

func (s *Stage) the_module_update_has_succeeded() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_module_update_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *Stage) the_module_is_deleted() *Stage {
	args := []string{"module", "delete", s.ModuleID}
	return s.execCommand(args...)
}

func (s *Stage) a_module_is_deleted(moduleID string) *Stage {
	args := []string{"module", "delete", moduleID}
	return s.execCommand(args...)
}

func (s *Stage) the_module_deletion_has_succeeded() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_module_deletion_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}
