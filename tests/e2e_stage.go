//go:build e2e
// +build e2e

package tests

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

type E2EStage struct {
	t *testing.T

	ChangesetName string
	ModuleID      string

	LastOutput   string
	LastError    string
	LastExitCode int
}

func scenario(t *testing.T) (*E2EStage, *E2EStage, *E2EStage) {
	stage := &E2EStage{t: t}
	return stage, stage, stage
}

func (s *E2EStage) a_blank_instance() *E2EStage {
	s.execRootQuery("DROP DATABASE IF EXISTS versource;")
	s.execRootQuery("CALL DOLT_CLONE('file:///datasets/blank-instance', 'versource')")
	return s
}

func (s *E2EStage) a_changeset_has_been_created(name string) *E2EStage {
	return s.a_changeset_is_created(name).and().
		the_changeset_is_created_successfully()
}

func (s *E2EStage) a_changeset_is_created(name string) *E2EStage {
	s.ChangesetName = name
	return s.execCommand("changeset", "create", "--name", name)
}

func (s *E2EStage) the_changeset_is_created_successfully() *E2EStage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *E2EStage) the_changeset_creation_has_failed() *E2EStage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *E2EStage) the_changeset_has_been_merged() *E2EStage {
	return s.the_changeset_is_merged().and().
		the_changeset_is_merged_successfully()
}

func (s *E2EStage) the_changeset_is_merged() *E2EStage {
	return s.execCommand("changeset", "merge", s.ChangesetName)
}

func (s *E2EStage) the_changeset_is_merged_successfully() *E2EStage {
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *E2EStage) a_module_has_been_created(changeset, source, version, variables string) *E2EStage {
	return s.a_module_is_created(changeset, source, version, variables).and().
		the_module_is_created_successfully()
}

func (s *E2EStage) a_module_is_created(changeset, source, version, variables string) *E2EStage {
	s.ChangesetName = changeset
	args := []string{"module", "create", "--changeset", changeset, "--source", source, "--variables", variables}
	if version != "" {
		args = append(args, "--version", version)
	}
	return s.execCommand(args...)
}

func (s *E2EStage) a_module_is_created_with_empty_source(changeset, version, variables string) *E2EStage {
	s.ChangesetName = changeset
	args := []string{"module", "create", "--changeset", changeset, "--source", "", "--variables", variables}
	if version != "" {
		args = append(args, "--version", version)
	}
	return s.execCommand(args...)
}

func (s *E2EStage) the_module_is_created_successfully() *E2EStage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *E2EStage) the_module_creation_has_failed() *E2EStage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *E2EStage) both_modules_are_created_successfully() *E2EStage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *E2EStage) the_module_is_updated(changeset, moduleID, source, version, variables string) *E2EStage {
	s.ChangesetName = changeset
	s.ModuleID = moduleID
	args := []string{"module", "update", moduleID, "--changeset", changeset}
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

func (s *E2EStage) the_module_is_updated_with_no_fields(changeset, moduleID string) *E2EStage {
	s.ChangesetName = changeset
	s.ModuleID = moduleID
	return s.execCommand("module", "update", moduleID, "--changeset", changeset)
}

func (s *E2EStage) the_module_is_updated_successfully() *E2EStage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *E2EStage) the_module_update_has_failed() *E2EStage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *E2EStage) a_plan_has_been_created(changeset, moduleID string) *E2EStage {
	return s.a_plan_is_created(changeset, moduleID).and().
		the_plan_is_created_successfully()
}

func (s *E2EStage) a_plan_is_created(changeset, moduleID string) *E2EStage {
	s.ChangesetName = changeset
	s.ModuleID = moduleID
	return s.execCommand("plan", "--module-id", moduleID, "--changeset", changeset)
}

func (s *E2EStage) a_plan_is_created_without_changeset(moduleID string) *E2EStage {
	s.ModuleID = moduleID
	return s.execCommand("plan", "--module-id", moduleID)
}

func (s *E2EStage) a_plan_is_created_without_module_id(changeset string) *E2EStage {
	s.ChangesetName = changeset
	return s.execCommand("plan", "--changeset", changeset)
}

func (s *E2EStage) the_plan_is_created_successfully() *E2EStage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *E2EStage) the_plan_creation_has_failed() *E2EStage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *E2EStage) and() *E2EStage {
	return s
}

func runDockerCompose(args ...string) error {
	return exec.Command("docker", append([]string{"compose"}, args...)...).Run()
}

func (s *E2EStage) execCommand(args ...string) *E2EStage {
	cmd := exec.Command("docker", append([]string{"compose", "exec", "-T", "client", "./versource"}, args...)...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	s.LastOutput = stdout.String()
	s.LastError = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			s.LastExitCode = exitErr.ExitCode()
		} else {
			s.LastExitCode = -1
		}
	} else {
		s.LastExitCode = 0
	}

	return s
}

func (s *E2EStage) execRootQuery(query string) string {
	cmd := exec.Command("docker", "compose", "exec", "-T", "db-client", "mysql", "-h", "dolt", "-u", "root", "-e", query)

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return string(output)
}

func (s *E2EStage) execQuery(query string) string {
	cmd := exec.Command("docker", "compose", "exec", "-T", "db-client", "mysql", "-h", "dolt", "-u", "versource", "-pversource", "versource", "-e", query)

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return string(output)
}
