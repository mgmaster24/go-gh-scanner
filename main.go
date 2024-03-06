package main

import (
	"fmt"

	"github.com/mgmaster24/go-gh-scanner/aws_sdk"
	"github.com/mgmaster24/go-gh-scanner/cli"
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/github_api"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
	"github.com/mgmaster24/go-gh-scanner/search"
	"github.com/mgmaster24/go-gh-scanner/tokens"
	"github.com/mgmaster24/go-gh-scanner/utils"
	"github.com/mgmaster24/go-gh-scanner/writer/results"
)

func main() {
	flags := cli.InitFlags()
	appConfig, err := config.Read(flags.AppConfig)
	if err != nil {
		panic(err)
	}

	// Get secrets service client
	secretsService := aws_sdk.NewSecretsManagerClient()
	// Get the auth token value from the secret service
	authToken, err := secretsService.GetSecretString(appConfig.AuthTokenKey)
	if err != nil {
		panic(err)
	}

	// Create the github api client
	client := github_api.NewClient(authToken)

	// Run scan to get results
	scanResults, err := client.ScanPackageDeps(appConfig)
	if err != nil {
		panic(err)
	}

	fmt.Println("Getting repository data for repos found during dependency scan.")
	ghRepoResults, err := scanResults.ToRepoData(appConfig.Owner, appConfig.TeamsToIgnore, client.GetRepoData)
	if err != nil {
		panic(err)
	}

	fmt.Println("Saving repository results")
	writer, err := results.CreateResultsWriter(appConfig.RepoResultsConfig)
	if err != nil {
		panic(err)
	}

	err = writer.WriteRepoResults(ghRepoResults.Repos.ToWriteResults())
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

	err = tokenRetriever.Fetch(flags.TokensConfig)
	if err != nil {
		panic(err)
	}

	writer.UpdateDestination(appConfig.TokenResultsConfig.Destination)
	for _, repo := range ghRepoResults.Repos {
		fmt.Println("Attempting to get repo archive for repo", repo.Name)
		archiveFile, err := repo.GetRepoArchive(authToken, api_results.Tarball, appConfig.ExtractDir)
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
		err = writer.WriteTokenResults(tokenResults)
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

	fmt.Println("Scan operation complete.  Exiting...")
	// TODO
	// Seperate module, component, class results
}
