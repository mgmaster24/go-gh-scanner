package github_api

import (
	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

// Get the organization's repositories for the provided config values
//
// Config values of interest
// - PerPage
// - Owner
func (ghClient *GHClient) GetReposForOrg(config *config.AppConfig) ([]api_results.GHRepo, error) {
	orgRepos, err := GetPagedResults(config, config.ToListOptions(), ghClient.getOrgRepoList)
	return orgRepos, err
}

func (ghClient *GHClient) getOrgRepoList(config *config.AppConfig, options *github.ListOptions) ([]api_results.GHRepo, *github.Response, error) {
	var orgRepos []api_results.GHRepo
	repos, resp, err := ghClient.Client.Repositories.ListByOrg(
		ghClient.Ctx,
		config.Owner,
		&github.RepositoryListByOrgOptions{ListOptions: *options})
	if err != nil {
		return nil, resp, err
	}
	for _, repo := range repos {
		description := ""
		if repo.Description != nil {
			description = *repo.Description
		}

		orgRepos = append(orgRepos, api_results.GHRepo{
			Name:        *repo.Name,
			Description: description,
			Url:         *repo.HTMLURL,
			Owner:       *repo.Owner.Login,
		})
	}

	return orgRepos, resp, nil
}

// Retrieves the repository data for the provided dependency scan results
func (ghClient *GHClient) GetRepoData(
	repoScanResults api_results.RepoScanResult,
	owner string,
	teamsToIgnore config.TeamsToIgnore) (*api_results.GHRepo, error) {
	repo, _, err := ghClient.Client.Repositories.Get(ghClient.Ctx, owner, repoScanResults.RepoName)
	if err != nil {
		return nil, err
	}

	teams, err := ghClient.GetTeams(*repo.TeamsURL, teamsToIgnore)
	if err != nil {
		return nil, err
	}

	description := ""
	if repo.Description != nil {
		description = *repo.Description
	}

	languages, err := ghClient.GetLanguages(*repo.LanguagesURL)
	if err != nil {
		return nil, err
	}

	ghRepo := &api_results.GHRepo{
		Name:              *repo.Name,
		FullName:          *repo.FullName,
		Description:       description,
		Owner:             *repo.Owner.Login,
		Url:               *repo.HTMLURL,
		Languages:         languages,
		LastModified:      repo.GetPushedAt().Time,
		Team:              teams,
		APIUrl:            *repo.URL,
		DefaultBranch:     *repo.DefaultBranch,
		DependencyVersion: repoScanResults.DependencyVersion,
	}

	return ghRepo, nil
}
