package utils

import (
	"path/filepath"
	"testing"
)

func TestStringsMatch(t *testing.T) {
	ok := StringsMatch("test-*", "test-lib")
	if !ok {
		t.Fatal("Fail wildcard match. Expected test-* to equal test-lib")
	}

	ok = StringsMatch("test-*-lib", "test-my-lib")
	if !ok {
		t.Fatal("Fail wildcard match. Expected test-* to equal test-lib")
	}

	ok = StringsMatch("*test-*", "my-test-lib")
	if !ok {
		t.Fatal("Fail wildcard match. Expected *test- to equal my-test-lib")
	}

	ok = StringsMatch("test*-*-lib", "testmy-extra-lib")
	if !ok {
		t.Fatal("Fail wildcard match. Expected *test- to equal my-test-lib")
	}
}

func TestExtractGZFile(t *testing.T) {
	location, err := ExtractGZIP("../temp/TestDir.tar.gz", "../temp")
	if err != nil {
		t.Fatal(err)
	}

	if location != "../temp/TestDir" {
		t.Fatal("Incorrect location")
	}

	err = RemoveDir(location)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExtractAndGetFilesByExts(t *testing.T) {
	location, err := ExtractGZIP("../temp/TestDir.tar.gz", "../temp")
	if err != nil {
		t.Fatal(err)
	}

	if location != "../temp/TestDir" {
		t.Fatal("Incorrect location")
	}

	extensions := []string{".ts", ".html"}
	files, err := GetFilesByExtension("../temp/TestDir", extensions)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) <= 0 {
		t.Fatal("No files were found")
	}

	for _, f := range files {
		ext := filepath.Ext(f)
		found := false
		for _, e := range extensions {
			if ext == e {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("Found file without required extension")
		}
	}

	err = RemoveDir(location)
	if err != nil {
		t.Fatal(err)
	}
}
