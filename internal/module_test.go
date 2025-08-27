package internal

import (
	"net/url"
	"strings"
	"testing"
)

func TestCreateModuleWithVersion(t *testing.T) {
	tests := []struct {
		name    string
		request CreateModuleRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid local module",
			request: CreateModuleRequest{
				Source:  "./local/modules/test-module",
				Version: "",
			},
			wantErr: false,
		},
		{
			name: "valid registry module",
			request: CreateModuleRequest{
				Source:  "hashicorp/consul/aws",
				Version: "0.1.0",
			},
			wantErr: false,
		},
		{
			name: "valid github module",
			request: CreateModuleRequest{
				Source:  "github.com/hashicorp/example?ref=v1.2.0",
				Version: "",
			},
			wantErr: false,
		},
		{
			name: "valid git module",
			request: CreateModuleRequest{
				Source:  "git::https://example.com/network.git?ref=v1.2.0",
				Version: "",
			},
			wantErr: false,
		},
		{
			name: "valid bitbucket module",
			request: CreateModuleRequest{
				Source:  "bitbucket.org/hashicorp/terraform-consul-aws?ref=v1.0.0",
				Version: "",
			},
			wantErr: false,
		},
		{
			name: "valid mercurial module",
			request: CreateModuleRequest{
				Source:  "hg::http://example.com/vpc.hg?ref=v1.2.0",
				Version: "",
			},
			wantErr: false,
		},
		{
			name: "valid s3 module",
			request: CreateModuleRequest{
				Source:  "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip?versionId=abc123",
				Version: "",
			},
			wantErr: false,
		},
		{
			name: "valid gcs module",
			request: CreateModuleRequest{
				Source:  "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip?generation=123456789",
				Version: "",
			},
			wantErr: false,
		},
		{
			name: "empty source",
			request: CreateModuleRequest{
				Source:  "",
				Version: "1.0.0",
			},
			wantErr: true,
			errMsg:  "source is required",
		},
		{
			name: "local module with version should fail",
			request: CreateModuleRequest{
				Source:  "./local/modules/test-module",
				Version: "1.0.0",
			},
			wantErr: true,
			errMsg:  "local paths do not support version parameter",
		},
		{
			name: "registry module without version should fail",
			request: CreateModuleRequest{
				Source:  "hashicorp/consul/aws",
				Version: "",
			},
			wantErr: true,
			errMsg:  "terraform registry sources require version parameter",
		},
		{
			name: "github module without ref should fail",
			request: CreateModuleRequest{
				Source:  "github.com/hashicorp/example",
				Version: "",
			},
			wantErr: true,
			errMsg:  "git/mercurial sources require ref parameter in source string",
		},
		{
			name: "github module with version should fail",
			request: CreateModuleRequest{
				Source:  "github.com/hashicorp/example?ref=v1.2.0",
				Version: "1.0.0",
			},
			wantErr: true,
			errMsg:  "git/mercurial sources do not support version parameter, use ref parameter in source string",
		},
		{
			name: "s3 module without versionId should fail",
			request: CreateModuleRequest{
				Source:  "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip",
				Version: "",
			},
			wantErr: true,
			errMsg:  "S3 sources require versionId parameter in source string",
		},
		{
			name: "s3 module with version should fail",
			request: CreateModuleRequest{
				Source:  "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip?versionId=abc123",
				Version: "1.0.0",
			},
			wantErr: true,
			errMsg:  "S3 sources do not support version parameter, use versionId parameter in source string",
		},
		{
			name: "gcs module without generation should fail",
			request: CreateModuleRequest{
				Source:  "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip",
				Version: "",
			},
			wantErr: true,
			errMsg:  "GCS sources require generation parameter in source string",
		},
		{
			name: "gcs module with version should fail",
			request: CreateModuleRequest{
				Source:  "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip?generation=123456789",
				Version: "1.0.0",
			},
			wantErr: true,
			errMsg:  "GCS sources do not support version parameter, use generation parameter in source string",
		},
		{
			name: "http module should fail",
			request: CreateModuleRequest{
				Source:  "https://example.com/vpc-module.zip",
				Version: "",
			},
			wantErr: true,
			errMsg:  "HTTP/HTTPS sources are not supported",
		},
		{
			name: "unknown source type should fail",
			request: CreateModuleRequest{
				Source:  "unknown::source/type",
				Version: "",
			},
			wantErr: true,
			errMsg:  "unknown module source type",
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

			if module.Source != tt.request.Source {
				t.Errorf("createModuleWithVersion() module.Source = %v, want %v", module.Source, tt.request.Source)
			}

			expectedVersion := tt.request.Version
			if strings.Contains(tt.request.Source, "?ref=") {
				u, _ := url.Parse(tt.request.Source)
				expectedVersion = u.Query().Get("ref")
			} else if strings.Contains(tt.request.Source, "?versionId=") {
				u, _ := url.Parse(strings.TrimPrefix(tt.request.Source, "s3::"))
				expectedVersion = u.Query().Get("versionId")
			} else if strings.Contains(tt.request.Source, "?generation=") {
				u, _ := url.Parse(strings.TrimPrefix(tt.request.Source, "gcs::"))
				expectedVersion = u.Query().Get("generation")
			}

			if moduleVersion.Version != expectedVersion {
				t.Errorf("createModuleWithVersion() moduleVersion.Version = %v, want %v", moduleVersion.Version, expectedVersion)
			}

			if moduleVersion.ModuleID != 0 {
				t.Errorf("createModuleWithVersion() moduleVersion.ModuleID should be 0, got %v", moduleVersion.ModuleID)
			}
		})
	}
}
