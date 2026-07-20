// archive-utils.go - Provides utility methods for handling archives
package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// maxExtractedFileSize caps individual file extraction at 500 MB to guard
// against decompression bombs in archive entries.
const maxExtractedFileSize = 500 * 1024 * 1024

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

	// Resolve destination to an absolute path once so every entry can be
	// checked against it without repeated syscalls.
	destAbs, err := filepath.Abs(destination)
	if err != nil {
		return "", fmt.Errorf("error resolving destination path: %w", err)
	}
	destPrefix := destAbs + string(os.PathSeparator)

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

		target := filepath.Join(destAbs, hdr.Name)

		// Guard against Zip Slip: every extracted path must be inside destination.
		if !strings.HasPrefix(target+string(os.PathSeparator), destPrefix) {
			return "", fmt.Errorf("illegal file path in archive: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if root == "" {
				root = target
			}

			if err = CreateDir(target); err != nil {
				return "", err
			}
		case tar.TypeReg:
			f, err := os.Create(target)
			if err != nil {
				return "", err
			}

			_, copyErr := io.Copy(f, io.LimitReader(tr, maxExtractedFileSize))
			f.Close()
			if copyErr != nil {
				return "", copyErr
			}
		}
	}
	return root, nil
}
