package api_results

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/mgmaster24/go-gh-scanner/writer"
)

type ScanResults struct {
	RepoScanResults RepoScanResults `json:"results"`
	Count           int             `json:"count"`
}

type RepoScanResults []RepoScanResult

type RepoScanResult struct {
	RepoName          string `json:"repoName"`
	RepoOwner         string `json:"repoOwner"`
	Dependency        string `json:"dependency"`
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
	getRepoData func(sr RepoScanResult) (*GHRepo, error)) (*GHRepoResults, error) {
	repoResults := make(GHRepos, 0)
	for _, sr := range scanResults {
		repoData, err := getRepoData(sr)
		if err != nil {
			return nil, err
		}
		repoResults = append(repoResults, *repoData)
	}

	return &GHRepoResults{Repos: repoResults, Count: len(repoResults)}, nil
}

func (scanResults RepoScanResults) ToRepoDataAsync(
	getRepoData func(sr RepoScanResult) (*GHRepo, error)) (*GHRepoResults, error) {
	wg := sync.WaitGroup{}
	errCh := make(chan error, len(scanResults))
	var asyncResults AsyncRepoResults
	for _, sr := range scanResults {
		wg.Add(1)
		go func(sr RepoScanResult) {
			defer wg.Done()
			repoData, err := getRepoData(sr)
			if err != nil {
				errCh <- err
				return
			}

			asyncResults.mutex.Lock()
			asyncResults.results = append(asyncResults.results, *repoData)
			asyncResults.mutex.Unlock()
		}(sr)
	}

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return nil, err
	}

	return &GHRepoResults{Repos: asyncResults.results, Count: len(asyncResults.results)}, nil
}

// RemoveDuplicates deduplicates scan results by repo+dependency pair, keeping
// the entry with the highest declared version when a repo appears more than
// once (e.g. multiple package.json files in a monorepo).
func (scanResults RepoScanResults) RemoveDuplicates() RepoScanResults {
	// key: repoName + "|" + dependency
	best := make(map[string]RepoScanResult)
	for _, item := range scanResults {
		key := item.RepoName + "|" + item.Dependency
		if existing, found := best[key]; !found {
			best[key] = item
		} else if semverGT(item.DependencyVersion, existing.DependencyVersion) {
			fmt.Printf("Replacing %s version %s with higher version %s\n",
				item.RepoName, existing.DependencyVersion, item.DependencyVersion)
			best[key] = item
		}
	}

	list := make(RepoScanResults, 0, len(best))
	for _, item := range best {
		list = append(list, item)
	}
	return list
}

// semverGT reports whether version string a is greater than b.
// Compares major.minor.patch numerically; non-numeric parts fall back to
// string comparison so the function never panics on unexpected input.
func semverGT(a, b string) bool {
	partsA := strings.SplitN(a, ".", 3)
	partsB := strings.SplitN(b, ".", 3)
	for len(partsA) < 3 {
		partsA = append(partsA, "0")
	}
	for len(partsB) < 3 {
		partsB = append(partsB, "0")
	}
	for i := range 3 {
		na, errA := strconv.Atoi(partsA[i])
		nb, errB := strconv.Atoi(partsB[i])
		if errA != nil || errB != nil {
			// Non-numeric segment (pre-release tag, etc.) — don't guess an order.
			return false
		}
		if na != nb {
			return na > nb
		}
	}
	return false
}
