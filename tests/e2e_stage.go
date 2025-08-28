//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

type Stage struct {
	t *testing.T

	ModuleID      string
	ChangesetName string
	ComponentID   string

	LastOutputMap map[string]any
	LastError     string
	LastExitCode  int
}

func scenario(t *testing.T) (*Stage, *Stage, *Stage) {
	stage := &Stage{t: t}
	return stage, stage, stage
}

func (s *Stage) a_blank_instance() *Stage {
	if os.Getenv("USE_DATASET") == "false" {
		s.execRootQuery("DROP DATABASE IF EXISTS versource;")
		s.runDockerCompose("restart", "db-init")
		s.runDockerCompose("restart", "migrate")
		s.runDockerCompose("restart", "server")
		s.execQuery("CALL DOLT_REMOTE('add', 'origin', 'file:///datasets/blank-instance')")
		s.execQuery("CALL DOLT_PUSH('origin', 'main')")
	} else {
		s.execRootQuery("DROP DATABASE IF EXISTS versource;")
		s.execRootQuery("CALL DOLT_CLONE('file:///datasets/blank-instance', 'versource')")
	}

	return s
}

func (s *Stage) and() *Stage {
	return s
}

func runDockerCompose(args ...string) error {
	return exec.Command("docker", append([]string{"compose"}, args...)...).Run()
}

func (s *Stage) runDockerCompose(args ...string) error {
	return runDockerCompose(args...)
}

func (s *Stage) execCommand(args ...string) *Stage {
	args = append(args, "--output", "json")
	cmd := exec.Command("docker", append([]string{"compose", "exec", "-T", "client", "./versource"}, args...)...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	s.LastError = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			s.LastExitCode = exitErr.ExitCode()
		} else {
			s.LastExitCode = -1
		}
	} else {
		s.LastExitCode = 0
		output := stdout.String()
		fmt.Println(output)
		if output != "" {
			var outputMap map[string]any
			jsonErr := json.Unmarshal([]byte(output), &outputMap)
			if jsonErr == nil {
				s.LastOutputMap = outputMap
			}
		}
	}

	return s
}

func (s *Stage) execRootQuery(query string) string {
	cmd := exec.Command("docker", "compose", "exec", "-T", "db-client", "mysql", "-h", "dolt", "-u", "root", "-e", query)

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return string(output)
}

func (s *Stage) execQuery(query string) string {
	cmd := exec.Command("docker", "compose", "exec", "-T", "db-client", "mysql", "-h", "dolt", "-u", "versource", "-pversource", "versource", "-e", query)

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return string(output)
}
