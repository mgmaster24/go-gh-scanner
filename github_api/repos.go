package github_api

import (
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

// GetRepoData fetches metadata for a single repo identified by the scan result.
func (ghClient *GHClient) GetRepoData(sr api_results.RepoScanResult) (*api_results.GHRepo, error) {
	repo, _, err := ghClient.Client.Repositories.Get(ghClient.Ctx, sr.RepoOwner, sr.RepoName)
	if err != nil {
		return nil, err
	}

	langs, err := ghClient.GetLanguages(*repo.LanguagesURL)
	if err != nil {
		return nil, err
	}

	description := ""
	if repo.Description != nil {
		description = *repo.Description
	}

	return &api_results.GHRepo{
		Name:              *repo.Name,
		FullName:          *repo.FullName,
		Description:       description,
		Owner:             *repo.Owner.Login,
		Url:               *repo.HTMLURL,
		Languages:         langs,
		LastModified:      repo.GetPushedAt().Time,
		APIUrl:            *repo.URL,
		DefaultBranch:     *repo.DefaultBranch,
		Dependency:        sr.Dependency,
		DependencyVersion: sr.DependencyVersion,
		Directory:         sr.Directory,
	}, nil
}
