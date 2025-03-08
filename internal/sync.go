package internal

import (
	"context"
	"encoding/json"
	"github.com/google/go-jsonnet"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"os"
	"path"
	"strings"
)

// Sync TODO jb install
func Sync(ctx context.Context, include, exclude []string, configDir, dataDir string) error {
	vendorDir := path.Join(configDir, "vendor")
	mainFile := path.Join(configDir, "main.jsonnet")

	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{
		JPaths: []string{vendorDir},
	})
	files, err := vm.EvaluateFileMulti(mainFile)
	if err != nil {
		return err
	}
	for file, jsonContent := range files {
		target := path.Join(dataDir, file)

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
	err = os.Setenv("JSONNET_PATH", strings.Join([]string{vendorDir, configDir}, string(os.PathListSeparator)))
	if err != nil {
		return err
	}

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.8.0")),
	}
	execPath, err := installer.Install(ctx)
	if err != nil {
		return err
	}
	dbDir := path.Join(dataDir, "db")
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		return err
	}

	excludeSet := make(map[string]struct{})
	for _, e := range exclude {
		excludeSet[e] = struct{}{}
	}
	include = append([]string{"ddl"}, include...)
	var paths []string
	for _, i := range include {
		if _, ok := excludeSet[i]; ok {
			continue
		}
		paths = append(paths, i)
	}

	for _, p := range paths {
		syncDir := path.Join(dataDir, "sync", p)
		tf, err := tfexec.NewTerraform(syncDir, execPath)
		if err != nil {
			return err
		}
		tf.SetStdout(os.Stderr)
		err = tf.Init(ctx, tfexec.Upgrade(true))
		if err != nil {
			return err
		}
		err = tf.Apply(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
