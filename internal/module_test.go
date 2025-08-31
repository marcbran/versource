package internal

import (
	"testing"
)

func TestCreateModuleWithVersion(t *testing.T) {
	tests := []struct {
		name            string
		request         CreateModuleRequest
		expectedSource  string
		expectedVersion string
		wantErr         bool
		errMsg          string
	}{
		{
			name: "valid local module",
			request: CreateModuleRequest{
				Source:  "./local/modules/test-module",
				Version: "",
			},
			expectedSource:  "./local/modules/test-module",
			expectedVersion: "",
			wantErr:         false,
		},
		{
			name: "valid registry module",
			request: CreateModuleRequest{
				Source:  "hashicorp/consul/aws",
				Version: "0.1.0",
			},
			expectedSource:  "hashicorp/consul/aws",
			expectedVersion: "0.1.0",
			wantErr:         false,
		},
		{
			name: "valid github module",
			request: CreateModuleRequest{
				Source:  "github.com/hashicorp/example?ref=v1.2.0",
				Version: "",
			},
			expectedSource:  "github.com/hashicorp/example",
			expectedVersion: "v1.2.0",
			wantErr:         false,
		},
		{
			name: "valid git module",
			request: CreateModuleRequest{
				Source:  "git::https://example.com/network.git?ref=v1.2.0",
				Version: "",
			},
			expectedSource:  "git::https://example.com/network.git",
			expectedVersion: "v1.2.0",
			wantErr:         false,
		},
		{
			name: "valid bitbucket module",
			request: CreateModuleRequest{
				Source:  "bitbucket.org/hashicorp/terraform-consul-aws?ref=v1.0.0",
				Version: "",
			},
			expectedSource:  "bitbucket.org/hashicorp/terraform-consul-aws",
			expectedVersion: "v1.0.0",
			wantErr:         false,
		},
		{
			name: "valid mercurial module",
			request: CreateModuleRequest{
				Source:  "hg::http://example.com/vpc.hg?ref=v1.2.0",
				Version: "",
			},
			expectedSource:  "hg::http://example.com/vpc.hg",
			expectedVersion: "v1.2.0",
			wantErr:         false,
		},
		{
			name: "valid s3 module",
			request: CreateModuleRequest{
				Source:  "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip?versionId=abc123",
				Version: "",
			},
			expectedSource:  "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip",
			expectedVersion: "abc123",
			wantErr:         false,
		},
		{
			name: "valid gcs module",
			request: CreateModuleRequest{
				Source:  "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip?generation=123456789",
				Version: "",
			},
			expectedSource:  "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip",
			expectedVersion: "123456789",
			wantErr:         false,
		},
		{
			name: "empty source",
			request: CreateModuleRequest{
				Source:  "",
				Version: "1.0.0",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "source is required",
		},
		{
			name: "local module with version should fail",
			request: CreateModuleRequest{
				Source:  "./local/modules/test-module",
				Version: "1.0.0",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "local paths do not support version parameter",
		},
		{
			name: "registry module without version should fail",
			request: CreateModuleRequest{
				Source:  "hashicorp/consul/aws",
				Version: "",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "terraform registry sources require version parameter",
		},
		{
			name: "github module without ref should fail",
			request: CreateModuleRequest{
				Source:  "github.com/hashicorp/example",
				Version: "",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "git/mercurial sources require ref parameter in source string",
		},
		{
			name: "github module with version should fail",
			request: CreateModuleRequest{
				Source:  "github.com/hashicorp/example?ref=v1.2.0",
				Version: "1.0.0",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "git/mercurial sources do not support version parameter, use ref parameter in source string",
		},
		{
			name: "s3 module without versionId should fail",
			request: CreateModuleRequest{
				Source:  "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip",
				Version: "",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "S3 sources require versionId parameter in source string",
		},
		{
			name: "s3 module with version should fail",
			request: CreateModuleRequest{
				Source:  "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip?versionId=abc123",
				Version: "1.0.0",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "S3 sources do not support version parameter, use versionId parameter in source string",
		},
		{
			name: "gcs module without generation should fail",
			request: CreateModuleRequest{
				Source:  "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip",
				Version: "",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "GCS sources require generation parameter in source string",
		},
		{
			name: "gcs module with version should fail",
			request: CreateModuleRequest{
				Source:  "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip?generation=123456789",
				Version: "1.0.0",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "GCS sources do not support version parameter, use generation parameter in source string",
		},
		{
			name: "http module should fail",
			request: CreateModuleRequest{
				Source:  "https://example.com/vpc-module.zip",
				Version: "",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "HTTP/HTTPS sources are not supported",
		},
		{
			name: "unknown source type should fail",
			request: CreateModuleRequest{
				Source:  "unknown::source/type",
				Version: "",
			},
			expectedSource:  "",
			expectedVersion: "",
			wantErr:         true,
			errMsg:          "unknown module source type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, moduleVersion, err := createModuleWithVersion(tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("createModuleWithVersion() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("createModuleWithVersion() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("createModuleWithVersion() unexpected error = %v", err)
				return
			}

			if module == nil {
				t.Errorf("createModuleWithVersion() module is nil")
				return
			}

			if moduleVersion == nil {
				t.Errorf("createModuleWithVersion() moduleVersion is nil")
				return
			}

			if module.Source != tt.expectedSource {
				t.Errorf("createModuleWithVersion() module.Source = %v, want %v", module.Source, tt.expectedSource)
			}

			if moduleVersion.Version != tt.expectedVersion {
				t.Errorf("createModuleWithVersion() moduleVersion.Version = %v, want %v", moduleVersion.Version, tt.expectedVersion)
			}

			if moduleVersion.ModuleID != 0 {
				t.Errorf("createModuleWithVersion() moduleVersion.ModuleID should be 0, got %v", moduleVersion.ModuleID)
			}
		})
	}
}
