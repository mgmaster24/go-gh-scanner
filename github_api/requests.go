package github_api

import (
	"net/http"

	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

// GitHub Teams Request
//
// There is not a provided method for retrieving the list of teams
// for a repository in the Go GitHub API
//
// Create and call the request using the provided teams URL
func (ghClient *GHClient) GetTeams(teamsUrl string, config *config.AppConfig) (string, error) {
	req, err := ghClient.createGetRequest(teamsUrl)
	if err != nil {
		return "", err
	}
	var teamsResponse *api_results.TeamsResponse = &api_results.TeamsResponse{}
	resp, err := ghClient.Client.Do(ghClient.Ctx, req, teamsResponse)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	return teamsResponse.GetTeamsString(config), nil
}

// GitHub Languages Request
//
// There is not a provided method for retrieving the list of language
// for a repository in the Go GitHub API.
//
// Create and call the request using the provided teams URL
func (ghClient *GHClient) GetLanguages(languagesUrl string) ([]api_results.GHLanguage, error) {
	req, err := ghClient.createGetRequest(languagesUrl)
	if err != nil {
		return []api_results.GHLanguage{}, err
	}

	var languages map[string]int
	resp, err := ghClient.Client.Do(ghClient.Ctx, req, &languages)
	if err != nil {
		return []api_results.GHLanguage{}, err
	}

	defer resp.Body.Close()

	values := api_results.ToGHLanguageSlice(languages)
	return values, nil
}

func (ghClient *GHClient) createGetRequest(url string) (*http.Request, error) {
	req, err := ghClient.Client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}
