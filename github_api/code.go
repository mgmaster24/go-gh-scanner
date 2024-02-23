package github_api

import (
	"fmt"

	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/models"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

func (ghClient *GHClient) ScanPackageDeps(config *config.AppConfig) ([]api_results.RepoScanResult, error) {
	options := config.ToListOptions()
	results := make([]api_results.RepoScanResult, 0)

	for _, dep := range config.Dependencies {
		config.CurrentDep = dep
		fmt.Println("Scanning for repos using dependency", dep)
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

func (ghClient *GHClient) CodeSearch(repo api_results.GHRepo, tokens []string, config *config.AppConfig) (*api_results.CodeScanResults, error) {
	tokenRefs := make([]*api_results.TokenReference, 0)
	for _, token := range tokens {
		trefs, _, err := ghClient.Search(repo.Name, token, config.Owner, config.ToListOptions())
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

func (ghClient *GHClient) Search(repoName string, token string, org string, listOpts *github.ListOptions) ([]*api_results.TokenReference, *github.Response, error) {
	query := fmt.Sprintf("%s in:file org:%s repo:%s", token, org, repoName)
	fmt.Println("Executing query:", query)
	csrs, resp, err := ghClient.Client.Search.Code(ghClient.Ctx, query, &github.SearchOptions{
		TextMatch:   true,
		ListOptions: *listOpts,
	})

	if err != nil {
		if WaitForRateLimit(err, resp) {
			ghClient.Search(repoName, token, org, listOpts)
		}

		return nil, resp, err
	}

	return api_results.ToTokenRefs(csrs, token), resp, nil
}
