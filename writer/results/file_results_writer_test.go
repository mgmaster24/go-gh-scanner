package results

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

func TestDepToFileName(t *testing.T) {
	cases := []struct {
		dep  string
		want string
	}{
		{"@m2s2/ng-lib", "m2s2-ng-lib.json"},
		{"@m2s2/react-lib", "m2s2-react-lib.json"},
		{"@m2s2/vue-lib", "m2s2-vue-lib.json"},
		{"my-lib", "my-lib.json"},
		{"repos", "repos.json"},
	}

	for _, c := range cases {
		got := depToFileName(c.dep)
		if got != c.want {
			t.Errorf("depToFileName(%q) = %q, want %q", c.dep, got, c.want)
		}
	}
}

func TestWriteRepoResults_GroupsByDependency(t *testing.T) {
	dest := t.TempDir()
	w := NewFileResultsWriter(dest)

	now := time.Now().Truncate(time.Second)
	results := api_results.GHRepoWriteResults{
		{Repo: "org/app-a", Sk: "DEP#@m2s2/ng-lib", Dependency: "@m2s2/ng-lib", Version: "2.0.0", ScmSite: "GitHub", LastModified: now},
		{Repo: "org/app-b", Sk: "DEP#@m2s2/ng-lib", Dependency: "@m2s2/ng-lib", Version: "1.5.0", ScmSite: "GitHub", LastModified: now},
		{Repo: "org/app-c", Sk: "DEP#@m2s2/react-lib", Dependency: "@m2s2/react-lib", Version: "3.0.0", ScmSite: "GitHub", LastModified: now},
	}

	if err := w.WriteRepoResults(results); err != nil {
		t.Fatalf("WriteRepoResults error: %v", err)
	}

	// Expect one file per dependency under the repos/ subdirectory.
	ngFile := filepath.Join(dest, "repos", "m2s2-ng-lib.json")
	reactFile := filepath.Join(dest, "repos", "m2s2-react-lib.json")

	for _, path := range []string{ngFile, reactFile} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected output file not found: %s", path)
		}
	}

	// Verify ng-lib file contains exactly 2 entries.
	raw, err := os.ReadFile(ngFile)
	if err != nil {
		t.Fatalf("failed to read ng-lib output: %v", err)
	}
	var ngResults api_results.GHRepoWriteResults
	if err := json.Unmarshal(raw, &ngResults); err != nil {
		t.Fatalf("failed to parse ng-lib output: %v", err)
	}
	if len(ngResults) != 2 {
		t.Errorf("expected 2 ng-lib results, got %d", len(ngResults))
	}

	// Verify react-lib file contains exactly 1 entry.
	raw, err = os.ReadFile(reactFile)
	if err != nil {
		t.Fatalf("failed to read react-lib output: %v", err)
	}
	var reactResults api_results.GHRepoWriteResults
	if err := json.Unmarshal(raw, &reactResults); err != nil {
		t.Fatalf("failed to parse react-lib output: %v", err)
	}
	if len(reactResults) != 1 {
		t.Errorf("expected 1 react-lib result, got %d", len(reactResults))
	}
}

func TestWriteRepoResults_EmptyResultsNoOp(t *testing.T) {
	dest := t.TempDir()
	w := NewFileResultsWriter(dest)

	if err := w.WriteRepoResults(nil); err != nil {
		t.Errorf("expected nil error for empty results, got %v", err)
	}

	entries, _ := os.ReadDir(dest)
	if len(entries) != 0 {
		t.Errorf("expected no files written for empty results, found %d", len(entries))
	}
}
