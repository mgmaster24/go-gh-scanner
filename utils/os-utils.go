// os-utils.go - Provide operating system IO related utility methods
//
// I.E.) opening files, creating directories, remove files/directories, getting file info.
package utils

import (
	"io/fs"
	"os"
	"path/filepath"
)

func CreateDir(path string) error {
	// check if it exists, if not, create it
	if _, err := os.Stat(path); err != nil {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

func CreateFile(fileName string) (*os.File, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func RemoveDir(dir string) error {
	return os.RemoveAll(dir)
}

func RemoveFile(filePath string) error {
	return os.Remove(filePath)
}

func OpenFile(filepath string) (*os.File, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func GetFilesByExtension(root string, extensions []string) ([]string, error) {
	files := make([]string, 0)
	filepath.WalkDir(root, func(file string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}

		// skip directories
		if d.IsDir() {
			return nil
		}

		for _, ext := range extensions {
			if filepath.Ext(d.Name()) == ext {
				files = append(files, file)
			}
		}

		return nil
	})

	return files, nil
}

func GetFileContents(file string) (string, error) {
	bytes, err := os.ReadFile(file)
	return string(bytes), err
}
