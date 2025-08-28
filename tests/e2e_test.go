//go:build e2e && (all || datasets)

package tests

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	err := runDockerCompose("down")
	if err != nil {
		os.Exit(1)
	}
	err = runDockerCompose("build", "--no-cache")
	if err != nil {
		os.Exit(1)
	}
	err = runDockerCompose("up", "-d")
	if err != nil {
		os.Exit(1)
	}

	code := m.Run()

	//runDockerCompose("down")

	os.Exit(code)
}
