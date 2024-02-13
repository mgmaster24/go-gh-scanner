package ghapi

import (
	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/config"
	api_results "github.com/mgmaster24/go-gh-scanner/models/api-results"
)

func (ghClient *GHClient) GetReposForOrg(config *config.AppConfig) ([]api_results.GHRepo, error) {
	options := config.ToListOptions()

	orgRepos, err := GetPagedResults[api_results.GHRepo](config, options, ghClient.GetOrgRepoList)
	return orgRepos, err
}

func (ghClient *GHClient) GetOrgRepoList(config *config.AppConfig, options *github.ListOptions) ([]api_results.GHRepo, *github.Response, error) {
	var orgRepos []api_results.GHRepo
	repos, resp, err := ghClient.Client.Repositories.ListByOrg(
		ghClient.Ctx,
		config.Organization,
		&github.RepositoryListByOrgOptions{ListOptions: *options})
	if err != nil {
		return nil, resp, err
	}
	for _, repo := range repos {
		description := ""
		if repo.Description != nil {
			description = *repo.Description
		}

		language := ""
		if repo.Language != nil {
			language = *repo.Language
		}

		_, ok := config.GetLanguagesMap()[language]
		if ok {
			orgRepos = append(orgRepos, api_results.GHRepo{
				Name:        *repo.Name,
				Description: description,
				Url:         *repo.HTMLURL,
				Language:    language,
				Owner:       *repo.Owner.Login,
			})
		}
	}

	return orgRepos, resp, nil
}

func (ghClient *GHClient) GetRepoData(repoScanResults api_results.RepoScanResult, config *config.AppConfig) (*api_results.GHRepoUsingDependencies, error) {
	repo, _, err := ghClient.Client.Repositories.Get(ghClient.Ctx, config.Organization, repoScanResults.RepoName)
	if err != nil {
		return nil, err
	}

	teams, err := ghClient.GetTeams(*repo.TeamsURL, config)
	if err != nil {
		return nil, err
	}

	description := ""
	if repo.Description != nil {
		description = *repo.Description
	}

	language := ""
	if repo.Language != nil {
		language = *repo.Language
	}

	ghRepo := &api_results.GHRepoUsingDependencies{
		Repo: &api_results.GHRepo{
			Name:         *repo.Name,
			Description:  description,
			Language:     language,
			Owner:        *repo.Owner.Login,
			Url:          *repo.HTMLURL,
			LastModified: repo.GetPushedAt().Time,
			Team:         teams,
		},
		DependencyVersions: repoScanResults.DependencyVersion,
	}

	return ghRepo, nil
}
