package tfjsonnet

import (
	"strings"

	v1 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v1"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
	"github.com/marcbran/versource/internal"
)

func newJsonnetBundlerFromComponent(component *internal.Component) v1.JsonnetFile {
	file := v1.New()
	source := component.ModuleVersion.Module.Source
	version := component.ModuleVersion.Version

	var dependency deps.Dependency

	if strings.HasPrefix(source, deps.GitSchemeSSH) || strings.HasPrefix(source, deps.GitSchemeHTTPS) {
		gitSource := parseGitSource(source)
		dependency = deps.Dependency{
			Source: deps.Source{
				GitSource: gitSource,
			},
			Version: version,
		}
	} else {
		dependency = deps.Dependency{
			Source: deps.Source{
				LocalSource: &deps.Local{
					Directory: source,
				},
			},
			Version: "",
		}
	}

	file.Dependencies.Set(dependency.Name(), dependency)
	return file
}

func parseGitSource(source string) *deps.Git {
	cleanSource := strings.TrimPrefix(strings.TrimPrefix(source, deps.GitSchemeHTTPS), deps.GitSchemeSSH)
	parts := strings.Split(cleanSource, "/")

	if len(parts) < 3 {
		return &deps.Git{
			Scheme: deps.GitSchemeHTTPS,
			Host:   parts[0],
			User:   parts[1],
			Repo:   parts[2],
			Subdir: "",
		}
	}

	host := parts[0]
	user := parts[1]
	repo := strings.TrimSuffix(parts[2], ".git")
	subdir := ""

	if len(parts) > 3 {
		subdir = "/" + strings.Join(parts[3:], "/")
	}

	scheme := deps.GitSchemeHTTPS
	if strings.HasPrefix(source, deps.GitSchemeSSH) {
		scheme = deps.GitSchemeSSH
	}

	return &deps.Git{
		Scheme: scheme,
		Host:   host,
		User:   user,
		Repo:   repo,
		Subdir: subdir,
	}
}
