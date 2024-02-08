package api_results

type ScanResults struct {
	RepoScanResults []RepoScanResult `json:"results"`
	Count           int              `json:"count"`
}

type RepoScanResult struct {
	RepoName          string
	DependencyVersion string
}
