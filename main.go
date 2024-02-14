package main

import (
	"fmt"
	"time"

	"github.com/mgmaster24/go-gh-scanner/config"
	ghapi "github.com/mgmaster24/go-gh-scanner/github-api"
	api_results "github.com/mgmaster24/go-gh-scanner/models/api-results"
	"github.com/mgmaster24/go-gh-scanner/tokens"
	"github.com/mgmaster24/go-gh-scanner/writer"
)

func main() {
	appConfig := &config.AppConfig{}
	err := appConfig.Read("app-config.json")
	if err != nil {
		panic(err)
	}

	client := ghapi.NewClient(appConfig.GHAuthToken)

	// Run scan to get results
	scanResults, err := client.ScanPackageDeps(appConfig)
	if err != nil {
		panic(err)
	}

	repoResults := make([]api_results.GHRepoUsingDependencies, 0)
	for _, sr := range scanResults {
		repoData, err := client.GetRepoData(sr, appConfig)
		if err != nil {
			fmt.Println(err)
			break
		}

		repoResults = append(repoResults, *repoData)
	}

	fmt.Println("Saving repository results")
	err = writer.SaveRepoResults("repo-results.json", &api_results.GHRepoResults{Repos: repoResults, Count: len(repoResults)})
	if err != nil {
		panic(err)
	}

	// get tokens - This is just and example of how to read tokens of different types for search
	var tokenRetriever tokens.TokenReader = &tokens.NgComponentReader{}
	err = tokenRetriever.Fetch("ng-tokens.json")
	if err != nil {
		panic(err)
	}

	tokens := tokenRetriever.ToTokens()
	codeScanResults := make([]*api_results.CodeScanResults, 0)
	numRepos := len(repoResults)
	fmt.Println("Starting code scan for tokens. Current Time:", time.Now(), "Total Repos:", numRepos)
	for _, repo := range repoResults {
		fmt.Println("Scanning for tokens in repo", repo.Repo.Name)
		codeScanResult, _, err := client.CodeSearch(*repo.Repo, tokens, appConfig)
		if err != nil {
			break
		}

		codeScanResults = append(codeScanResults, codeScanResult)
		fmt.Println("Finshed scanning", repo.Repo.Name)
		numRepos = numRepos - 1
		fmt.Println(numRepos, "to go!")
	}

	writer.SaveCodeScanResults("code-scan-results.json", codeScanResults)
}
