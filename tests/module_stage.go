//go:build e2e

package tests

import (
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (s *Stage) a_module_has_been_created(source, version string) *Stage {
	return s.a_module_is_created(source, version).and().
		the_module_is_created_successfully()
}

func (s *Stage) a_module_is_created(source, version string) *Stage {
	args := []string{"module", "create", "--source", source, "--version", version}
	return s.execCommand(args...)
}

func (s *Stage) a_module_is_created_with_empty_source(version string) *Stage {
	args := []string{"module", "create", "--source", "", "--version", version}
	return s.execCommand(args...)
}

func (s *Stage) a_module_is_created_with_empty_version(source string) *Stage {
	args := []string{"module", "create", "--source", source, "--version", ""}
	return s.execCommand(args...)
}

func (s *Stage) the_module_is_created_successfully() *Stage {
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

func (s *Stage) the_module_creation_output_contains(expected string) *Stage {
	assert.Contains(s.t, s.LastOutput, expected)
	return s
}
