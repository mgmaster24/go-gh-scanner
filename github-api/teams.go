package ghapi

import (
	"net/http"

	"github.com/mgmaster24/go-gh-scanner/config"
	api_results "github.com/mgmaster24/go-gh-scanner/models/api-results"
)

func (ghClient *GHClient) GetTeams(teamsUrl string, config *config.AppConfig) (string, error) {
	req, err := ghClient.Client.NewRequest(http.MethodGet, teamsUrl, nil)
	if err != nil {
		return "", err
	}

	var teamsResponse *api_results.TeamsResponse = &api_results.TeamsResponse{}
	_, err = ghClient.Client.Do(ghClient.Ctx, req, teamsResponse)
	if err != nil {
		return "", err
	}

	return teamsResponse.GetTeamsString(config), nil
}
