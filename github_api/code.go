package github_api

import (
	"fmt"

	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/models"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

// ScanPackageDeps searches all public GitHub repos for package.json files that
// reference any of the configured dependencies and returns the matching repos
// with their declared versions.
func (ghClient *GHClient) ScanPackageDeps(config *config.AppConfig) (api_results.RepoScanResults, error) {
	options := config.ToListOptions()
	results := make(api_results.RepoScanResults, 0)

	for _, dep := range config.Dependencies {
		config.CurrentDep = dep
		fmt.Println("Scanning for repos using dependency", dep)
		scanResults, err := GetPagedResults(config, options, ghClient.searchPackageFilesForDeps)
		if err != nil {
			return nil, err
		}
		results = append(results, scanResults...)
	}

	return results, nil
}

func (ghClient *GHClient) searchPackageFilesForDeps(config *config.AppConfig, options *github.ListOptions) ([]api_results.RepoScanResult, *github.Response, error) {
	orgFilter := ""
	if config.Owner != "" {
		orgFilter = fmt.Sprintf("org:%s ", config.Owner)
	}
	searchResult, resp, err := ghClient.Client.Search.Code(
		ghClient.Ctx,
		fmt.Sprintf("%sfilename:%s %s", orgFilter, config.PackageFile, config.CurrentDep),
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
			path := *item.Path
			results = append(results, api_results.RepoScanResult{
				RepoName:          repoName,
				RepoOwner:         *item.Repository.Owner.Login,
				Dependency:        config.CurrentDep,
				DependencyVersion: dependencyVersion,
				Directory:         path[:len(path)-len(config.PackageFile)],
			})
		}
	}

	return results, resp, nil
}
