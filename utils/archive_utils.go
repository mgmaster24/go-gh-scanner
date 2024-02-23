package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Extracts the gzip file represented by the file name and
// extracts the content to the provided destination
func ExtractGZIP(gzFileName string, destination string) (string, error) {
	root := ""
	gzFile, err := os.Open(gzFileName)
	if err != nil {
		return "", fmt.Errorf("error opening file. %e", err)
	}

	defer gzFile.Close()

	gzr, err := gzip.NewReader(gzFile)
	if err != nil {
		return "", fmt.Errorf("error reading gzip from file. %e", err)
	}

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}

		if hdr == nil {
			continue
		}

		target := filepath.Join(destination, hdr.Name)

		switch hdr.Typeflag {
		// directory
		case tar.TypeDir:
			// set root directory
			if root == "" {
				root = target
			}

			err = CreateDir(target)
			if err != nil {
				return "", err
			}
		// file
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return "", err
			}

			// make sure to close the file
			// we don't want to wait until the extraction operation completes
			f.Close()
		}
	}
	return root, nil
}
