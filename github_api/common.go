package github_api

import (
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/config"
)

func GetPagedResults[T any](
	config *config.AppConfig,
	options *github.ListOptions,
	operation func(config *config.AppConfig, options *github.ListOptions) ([]T, *github.Response, error)) ([]T, error) {
	var results []T
	for {
		opResults, resp, err := operation(config, options)
		if err != nil {
			if WaitForRateLimit(err, resp) {
				continue
			}

			return nil, err
		}

		results = append(results, opResults...)
		if resp.NextPage == 0 {
			break
		}

		options.Page = resp.NextPage
	}

	return results, nil
}

func WaitForRateLimit(err error, resp *github.Response) bool {
	if _, ok := err.(*github.RateLimitError); ok {
		// Waiting for rate limit to refresh
		currentTime := time.Now()
		resetTime := resp.Rate.Reset.Time
		for currentTime.Before(resetTime) {
			sleepTime := resetTime.Sub(currentTime)
			fmt.Printf("Waiting for GitHub API Rate Limit reset.  Reset Time: %v\n", resetTime)
			time.Sleep(sleepTime)
			currentTime = time.Now()
		}

		return true
	}

	return false
}
