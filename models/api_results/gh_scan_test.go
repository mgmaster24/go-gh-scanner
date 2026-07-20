package api_results

import (
	"testing"
)

func TestSemverGT(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"2.0.0", "1.0.0", true},
		{"1.0.0", "2.0.0", false},
		{"1.0.0", "1.0.0", false},
		{"1.1.0", "1.0.9", true},
		{"1.0.1", "1.0.0", true},
		{"10.0.0", "9.0.0", true},
		{"1.0", "0.9", true},   // handles missing patch
		{"1", "0", true},       // handles major-only
		{"bad", "also-bad", false}, // non-numeric falls back gracefully
	}

	for _, c := range cases {
		got := semverGT(c.a, c.b)
		if got != c.want {
			t.Errorf("semverGT(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestRemoveDuplicates(t *testing.T) {
	t.Run("keeps unique repo+dependency pairs", func(t *testing.T) {
		input := RepoScanResults{
			{RepoName: "repo-a", Dependency: "@m2s2/ng-lib", DependencyVersion: "1.0.0"},
			{RepoName: "repo-a", Dependency: "@m2s2/react-lib", DependencyVersion: "2.0.0"},
			{RepoName: "repo-b", Dependency: "@m2s2/ng-lib", DependencyVersion: "1.0.0"},
		}
		got := input.RemoveDuplicates()
		if len(got) != 3 {
			t.Fatalf("expected 3 results (different dep pairs), got %d", len(got))
		}
	})

	t.Run("keeps higher version for same repo+dep", func(t *testing.T) {
		input := RepoScanResults{
			{RepoName: "repo-a", Dependency: "@m2s2/ng-lib", DependencyVersion: "1.0.0", Directory: "packages/web"},
			{RepoName: "repo-a", Dependency: "@m2s2/ng-lib", DependencyVersion: "2.0.0", Directory: "packages/mobile"},
		}
		got := input.RemoveDuplicates()
		if len(got) != 1 {
			t.Fatalf("expected 1 result after dedup, got %d", len(got))
		}
		if got[0].DependencyVersion != "2.0.0" {
			t.Errorf("expected higher version 2.0.0, got %s", got[0].DependencyVersion)
		}
	})

	t.Run("single entry passes through unchanged", func(t *testing.T) {
		input := RepoScanResults{
			{RepoName: "repo-a", Dependency: "@m2s2/ng-lib", DependencyVersion: "1.0.0"},
		}
		got := input.RemoveDuplicates()
		if len(got) != 1 {
			t.Fatalf("expected 1 result, got %d", len(got))
		}
	})
}
