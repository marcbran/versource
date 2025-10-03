//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Stage struct {
	t require.TestingT

	ModuleID      string
	ChangesetName string
	ComponentID   string
	PlanID        string
	MergeID       string
	RebaseID      string

	LastOutputArray []any
	LastOutputMap   map[string]any
	LastError       string
	LastExitCode    int

	LastQueryResult string
}

func scenario(t *testing.T) (*Stage, *Stage, *Stage) {
	stage := &Stage{
		t: t,
	}
	return stage, stage, stage
}

func (s *Stage) the_stage_is_cleared() *Stage {
	s.ModuleID = ""
	s.ChangesetName = ""
	s.ComponentID = ""
	s.PlanID = ""
	s.MergeID = ""
	s.RebaseID = ""
	s.LastOutputMap = nil
	s.LastError = ""
	s.LastExitCode = 0
	s.LastQueryResult = ""
	return s
}

func (s *Stage) the_module_id_is(moduleID string) *Stage {
	s.ModuleID = moduleID
	return s
}

func (s *Stage) the_changeset_name_is(changesetName string) *Stage {
	s.ChangesetName = changesetName
	return s
}

func (s *Stage) the_component_id_is(componentID string) *Stage {
	s.ComponentID = componentID
	return s
}

func (s *Stage) the_plan_id_is(planID string) *Stage {
	s.PlanID = planID
	return s
}

func (s *Stage) the_merge_id_is(mergeID string) *Stage {
	s.MergeID = mergeID
	return s
}

func (s *Stage) the_rebase_id_is(rebaseID string) *Stage {
	s.RebaseID = rebaseID
	return s
}

func (s *Stage) a_recreated_dbms() *Stage {
	return s.
		a_docker_compose_command_is_executed("down", "--remove-orphans").and().
		a_docker_compose_command_is_executed("up", "-d", "dolt", "dolt-client", "migrate")
}

func (s *Stage) a_database_user() *Stage {
	return s.
		a_db_query_has_been_run_as_root(`CREATE USER IF NOT EXISTS "versource"@"%" IDENTIFIED BY "versource";`).and().
		a_db_query_has_been_run_as_root(`GRANT ALL PRIVILEGES ON *.* TO "versource"@"%";`).
		a_db_query_has_been_run_as_root("FLUSH PRIVILEGES;")
}

func (s *Stage) a_created_server() *Stage {
	return s.a_docker_compose_command_is_executed("up", "-d", "server", "client")
}

type Dataset struct {
	Name string
}

var blank_instance = Dataset{Name: "blank-instance"}
var module_and_changeset = Dataset{Name: "module-and-changeset"}
var deleted_component = Dataset{Name: "deleted-component"}
var changeset_and_new_component = Dataset{Name: "changeset-and-new-component"}
var two_changesets_with_changes = Dataset{Name: "two-changesets-with-changes"}

func (s *Stage) a_clean_slate() *Stage {
	return s.an_empty_database().and().
		the_migrations_are_run().and().
		a_restarted_server().and().
		the_stage_is_cleared()
}

func (s *Stage) an_empty_database() *Stage {
	return s.
		a_db_query_has_been_run_as_root("DROP DATABASE IF EXISTS versource;").and().
		a_db_query_has_been_run_as_root("CREATE DATABASE IF NOT EXISTS versource;")
}

func (s *Stage) the_migrations_are_run() *Stage {
	return s.a_command_is_executed("migrate", "versource", "migrate").and().
		the_command_has_to_succeeded()
}

func (s *Stage) the_dataset(dataset Dataset) *Stage {
	return s.a_dataset_is_cloned(dataset).and().
		a_restarted_server().and().
		the_stage_is_cleared()
}

func (s *Stage) a_dataset_is_cloned(dataset Dataset) *Stage {
	return s.
		a_db_query_has_been_run_as_root("DROP DATABASE IF EXISTS versource;").and().
		a_db_query_has_been_run_as_root("CALL DOLT_CLONE('file:///datasets/" + dataset.Name + "', 'versource')").and().
		remote_branches_are_tracked_locally()
}

func (s *Stage) remote_branches_are_tracked_locally() *Stage {
	s.a_db_query_has_been_run_as_versource("SELECT name FROM dolt_remote_branches")
	lines := strings.Split(s.LastQueryResult, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "name" {
			continue
		}

		if !strings.HasPrefix(line, "remotes/origin/") {
			continue
		}

		branchName := strings.TrimPrefix(line, "remotes/origin/")
		if branchName == "main" {
			continue
		}

		s.a_db_query_has_been_run_as_versource(fmt.Sprintf("CALL DOLT_CHECKOUT('%s')", branchName))
	}

	return s
}

