package builder

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed installer
var installerFs embed.FS

//go:embed os
var osFs embed.FS

//go:embed templates
var templateFs embed.FS

func copyFsToTempDir(fileSystem embed.FS, path string) (tempDir string, err error) {
	tempDir, err = os.MkdirTemp("", path)
	if err != nil {
		return
	}

	err = fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Construct the full path in the temporary directory
		targetPath := filepath.Join(tempDir, path)

		if d.IsDir() {
			// Create the directory in the temporary location
			return os.MkdirAll(targetPath, os.ModePerm)
		} else {
			// Open the embedded file
			srcFile, err := fileSystem.Open(path)
			if err != nil {
				return fmt.Errorf("error opening embedded file %s: %w", path, err)
			}
			defer func(srcFile fs.File) {
				e := srcFile.Close()
				if e != nil {
					err = e
				}
			}(srcFile)

			// Create the target file in the temporary directory
			dstFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("error creating target file %s: %w", targetPath, err)
			}
			defer func(dstFile *os.File) {
				e := dstFile.Close()
				if e != nil {
					err = e
				}
			}(dstFile)

			// Copy contents from embedded file to target file
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return fmt.Errorf("error copying file %s to %s: %w", path, targetPath, err)
			}
			return nil
		}
	})

	return
}

func copyInstallerFsToTempDir() (tempDir string, err error) {
	return copyFsToTempDir(installerFs, "installer")
}

func copyOsFsToTempDir() (tempDir string, err error) {
	return copyFsToTempDir(osFs, "os")
}
