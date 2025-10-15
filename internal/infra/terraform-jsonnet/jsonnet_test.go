package tfjsonnet

import (
	"testing"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
	"github.com/marcbran/versource/internal"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestNewJsonnetBundlerFromComponent(t *testing.T) {
	tests := []struct {
		name            string
		component       *versource.Component
		expectedGit     *deps.Git
		expectedLocal   *deps.Local
		expectedVersion string
	}{
		{
			name: "HTTPS git source with subdir",
			component: &versource.Component{
				ModuleVersion: versource.ModuleVersion{
					Module: versource.Module{
						Source: "https://git.brndn.live/marcbran/raspi/tf/modules/github-repository",
					},
					Version: "73cfe497e69a8d89c00f3cbf49d69b94afe7049c",
				},
				Variables: datatypes.JSON(`{"name":"test"}`),
			},
			expectedGit: &deps.Git{
				Scheme: deps.GitSchemeHTTPS,
				Host:   "git.brndn.live",
				User:   "marcbran",
				Repo:   "raspi",
				Subdir: "/tf/modules/github-repository",
			},
			expectedLocal:   nil,
			expectedVersion: "73cfe497e69a8d89c00f3cbf49d69b94afe7049c",
		},
		{
			name: "SSH git source without subdir",
			component: &versource.Component{
				ModuleVersion: versource.ModuleVersion{
					Module: versource.Module{
						Source: "ssh://git@github.com/owner/repo.git",
					},
					Version: "v1.0.0",
				},
				Variables: datatypes.JSON(`{}`),
			},
			expectedGit: &deps.Git{
				Scheme: deps.GitSchemeSSH,
				Host:   "github.com",
				User:   "owner",
				Repo:   "repo",
				Subdir: "",
			},
			expectedLocal:   nil,
			expectedVersion: "v1.0.0",
		},
		{
			name: "Local source",
			component: &versource.Component{
				ModuleVersion: versource.ModuleVersion{
					Module: versource.Module{
						Source: "./local/module",
					},
					Version: "1.0.0",
				},
				Variables: datatypes.JSON(`{"env":"dev"}`),
			},
			expectedGit: nil,
			expectedLocal: &deps.Local{
				Directory: "./local/module",
			},
			expectedVersion: "",
		},
		{
			name: "Relative local source",
			component: &versource.Component{
				ModuleVersion: versource.ModuleVersion{
					Module: versource.Module{
						Source: "../parent/module",
					},
					Version: "2.0.0",
				},
				Variables: datatypes.JSON(`{}`),
			},
			expectedGit: nil,
			expectedLocal: &deps.Local{
				Directory: "../parent/module",
			},
			expectedVersion: "",
		},
		{
			name: "HTTPS git source with .git suffix",
			component: &versource.Component{
				ModuleVersion: versource.ModuleVersion{
					Module: versource.Module{
						Source: "https://github.com/terraform-aws-modules/terraform-aws-vpc.git",
					},
					Version: "v5.0.0",
				},
				Variables: datatypes.JSON(`{}`),
			},
			expectedGit: &deps.Git{
				Scheme: deps.GitSchemeHTTPS,
				Host:   "github.com",
				User:   "terraform-aws-modules",
				Repo:   "terraform-aws-vpc",
				Subdir: "",
			},
			expectedLocal:   nil,
			expectedVersion: "v5.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newJsonnetBundlerFromComponent(tt.component)

			assert.NotNil(t, result)
			assert.Equal(t, 1, result.Dependencies.Len())

			deps := result.Dependencies
			keys := deps.Keys()
			assert.Len(t, keys, 1)

			dependency, exists := deps.Get(keys[0])
			assert.True(t, exists)

			if tt.expectedGit != nil {
				assert.NotNil(t, dependency.Source.GitSource)
				assert.Nil(t, dependency.Source.LocalSource)
				assert.Equal(t, tt.expectedGit.Scheme, dependency.Source.GitSource.Scheme)
				assert.Equal(t, tt.expectedGit.Host, dependency.Source.GitSource.Host)
				assert.Equal(t, tt.expectedGit.User, dependency.Source.GitSource.User)
				assert.Equal(t, tt.expectedGit.Repo, dependency.Source.GitSource.Repo)
				assert.Equal(t, tt.expectedGit.Subdir, dependency.Source.GitSource.Subdir)
			}

			if tt.expectedLocal != nil {
				assert.NotNil(t, dependency.Source.LocalSource)
				assert.Nil(t, dependency.Source.GitSource)
				assert.Equal(t, tt.expectedLocal.Directory, dependency.Source.LocalSource.Directory)
			}

			assert.Equal(t, tt.expectedVersion, dependency.Version)
		})
	}
}

func TestParseGitSource(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected *deps.Git
	}{
		{
			name:   "HTTPS with subdir",
			source: "https://git.brndn.live/marcbran/raspi/tf/modules/github-repository",
			expected: &deps.Git{
				Scheme: deps.GitSchemeHTTPS,
				Host:   "git.brndn.live",
				User:   "marcbran",
				Repo:   "raspi",
				Subdir: "/tf/modules/github-repository",
			},
		},
		{
			name:   "SSH without subdir",
			source: "ssh://git@github.com/owner/repo.git",
			expected: &deps.Git{
				Scheme: deps.GitSchemeSSH,
				Host:   "github.com",
				User:   "owner",
				Repo:   "repo",
				Subdir: "",
			},
		},
		{
			name:   "HTTPS with .git suffix",
			source: "https://github.com/terraform-aws-modules/terraform-aws-vpc.git",
			expected: &deps.Git{
				Scheme: deps.GitSchemeHTTPS,
				Host:   "github.com",
				User:   "terraform-aws-modules",
				Repo:   "terraform-aws-vpc",
				Subdir: "",
			},
		},
		{
			name:   "HTTPS with complex subdir",
			source: "https://github.com/owner/repo/path/to/module/subdir",
			expected: &deps.Git{
				Scheme: deps.GitSchemeHTTPS,
				Host:   "github.com",
				User:   "owner",
				Repo:   "repo",
				Subdir: "/path/to/module/subdir",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGitSource(tt.source)

			assert.Equal(t, tt.expected.Scheme, result.Scheme)
			assert.Equal(t, tt.expected.Host, result.Host)
			assert.Equal(t, tt.expected.User, result.User)
			assert.Equal(t, tt.expected.Repo, result.Repo)
			assert.Equal(t, tt.expected.Subdir, result.Subdir)
		})
	}
}
