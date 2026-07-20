package github_api

import (
	"net/http"

	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

func (ghClient *GHClient) GetLanguages(languagesUrl string) ([]api_results.GHLanguage, error) {
	var languages map[string]int
	if err := ghClient.doWithRateLimit(languagesUrl, &languages); err != nil {
		return nil, err
	}
	return api_results.ToGHLanguageSlice(languages), nil
}

// doWithRateLimit executes a GET request against url, deserializing the
// response into dest, and retries automatically if a rate limit is hit.
func (ghClient *GHClient) doWithRateLimit(url string, dest any) error {
	for {
		req, err := ghClient.Client.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		resp, err := ghClient.Client.Do(ghClient.Ctx, req, dest)
		if err != nil {
			if resp != nil && WaitForRateLimit(err, &github.Response{Response: resp.Response}) {
				continue
			}
			return err
		}

		resp.Body.Close()
		return nil
	}
}
