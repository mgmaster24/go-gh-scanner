package utils

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// testArchivePath is the fixture used by archive extraction tests.
// TestMain creates it before any test runs and removes it after.
const testArchivePath = "testdata/TestDir.tar.gz"

func TestMain(m *testing.M) {
	if err := buildTestArchive(); err != nil {
		panic("failed to build test archive: " + err.Error())
	}
	code := m.Run()
	os.RemoveAll("testdata")
	os.Exit(code)
}

// buildTestArchive creates testdata/TestDir.tar.gz containing a small tree of
// .ts and .html files so the archive extraction tests have a reproducible fixture.
func buildTestArchive() error {
	if err := os.MkdirAll("testdata", 0755); err != nil {
		return err
	}

	f, err := os.Create(testArchivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	entries := []struct {
		name    string
		content string
		isDir   bool
	}{
		{"TestDir/", "", true},
		{"TestDir/component.ts", "export class AppComponent {}", false},
		{"TestDir/index.html", "<app-root></app-root>", false},
		{"TestDir/sub/", "", true},
		{"TestDir/sub/child.ts", "export class ChildComponent {}", false},
		{"TestDir/sub/child.html", "<div></div>", false},
		{"TestDir/ignored.json", `{"key":"value"}`, false},
	}

	for _, e := range entries {
		hdr := &tar.Header{Name: e.name}
		if e.isDir {
			hdr.Typeflag = tar.TypeDir
			hdr.Mode = 0755
		} else {
			hdr.Typeflag = tar.TypeReg
			hdr.Mode = 0644
			hdr.Size = int64(len(e.content))
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if !e.isDir {
			if _, err := tw.Write([]byte(e.content)); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestStringsMatch(t *testing.T) {
	cases := []struct {
		lhs, rhs string
		want     bool
	}{
		{"test-*", "test-lib", true},
		{"test-*-lib", "test-my-lib", true},
		{"*test-*", "my-test-lib", true},
		{"test*-*-lib", "testmy-extra-lib", true},
		{"no-match", "other", false},
	}
	for _, c := range cases {
		if got := StringsMatch(c.lhs, c.rhs); got != c.want {
			t.Errorf("StringsMatch(%q, %q) = %v, want %v", c.lhs, c.rhs, got, c.want)
		}
	}
}

func TestExtractGZFile(t *testing.T) {
	dest := t.TempDir()
	location, err := ExtractGZIP(testArchivePath, dest)
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(dest, "TestDir")
	if location != want {
		t.Fatalf("location = %q, want %q", location, want)
	}

	if err := RemoveDir(location); err != nil {
		t.Fatal(err)
	}
}

func TestExtractAndGetFilesByExts(t *testing.T) {
	dest := t.TempDir()
	location, err := ExtractGZIP(testArchivePath, dest)
	if err != nil {
		t.Fatal(err)
	}

	extensions := []string{".ts", ".html"}
	files, err := GetFilesByExtension(location, extensions)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Fatal("no files found")
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
			t.Errorf("unexpected extension on file %q", f)
		}
	}

	if err := RemoveDir(location); err != nil {
		t.Fatal(err)
	}
}

func TestExtractGZIP_ZipSlipRejected(t *testing.T) {
	dest := t.TempDir()

	// Build an archive with a path traversal entry.
	archivePath := filepath.Join(dest, "malicious.tar.gz")
	f, _ := os.Create(archivePath)
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "../../evil.txt",
		Size:     5,
		Mode:     0644,
	})
	tw.Write([]byte("pwned"))
	tw.Close()
	gw.Close()
	f.Close()

	extractDest := t.TempDir()
	_, err := ExtractGZIP(archivePath, extractDest)
	if err == nil {
		t.Fatal("expected error for path traversal entry, got nil")
	}
}
