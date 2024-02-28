package main

import (
	"fmt"

	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/github_api"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
	"github.com/mgmaster24/go-gh-scanner/search"
	"github.com/mgmaster24/go-gh-scanner/tokens"
	"github.com/mgmaster24/go-gh-scanner/utils"
	"github.com/mgmaster24/go-gh-scanner/writer/results"
)

func main() {
	appConfig, err := config.Read("app-config.json")
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
	ghRepoResults, err := scanResults.ToRepoData(appConfig, client.GetRepoData)
	if err != nil {
		panic(err)
	}

	fmt.Println("Saving repository results")
	err = ghRepoResults.SaveRepoResultsToFile("repo-results.json")
	if err != nil {
		panic(err)
	}

	err = utils.CreateDir(appConfig.ExtractDir)
	if err != nil {
		panic(err)
	}

	// get tokens - This is just and example of how to read tokens of different types for search
	tokenRetriever, err := tokens.CreateTokenReader(tokens.NGTokenReaderType)
	if err != nil {
		panic(err)
	}

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
		tokenRefs, err := search.FindTokenRefsInFiles(langFiles, tokenRetriever.ToTokens(), directory)
		if err != nil {
			panic(err)
		}

		fmt.Println("Saving token search results for repo", repo.Name)
		tokenResults := tokenRefs.ToTokenResults(repo.FullName, repo.Url, repo.DefaultBranch)
		// writing results - This is just and example of how to create an object as a
		// ResultsWriter and write the results where you would like
		resultsWriter, err := results.CreateResultsWriter(appConfig.WriterConfig)
		if err != nil {
			panic(err)
		}

		err = resultsWriter.Write(tokenResults)
		if err != nil {
			panic(err)
		}

		// Remove extracted directory
		err = utils.RemoveDir(directory)
		if err != nil {
			panic(err)
		}

		// Remove the archive file
		err = utils.RemoveFile(archiveFile)
		if err != nil {
			panic(err)
		}
	}

	// TODO
	// Seperate module, component, class results
	// Save results to Dynamo
}
