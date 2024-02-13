package api_results

import (
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

type ScanResults struct {
	RepoScanResults []RepoScanResult `json:"results"`
	Count           int              `json:"count"`
}

type RepoScanResult struct {
	RepoName          string `json:"repoName"`
	DependencyVersion string `json:"dependencyVersion"`
}

type CodeScanResults struct {
	RepoName   string            `json:"repoName"`
	RepoURL    string            `json:"repoUrl"`
	NumMatches int               `json:"numMatches"`
	Tokens     []*TokenReference `json:"tokenRefs"`
}

type TokenReference struct {
	Token string
	Path  string `json:"path"`
	Link  string `json:"link"`
}

func ToTokenRefs(codeSearchResults *github.CodeSearchResult, token string) []*TokenReference {
	tokenRefs := make([]*TokenReference, 0)
	for _, cr := range codeSearchResults.CodeResults {
		repo := cr.Repository
		for _, tm := range cr.TextMatches {
			if *tm.Property == "content" && strings.Contains(*tm.Fragment, token) {
				tokenRefs = append(tokenRefs, &TokenReference{
					Link:  fmt.Sprintf("%s/blob/%s/%s", *repo.HTMLURL, *repo.DefaultBranch, *cr.Path),
					Path:  *cr.Path,
					Token: token,
				})
			}
		}
	}
	return tokenRefs
}
