package main

import (
	"fmt"

	"github.com/mgmaster24/go-gh-scanner/config"
	ghapi "github.com/mgmaster24/go-gh-scanner/github-api"
	api_results "github.com/mgmaster24/go-gh-scanner/models/api-results"
	"github.com/mgmaster24/go-gh-scanner/tokens"
	"github.com/mgmaster24/go-gh-scanner/writer"
)

func main() {
	appConfig := &config.AppConfig{Location: "app-config.json"}
	err := appConfig.GetConfig()
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
	for _, repo := range repoResults {
		codeScanResult, _, err := client.CodeSearch(*repo.Repo, tokens, appConfig)
		if err != nil {
			break
		}

		codeScanResults = append(codeScanResults, codeScanResult)
	}

	writer.SaveCodeScanResults("code-scan-results.json", codeScanResults)
}
