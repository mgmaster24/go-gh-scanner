package api_results

import "github.com/mgmaster24/go-gh-scanner/writer"

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

func (scanResults *ScanResults) SaveScanResults(fileName string) error {
	return writer.MarshallAndSave(fileName, scanResults)
}

func SaveCodeScanResults(fileName string, results []*CodeScanResults) error {
	return writer.MarshallAndSave(fileName, results)
}
