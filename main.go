package main

import (
	"fmt"

	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/github_api"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
	"github.com/mgmaster24/go-gh-scanner/search"
	"github.com/mgmaster24/go-gh-scanner/tokens"
	"github.com/mgmaster24/go-gh-scanner/utils"
	"github.com/mgmaster24/go-gh-scanner/writer"
)

func main() {
	appConfig := &config.AppConfig{}
	err := appConfig.Read("app-config.json")
	if err != nil {
		panic(err)
	}

	client := github_api.NewClient(appConfig.AuthToken)

	// Run scan to get results
	scanResults, err := client.ScanPackageDeps(appConfig)
	if err != nil {
		panic(err)
	}

	fmt.Println("Getting repository data for repos found during dependency scan.")
	repoResults := make([]api_results.GHRepo, 0)
	for _, sr := range scanResults {
		repoData, err := client.GetRepoData(sr, appConfig)
		if err != nil {
			fmt.Println(err)
			break
		}
		repoResults = append(repoResults, *repoData)
	}

	fmt.Println("Saving repository results")
	ghRepoResults := &api_results.GHRepoResults{Repos: repoResults, Count: len(repoResults)}
	err = ghRepoResults.SaveRepoResultsToFile("repo-results.json")
	if err != nil {
		panic(err)
	}

	err = utils.CreateDir(appConfig.ExtractDir)
	if err != nil {
		panic(err)
	}

	// get tokens - This is just and example of how to read tokens of different types for search
	var tokenRetriever tokens.TokenReader = &tokens.NgComponentReader{}
	err = tokenRetriever.Fetch("ng-tokens.json")
	if err != nil {
		panic(err)
	}

	for _, repo := range ghRepoResults.Repos {
		fmt.Println("Attempting to get repo archive for repo", repo.Name)
		archiveFile, err := repo.GetRepoArchive(appConfig.AuthToken, api_results.Tarball, appConfig.ExtractDir)
		if err != nil {
			panic(fmt.Sprintf("Error getting repository archive %s", err))
		}

		// extract gzip
		directory, err := utils.ExtractGZIP(archiveFile, appConfig.ExtractDir)
		if err != nil {
			panic(err)
		}

		// get files for language extensions
		langFiles, err := utils.GetFilesByExtension(directory, appConfig.GetLanguageExts())
		if err != nil {
			panic(err)
		}

		// search language specific files
		tokenRefs, err := search.FindTokenRefsInFiles(langFiles, tokenRetriever.ToTokens())
		if err != nil {
			panic(err)
		}

		// create directory for repo results
		dir := "results/" + repo.Name
		err = utils.CreateDir(dir)
		if err != nil {
			panic(err)
		}

		fmt.Println("Saving token search results for repo", repo.Name)
		count := 0
		for _, v := range tokenRefs {
			resultFile := fmt.Sprintf("%s/results_%v.json", dir, count)
			err = writer.MarshallAndSave(resultFile, v)
			if err != nil {
				panic(err)
			}

			count++
		}

		err = utils.RemoveDir(directory)
		if err != nil {
			panic(err)
		}

		err = utils.RemoveFile(archiveFile)
		if err != nil {
			panic(err)
		}
	}

	// TODO
	// Seperate module, component, class results
	// Save results to Dynamo
}
