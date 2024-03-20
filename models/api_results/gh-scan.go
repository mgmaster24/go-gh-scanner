package api_results

import (
	"fmt"
	"sync"

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

type AsyncRepoResults struct {
	mutex   sync.Mutex
	results GHRepos
}

func (scanResults *ScanResults) SaveScanResults(fileName string) error {
	return writer.MarshallAndSave(fileName, scanResults)
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

func (scanResults RepoScanResults) ToRepoDataAsync(
	owner string,
	teamsToIgnore config.TeamsToIgnore,
	getRepoData func(sr RepoScanResult, owner string, teamsToIgnore config.TeamsToIgnore) (*GHRepo, error)) (*GHRepoResults, error) {
	wg := sync.WaitGroup{}
	var asyncResults AsyncRepoResults
	for _, sr := range scanResults {
		wg.Add(1)
		go func(sr RepoScanResult, owner string, teamsToIgnore config.TeamsToIgnore) {
			defer wg.Done()
			repoData, err := getRepoData(sr, owner, teamsToIgnore)
			if err != nil {
				panic(err)
			}

			asyncResults.mutex.Lock()
			asyncResults.results = append(asyncResults.results, *repoData)
			asyncResults.mutex.Unlock()
		}(sr, owner, teamsToIgnore)
	}

	wg.Wait()

	return &GHRepoResults{Repos: asyncResults.results, Count: len(asyncResults.results)}, nil
}

func (scanResults RepoScanResults) RemoveDuplicates() RepoScanResults {
	allKeys := make(map[string]bool)
	list := RepoScanResults{}
	for _, item := range scanResults {
		if _, value := allKeys[item.RepoName]; !value {
			allKeys[item.RepoName] = true
			list = append(list, item)
		} else {
			fmt.Println("Removing", item)
		}
	}
	return list
}
