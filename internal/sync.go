package internal

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/go-jsonnet"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"os"
	"os/exec"
	"path"
	"strings"
)

type SyncOptions struct {
	Include []string
	Exclude []string

	ConfigDir string
	DataDir   string

	ForceDownload   bool
	DownloadVersion string
}

// Sync TODO Make sure to run jb install as well.
func Sync(ctx context.Context, options SyncOptions) error {
	vendorDir := path.Join(options.ConfigDir, "vendor")
	mainFile := path.Join(options.ConfigDir, "main.jsonnet")

	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{
		JPaths: []string{vendorDir},
	})
	files, err := vm.EvaluateFileMulti(mainFile)
	if err != nil {
		return err
	}
	for file, jsonContent := range files {
		target := path.Join(options.DataDir, file)

		targetDir := path.Dir(target)
		err := os.MkdirAll(targetDir, 0755)
		if err != nil {
			return err
		}

		var content string
		err = json.Unmarshal([]byte(jsonContent), &content)
		if err != nil {
			return err
		}

		targetFile, err := os.Create(target)
		if err != nil {
			return err
		}
		_, err = targetFile.WriteString(content)
		if err != nil {
			return err
		}
		_, err = targetFile.WriteString("\n")
		if err != nil {
			return err
		}
	}
	err = os.Setenv("JSONNET_PATH", strings.Join([]string{vendorDir, options.ConfigDir}, string(os.PathListSeparator)))
	if err != nil {
		return err
	}

	dbDir := path.Join(options.DataDir, "db")
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		return err
	}

	excludeSet := make(map[string]struct{})
	for _, e := range options.Exclude {
		excludeSet[e] = struct{}{}
	}
	options.Include = append([]string{"ddl"}, options.Include...)
	var paths []string
	for _, i := range options.Include {
		if _, ok := excludeSet[i]; ok {
			continue
		}
		paths = append(paths, i)
	}

	execPath, err := fetchTerraformPath(ctx, options)
	if err != nil {
		return err
	}
	for _, p := range paths {
		syncDir := path.Join(options.DataDir, "sync", p)
		tf, err := tfexec.NewTerraform(syncDir, execPath)
		if err != nil {
			return err
		}
		err = tf.Init(ctx, tfexec.Upgrade(true))
		if err != nil {
			return err
		}
		tf.SetStdout(os.Stderr)
		err = tf.Apply(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func fetchTerraformPath(ctx context.Context, options SyncOptions) (string, error) {
	if options.ForceDownload {
		return downloadTerraform(ctx, options)
	}
	execPath, err := exec.LookPath("terraform")
	if err != nil {
		if !errors.Is(err, exec.ErrNotFound) {
			return "", err
		}
		execPath, err = downloadTerraform(ctx, options)
		if err != nil {
			return "", err
		}
	}
	return execPath, nil
}

func downloadTerraform(ctx context.Context, options SyncOptions) (string, error) {
	tfDir := path.Join(options.DataDir, "tf")
	err := os.MkdirAll(tfDir, 0755)
	if err != nil {
		return "", err
	}
	installer := &releases.ExactVersion{
		Product:    product.Terraform,
		Version:    version.Must(version.NewVersion(options.DownloadVersion)),
		InstallDir: tfDir,
	}
	execPath, err := installer.Install(ctx)
	if err != nil {
		return "", err
	}
	return execPath, nil
}
