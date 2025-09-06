package tfmodule

import (
	"net/url"
	"strings"

	"github.com/marcbran/versource/internal"
)

type ModuleIngester struct{}

func NewModuleIngester() *ModuleIngester {
	return &ModuleIngester{}
}

func (m *ModuleIngester) IngestModuleWithVersion(request internal.CreateModuleRequest) (*internal.Module, *internal.ModuleVersion, error) {
	return ingestModuleWithVersion(request)
}

func ingestModuleWithVersion(request internal.CreateModuleRequest) (*internal.Module, *internal.ModuleVersion, error) {
	if request.Source == "" {
		return nil, nil, internal.UserErr("source is required")
	}

	if strings.HasPrefix(request.Source, "./") || strings.HasPrefix(request.Source, "../") {
		if request.Version != "" {
			return nil, nil, internal.UserErr("local paths do not support version parameter")
		}
	} else if strings.HasPrefix(request.Source, "github.com/") || strings.HasPrefix(request.Source, "bitbucket.org/") || strings.HasPrefix(request.Source, "git::") || strings.HasPrefix(request.Source, "hg::") {
		if request.Version != "" {
			return nil, nil, internal.UserErr("git/mercurial sources do not support version parameter, use ref parameter in source string")
		}
	} else if strings.HasPrefix(request.Source, "s3::") {
		if request.Version != "" {
			return nil, nil, internal.UserErr("S3 sources do not support version parameter, use versionId parameter in source string")
		}
	} else if strings.HasPrefix(request.Source, "gcs::") {
		if request.Version != "" {
			return nil, nil, internal.UserErr("GCS sources do not support version parameter, use generation parameter in source string")
		}
	} else if !strings.Contains(request.Source, "::") && !strings.Contains(request.Source, "://") {
		if request.Version == "" {
			return nil, nil, internal.UserErr("terraform registry sources require version parameter")
		}
	}

	extractedVersion, err := extractVersionFromSource(request.Source)
	if err != nil {
		return nil, nil, err
	}

	cleanSource := request.Source
	if extractedVersion != "" {
		cleanSource = stripQueryParameters(request.Source)
	}

	module := &internal.Module{
		Source: cleanSource,
	}

	version := request.Version
	if extractedVersion != "" {
		version = extractedVersion
	}

	moduleVersion := &internal.ModuleVersion{
		Version: version,
	}

	return module, moduleVersion, nil
}

func stripQueryParameters(source string) string {
	if strings.HasPrefix(source, "s3::") {
		urlPart := strings.TrimPrefix(source, "s3::")
		u, err := url.Parse(urlPart)
		if err != nil {
			return source
		}
		u.RawQuery = ""
		return "s3::" + u.String()
	}

	if strings.HasPrefix(source, "gcs::") {
		urlPart := strings.TrimPrefix(source, "gcs::")
		u, err := url.Parse(urlPart)
		if err != nil {
			return source
		}
		u.RawQuery = ""
		return "gcs::" + u.String()
	}

	u, err := url.Parse(source)
	if err != nil {
		return source
	}
	u.RawQuery = ""
	return u.String()
}

func extractVersionFromSource(source string) (string, error) {
	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") {
		return "", nil
	}

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return "", internal.UserErr("HTTP/HTTPS sources are not supported")
	}

	if strings.HasPrefix(source, "s3::") {
		if !strings.Contains(source, "?versionId=") {
			return "", internal.UserErr("S3 sources require versionId parameter in source string")
		}
		u, err := url.Parse(strings.TrimPrefix(source, "s3::"))
		if err != nil {
			return "", internal.UserErr("invalid S3 source URL")
		}
		return u.Query().Get("versionId"), nil
	}

	if strings.HasPrefix(source, "gcs::") {
		if !strings.Contains(source, "?generation=") {
			return "", internal.UserErr("GCS sources require generation parameter in source string")
		}
		u, err := url.Parse(strings.TrimPrefix(source, "gcs::"))
		if err != nil {
			return "", internal.UserErr("invalid GCS source URL")
		}
		return u.Query().Get("generation"), nil
	}

	if strings.HasPrefix(source, "github.com/") || strings.HasPrefix(source, "bitbucket.org/") || strings.HasPrefix(source, "git::") || strings.HasPrefix(source, "hg::") {
		if !strings.Contains(source, "?ref=") {
			return "", internal.UserErr("git/mercurial sources require ref parameter in source string")
		}
		u, err := url.Parse(source)
		if err != nil {
			return "", internal.UserErr("invalid git/mercurial source URL")
		}
		return u.Query().Get("ref"), nil
	}

	if !strings.Contains(source, "::") && !strings.Contains(source, "://") {
		return "", nil
	}

	return "", internal.UserErr("unknown module source type")
}
