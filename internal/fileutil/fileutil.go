package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CopyDir(src, dst string) error {
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func PrintFileTree(dir string, prefix string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for i, entry := range entries {
		isLast := i == len(entries)-1
		currentPrefix := prefix
		if isLast {
			currentPrefix += "└── "
		} else {
			currentPrefix += "├── "
		}

		fmt.Printf("%s%s\n", currentPrefix, entry.Name())

		if entry.IsDir() {
			nextPrefix := prefix
			if isLast {
				nextPrefix += "    "
			} else {
				nextPrefix += "│   "
			}
			err = PrintFileTree(filepath.Join(dir, entry.Name()), nextPrefix)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
