package internal

import (
	"context"
	"embed"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

//go:embed init
var initFs embed.FS

func Init(ctx context.Context, configDir string) error {
	return fs.WalkDir(initFs, "init", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		sourceFile, err := initFs.Open(p)
		if err != nil {
			return err
		}
		defer sourceFile.Close()
		target := path.Join(configDir, strings.TrimPrefix(p, "init"))
		_, err = os.Stat(target)
		if err == nil {
			return nil
		}
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		targetFile, err := os.Create(target)
		if err != nil {
			return err
		}
		defer targetFile.Close()
		_, err = io.Copy(targetFile, sourceFile)
		if err != nil {
			return err
		}
		return nil
	})
}
