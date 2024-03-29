package api_results

import (
	"encoding/json"

	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/config"
)

type TeamsResponse struct {
	Teams []github.Team
}

// Implementing IO Writer in order to fill out TeamsResponse
func (resp *TeamsResponse) Write(data []byte) (n int, err error) {
	teams := make([]github.Team, 0)
	err = json.Unmarshal(data, &teams)
	if err != nil {
		return -1, nil
	}

	resp.Teams = teams
	return len(data), nil
}

func (teamsResponse *TeamsResponse) GetTeamsString(teamsToIgnore config.TeamsToIgnore) string {
	team := ""
	for _, tr := range teamsResponse.Teams {
		if !teamsToIgnore.ShouldIgnoreTeam(*tr.Slug) {
			if team == "" {
				team = *tr.Slug
			} else {
				team = team + "/" + *tr.Slug
			}
		}
	}

	return team
}
