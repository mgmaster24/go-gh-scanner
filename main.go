package main

import (
	"fmt"

	"github.com/mgmaster24/go-gh-scanner/config"
	ghapi "github.com/mgmaster24/go-gh-scanner/github-api"
	api_results "github.com/mgmaster24/go-gh-scanner/models/api-results"
	"github.com/mgmaster24/go-gh-scanner/writer"
)

func main() {
	configReader := config.JSONConfigReader{
		ScanConfigFile: "scan-config.json",
	}

	config, err := configReader.GetConfigValues()
	if err != nil {
		fmt.Println(err)
	}

	client := ghapi.NewClient(config.GHAuthToken)

	// Run scan to get results
	scanResults, err := client.ScanPackageDeps(config)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Saving results")
	//fmt.Println(scanResults)
	err = writer.SaveScanResults("scan-results.json", api_results.ScanResults{RepoScanResults: scanResults, Count: len(scanResults)})
	if err != nil {
		fmt.Println("Error saving", err)
	}

	repoResults := make([]api_results.GHRepoUsingDependencies, 0)
	for _, sr := range scanResults {
		repoData, err := client.GetRepoData(sr, config)
		if err != nil {
			fmt.Println(err)
			break
		}

		repoResults = append(repoResults, *repoData)
	}

	err = writer.SaveRepoResults("repo-results.json", api_results.GHRepoResults{Repos: repoResults, Count: len(repoResults)})
	if err != nil {
		fmt.Println("Error saving", err)
	}
	// Apply results to desired infrastructure

}