func (s *Stage) a_restarted_server() *Stage {
	return s.a_docker_compose_command_is_executed("restart", "server")
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

func (s *Stage) the_command_has_to_succeeded() *Stage {
	if s.LastExitCode != 0 {
		fmt.Println(s.LastError)
	}
	require.Equal(s.t, 0, s.LastExitCode)
	return s
}

func (s *Stage) the_command_has_failed() *Stage {
	assert.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *Stage) the_command_has_to_failed() *Stage {
	require.Equal(s.t, 1, s.LastExitCode)
	return s
}

func (s *Stage) a_client_command_is_executed(args ...string) *Stage {
	args = append(args, "--output", "json")
	args = append([]string{"versource"}, args...)
	return s.a_command_is_executed("client", args...)
}

func (s *Stage) a_command_is_executed(service string, args ...string) *Stage {
	cmd := exec.Command("docker", append([]string{"compose", "exec", "-T", service}, args...)...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if os.Getenv("VS_LOG") == "DEBUG" || os.Getenv("VS_LOG") == "TRACE" {
		fmt.Printf("%s: %s\n", service, strings.Join(args, " "))
	}

	err := cmd.Run()
	s.LastError = stderr.String()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			s.LastExitCode = exitErr.ExitCode()
		}
	} else {
		s.LastExitCode = 0
		output := stdout.String()
		if os.Getenv("VS_LOG") == "DEBUG" || os.Getenv("VS_LOG") == "TRACE" {
			fmt.Println(output)
		}
		if output != "" {
			var outputMap map[string]any
			jsonErr := json.Unmarshal([]byte(output), &outputMap)
			if jsonErr == nil {
				s.LastOutputMap = outputMap
			}
			var outputArray []any
			jsonErr = json.Unmarshal([]byte(output), &outputArray)
			if jsonErr == nil {
				s.LastOutputArray = outputArray
			}
		}
	}

	return s
}

func (s *Stage) a_db_query_has_been_run_as_root(query string) *Stage {
	return s.a_dolt_client_command_has_been_executed(query, "mysql", "-h", "dolt", "-u", "root")
}

func (s *Stage) a_db_query_has_been_run_as_versource(query string) *Stage {
	return s.a_dolt_client_command_has_been_executed(query, "mysql", "-h", "dolt", "-u", "versource", "-pversource", "versource")
}

func (s *Stage) a_dolt_client_command_has_been_executed(query string, args ...string) *Stage {
	s.a_dolt_client_command_is_executed(query, args...)

	if s.LastExitCode == 0 {
		return s
	}

	require.Fail(s.t, s.LastError)
	return s
}

func (s *Stage) a_db_query_is_run_as_root(query string) *Stage {
	return s.a_dolt_client_command_is_executed(query, "mysql", "-h", "dolt", "-u", "root")
}

func (s *Stage) a_db_query_is_run_as_versource(query string) *Stage {
	return s.a_dolt_client_command_is_executed(query, "mysql", "-h", "dolt", "-u", "versource", "-pversource", "versource")
}

func (s *Stage) a_dolt_client_command_is_executed(query string, args ...string) *Stage {
	args = append(args, "-e", query)
	cmd := exec.Command("docker", append([]string{"compose", "exec", "-T", "dolt-client"}, args...)...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if os.Getenv("VS_LOG") == "DEBUG" || os.Getenv("VS_LOG") == "TRACE" {
		fmt.Printf("root: %s\n", query)
	}

	err := cmd.Run()

	output := stdout.String()
	if os.Getenv("VS_LOG") == "DEBUG" || os.Getenv("VS_LOG") == "TRACE" {
		fmt.Println(output)
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			s.LastExitCode = exitErr.ExitCode()
		}
		s.LastError = stderr.String()
	} else {
		s.LastExitCode = 0
		s.LastError = ""
	}

	s.LastQueryResult = output

	return s
}

func (s *Stage) a_docker_compose_command_is_executed(args ...string) *Stage {
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)

	var stderr bytes.Buffer
	if os.Getenv("VS_LOG") == "TRACE" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stderr = &stderr
	}

	if os.Getenv("VS_LOG") == "DEBUG" || os.Getenv("VS_LOG") == "TRACE" {
		fmt.Printf("docker compose %s\n\n", strings.Join(args, " "))
	}

	err := cmd.Run()
	if err != nil {
		fmt.Println(stderr.String())
	}

	require.NoError(s.t, err)

	return s
}

func mainStage() *Stage {
	return &Stage{
		t: &mainT{},
	}
}

type mainT struct {
}

func (m mainT) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	m.FailNow()
}

func (m mainT) FailNow() {
	os.Exit(1)
}

func parseVariablesToArgs(variables string) []string {
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
