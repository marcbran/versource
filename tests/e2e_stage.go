//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Stage struct {
	t *testing.T

	ModuleID      string
	ChangesetName string
	ComponentID   string
	PlanID        string
	MergeID       string
	RebaseID      string

	LastOutputMap map[string]any
	LastError     string
	LastExitCode  int
}

func scenario(t *testing.T) (*Stage, *Stage, *Stage) {
	stage := &Stage{
		t: t,
	}
	return stage, stage, stage
}

func (s *Stage) where() *Stage {
	return s
}

func (s *Stage) and() *Stage {
	return s
}

func (s *Stage) the_command_has_succeeded() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	assert.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_command_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func runDockerCompose(args ...string) error {
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)

	var stderr bytes.Buffer

	if os.Getenv("VS_LOG") == "DEBUG" {
		fmt.Printf("DEBUG: Executing docker compose: %v\n", args)
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	if err != nil {
		fmt.Println(stderr.String())
	}
	return err
}

func (s *Stage) runDockerCompose(args ...string) error {
	return runDockerCompose(args...)
}

func (s *Stage) execCommand(args ...string) *Stage {
	args = append(args, "--output", "json")
	cmd := exec.Command("docker", append([]string{"compose", "exec", "-T", "client", "./versource"}, args...)...)

	if os.Getenv("VS_LOG") == "DEBUG" {
		fmt.Printf("DEBUG: Executing command: %v\n", args)
	}

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
		if os.Getenv("VS_LOG") == "DEBUG" {
			fmt.Printf("DEBUG: Command output: %s\n", output)
		}
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

func (s *Stage) parseVariablesToArgs(variables string) []string {
	if variables == "" {
		return []string{}
	}

	var args []string
	var variablesMap map[string]any
	err := json.Unmarshal([]byte(variables), &variablesMap)
	if err != nil {
		return []string{}
	}

	for key, value := range variablesMap {
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
		case bool:
			if v {
				valueStr = "true"
			} else {
				valueStr = "false"
			}
		case float64:
			valueStr = fmt.Sprintf("%.0f", v)
		case nil:
			valueStr = "null"
		default:
			valueBytes, _ := json.Marshal(value)
			valueStr = string(valueBytes)
		}
		args = append(args, "--variable", fmt.Sprintf("%s=%s", key, valueStr))
	}

	return args
}
