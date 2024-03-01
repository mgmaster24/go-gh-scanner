package api_results

import (
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/writer"
)

type ScanResults struct {
	RepoScanResults RepoScanResults `json:"results"`
	Count           int             `json:"count"`
}

type RepoScanResults []RepoScanResult

type RepoScanResult struct {
	RepoName          string `json:"repoName"`
	DependencyVersion string `json:"dependencyVersion"`
	Directory         string `json:"directory"`
}

type CodeScanResults struct {
	RepoName   string            `json:"repoName"`
	RepoURL    string            `json:"repoUrl"`
	NumMatches int               `json:"numMatches"`
	Tokens     []*TokenReference `json:"tokenRefs"`
}

func (scanResults *ScanResults) SaveScanResults(fileName string) error {
	return writer.MarshallAndSave(fileName, scanResults)
}

func SaveCodeScanResults(fileName string, results []*CodeScanResults) error {
	return writer.MarshallAndSave(fileName, results)
}

func (scanResults RepoScanResults) ToRepoData(
	owner string,
	teamsToIgnore config.TeamsToIgnore,
	getRepoData func(sr RepoScanResult, owner string, teamsToIgnore config.TeamsToIgnore) (*GHRepo, error)) (*GHRepoResults, error) {
	repoResults := make(GHRepos, 0)
	for _, sr := range scanResults {
		repoData, err := getRepoData(sr, owner, teamsToIgnore)
		if err != nil {
			return nil, err
		}
		repoResults = append(repoResults, *repoData)
	}

	return &GHRepoResults{Repos: repoResults, Count: len(repoResults)}, nil
}
