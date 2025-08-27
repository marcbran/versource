//go:build e2e

package tests

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	runDockerCompose("down")
	runDockerCompose("build", "--no-cache")
	runDockerCompose("up", "-d")

	code := m.Run()

	runDockerCompose("down")

	os.Exit(code)
}
