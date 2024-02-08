package ghapi

import (
	"fmt"

	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/models"
	api_results "github.com/mgmaster24/go-gh-scanner/models/api-results"
)

func (ghClient *GHClient) ScanPackageDeps(config *config.AppConfig) ([]api_results.RepoScanResult, error) {
	options := &github.ListOptions{
		PerPage: config.PerPage,
	}
	results := make([]api_results.RepoScanResult, 0)

	for _, dep := range config.Dependencies {
		config.CurrentDep = dep
		fmt.Println("Scanning for repos using", dep)
		scanResults, err := GetPagedResults[api_results.RepoScanResult](config, options, ghClient.SearchPackageFilesForDeps)
		if err != nil {
			return nil, err
		}
		results = append(results, scanResults...)
	}

	return results, nil
}

func (ghClient *GHClient) SearchPackageFilesForDeps(config *config.AppConfig, options *github.ListOptions) ([]api_results.RepoScanResult, *github.Response, error) {
	searchResult, resp, err := ghClient.Client.Search.Code(
		ghClient.Ctx,
		fmt.Sprintf("org:%s in:file filename:%s %s", config.Organization, config.PackageFile, config.GetShortDepName()),
		&github.SearchOptions{
			TextMatch:   true,
			ListOptions: *options,
		},
	)

	if err != nil {
		return nil, resp, err
	}

	results := make([]api_results.RepoScanResult, 0)
	for _, item := range searchResult.CodeResults {
		dependencyVersion := "0.0.0.0"
		shouldAdd := true
		for _, match := range item.TextMatches {
			var fragStr models.FragmentStr = models.FragmentStr(*match.Fragment)
			dependencyVersion, shouldAdd = fragStr.GetDepVersion(config.CurrentDep)
		}

		repoName := *item.Repository.Name
		if !config.ShouldIgnoreRepo(repoName) && shouldAdd {
			results = append(results, api_results.RepoScanResult{
				RepoName:          repoName,
				DependencyVersion: dependencyVersion,
			})
		}
	}

	return results, resp, nil
}
