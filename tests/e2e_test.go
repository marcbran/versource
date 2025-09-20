//go:build e2e

package tests

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	s := mainStage()
	s.a_recreated_dbms().and().
		an_empty_database().and().
		a_database_user().and().
		a_created_server()

	code := m.Run()

	os.Exit(code)
}
