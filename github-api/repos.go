package ghapi

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/models"
	"golang.org/x/oauth2"
)

func GetReposForOrg(org string, token string, languages map[string]struct{}) ([]*models.GHRepo, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	options := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 50},
	}

	var orgRepos []*models.GHRepo
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, options)
		for _, repo := range repos {
			description := ""
			if repo.Description != nil {
				description = *repo.Description
			}

			language := ""
			if repo.Language != nil {
				language = *repo.Language
			}

			_, ok := languages[language]
			if ok {
				orgRepos = append(orgRepos, &models.GHRepo{
					Name:        *repo.Name,
					Description: description,
					ContentsUrl: *repo.ContentsURL,
					Language:    language,
				})
			}
		}

		if err != nil {
			return nil, err
		}

		if resp.NextPage == 0 {
			break
		}

		options.Page = resp.NextPage
	}

	return orgRepos, nil
}
