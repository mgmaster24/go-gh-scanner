package github_api

import (
	"fmt"

	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/models"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

// Initiates a code search that looks for the instance of the dependecies defined
//
// in the application configuration.  It then returns a slice of RepoScanResult
//
// containing the repository, location and version of the dependency that was searched.
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

// Private function that is passed to GetPagedResults that calls the GitHuB
// search API and waits if the rate limit for search has been exceeded.
func (ghClient *GHClient) searchPackageFilesForDeps(config *config.AppConfig, options *github.ListOptions) ([]api_results.RepoScanResult, *github.Response, error) {
	searchResult, resp, err := ghClient.Client.Search.Code(
		ghClient.Ctx,
		fmt.Sprintf("org:%s in:file filename:%s %s", config.Owner, config.PackageFile, config.GetShortDepName()),
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
			depDirectory := ""
			path := *item.Path
			depDirectory = path[:len(path)-len(config.PackageFile)]
			results = append(results, api_results.RepoScanResult{
				RepoName:          repoName,
				DependencyVersion: dependencyVersion,
				Directory:         depDirectory,
			})
		}
	}

	return results, resp, nil
}

// Calls the GitHub search API looking for the set of tokens provided.
func (ghClient *GHClient) CodeSearch(repo api_results.GHRepo, tokens []string, config *config.AppConfig) (*api_results.CodeScanResults, error) {
	tokenRefs := make([]*api_results.TokenReference, 0)
	for _, token := range tokens {
		trefs, _, err := ghClient.search(repo.Name, token, config.Owner, config.ToListOptions())
		if err != nil {
			return nil, err
		}
		tokenRefs = append(tokenRefs, trefs...)
	}

	return &api_results.CodeScanResults{
		NumMatches: len(tokenRefs),
		RepoName:   repo.Name,
		RepoURL:    repo.Url,
		Tokens:     tokenRefs,
	}, nil
}

// Executes the GitHub search api and will wait if a rate limit error is received.
func (ghClient *GHClient) search(repoName string, token string, org string, listOpts *github.ListOptions) ([]*api_results.TokenReference, *github.Response, error) {
	query := fmt.Sprintf("%s in:file org:%s repo:%s", token, org, repoName)
	var csrs *github.CodeSearchResult
	var err error
	var resp *github.Response
	for {
		csrs, resp, err = ghClient.Client.Search.Code(ghClient.Ctx, query, &github.SearchOptions{
			TextMatch:   true,
			ListOptions: *listOpts,
		})

		if err != nil {
			if WaitForRateLimit(err, resp) {
				continue
			}

			return nil, resp, err
		} else {
			break
		}
	}

	return api_results.ToTokenRefs(csrs, token), resp, nil
}
